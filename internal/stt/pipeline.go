package stt

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/alvisLu/go-shorten/internal/session"
	"github.com/alvisLu/go-shorten/internal/stt/whisper"
)

// Translator is the interface for translation providers (Phase 6+).
type Translator interface {
	Translate(ctx context.Context, text, sourceLang, targetLang string) (string, error)
}

// TranscriptMsg is sent over the WebSocket back to the client.
type TranscriptMsg struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	ID      string `json:"id"`
	Text    string `json:"text"`
	Final   bool   `json:"final"`
}

// Pipeline handles the STT flow: interim debounce and final serialisation.
type Pipeline struct {
	w          *whisper.Whisper
	translator Translator // nil in Phase 5
}

// NewPipeline creates a Pipeline. w may be nil (STT disabled). t may be nil (no translation).
func NewPipeline(w *whisper.Whisper, t Translator) *Pipeline {
	return &Pipeline{w: w, translator: t}
}

// OnInterimFrame appends pcm to the channel buffer and resets the 200ms debounce timer.
func (p *Pipeline) OnInterimFrame(sess *session.Session, chName, id string, pcm []float32) {
	ch := sess.Channel(chName)
	ch.Lock()
	ch.StreamBuffer = append(ch.StreamBuffer, pcm)
	ch.CurrentSegID = id
	// take snapshot for the closure (copy slice headers; float32 data owned by pcm arg)
	snapshot := make([][]float32, len(ch.StreamBuffer))
	copy(snapshot, ch.StreamBuffer)
	if ch.InterimTimer != nil {
		ch.InterimTimer.Stop()
	}
	capturedEpoch := ch.Epoch()
	ch.InterimTimer = time.AfterFunc(session.InterimDebounce, func() {
		if !sess.IsRunning() {
			return
		}
		ch.Lock()
		if ch.Epoch() != capturedEpoch || ch.Processing || len(ch.StreamBuffer) == 0 {
			ch.Unlock()
			return
		}

		currentId := ch.CurrentSegID
		ch.Processing = true
		ch.Unlock()
		p.transcribeInterim(sess, chName, currentId, flattenPCM(snapshot))
		ch.Lock()

		ch.Processing = false
		ch.Unlock()
		p.runPendingFinal(sess, chName)
	})
	ch.Unlock()
}

// OnFinalFrame stops the interim timer, drains the buffer, and runs (or queues) transcribeSegment.
func (p *Pipeline) OnFinalFrame(sess *session.Session, chName, id string, pcm []float32) {
	ch := sess.Channel(chName)
	ch.Lock()
	if ch.InterimTimer != nil {
		ch.InterimTimer.Stop()
		ch.InterimTimer = nil
	}
	ch.StreamBuffer = nil
	ch.CurrentSegID = ""
	pf := session.PendingFinal{ID: id, Data: pcm}
	if ch.Processing {
		ch.PendingFinals = append(ch.PendingFinals, pf)
		ch.Unlock()
		return
	}
	ch.Processing = true
	ch.Unlock()
	go p.transcribeSegment(sess, chName, pf)
}

func (p *Pipeline) transcribeInterim(sess *session.Session, chName, id string, pcm []float32) {
	log.Printf("transcribeInterim: ch=%s id=%s", chName, id)
	if p.w == nil || rms(pcm) < 0.01 {
		return
	}
	rate := sess.SampleRate()
	if rate == 0 {
		rate = 48000
	}
	resampled := whisper.Resample(pcm, rate)
	text, err := p.w.Transcribe(resampled, sess.SourceLang())
	if err != nil || text == "" {
		return
	}
	sess.Send(TranscriptMsg{Type: "transcript", Channel: chName, ID: id, Text: text, Final: false})
}

func (p *Pipeline) transcribeSegment(sess *session.Session, chName string, pf session.PendingFinal) {
	log.Printf("transcribeSegment: ch=%s id=%s", chName, pf.ID)
	defer p.runPendingFinal(sess, chName)
	if p.w == nil || rms(pf.Data) < 0.01 {
		return
	}
	rate := sess.SampleRate()
	if rate == 0 {
		rate = 48000
	}
	resampled := whisper.Resample(pf.Data, rate)
	if sess.IsEnableDenoise() {
		frame, err := audioFrame(chName, pf.ID, resampled)
		if err != nil {
			log.Printf("transcribeSegment error: %v", err)
			return
		}
		sess.Send(frame)
	}
	text, err := p.w.Transcribe(resampled, sess.SourceLang())
	if err != nil || text == "" {
		return
	}
	sess.Send(TranscriptMsg{Type: "transcript", Channel: chName, ID: pf.ID, Text: text, Final: true})
}

// audioFrame encodes resampled 16kHz PCM as a binary WebSocket frame.
// Frame layout: [0xDA][channel][id: 21 bytes][Float32LE PCM bytes]
// id must be at most 21 bytes; longer values return an error.
func audioFrame(chName, id string, pcm []float32) ([]byte, error) {
	if len(id) > 21 {
		return nil, fmt.Errorf("audioFrame id exceeds 21 bytes")
	}
	buf := make([]byte, 23+len(pcm)*4)
	buf[0] = 0xDA
	if chName == "loopback" {
		buf[1] = 1
	}
	copy(buf[2:23], id)
	for i, v := range pcm {
		binary.LittleEndian.PutUint32(buf[23+i*4:], math.Float32bits(v))
	}
	return buf, nil
}

func (p *Pipeline) runPendingFinal(sess *session.Session, chName string) {
	ch := sess.Channel(chName)
	ch.Lock()
	if len(ch.PendingFinals) == 0 {
		ch.Processing = false
		ch.Unlock()
		return
	}
	next := ch.PendingFinals[0]
	ch.PendingFinals = ch.PendingFinals[1:]
	ch.Unlock()
	go p.transcribeSegment(sess, chName, next)
}

func rms(pcm []float32) float64 {
	if len(pcm) == 0 {
		return 0
	}
	var sum float64
	for _, v := range pcm {
		sum += float64(v) * float64(v)
	}
	return math.Sqrt(sum / float64(len(pcm)))
}

func flattenPCM(chunks [][]float32) []float32 {
	total := 0
	for _, c := range chunks {
		total += len(c)
	}
	out := make([]float32, 0, total)
	for _, c := range chunks {
		out = append(out, c...)
	}
	return out
}
