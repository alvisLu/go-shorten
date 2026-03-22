package session

import (
	"sync"
)

type Session struct {
	mu            sync.RWMutex
	send          chan<- any
	channels      map[string]*ChannelState
	isRunning     bool
	sourceLang    string
	targetLang    string
	sampleRate    int
	enableDenoise bool
}

func NewSession(send chan<- any) *Session {
	return &Session{
		send: send,
		channels: map[string]*ChannelState{
			"mic":      newChannelState(),
			"loopback": newChannelState(),
		},
	}
}

func (s *Session) Send(v any) {
	s.send <- v
}

type WsResp struct {
	Status string `json:"status"`
}

type WsErrorResp struct {
	Error string `json:"error"`
}

func (s *Session) Health() {
	s.Send(WsResp{Status: "ok"})
}

func (s *Session) Start(sourceLang, targetLang string, sampleRate int, enableDenoise bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sourceLang = sourceLang
	s.targetLang = targetLang
	s.sampleRate = sampleRate
	s.enableDenoise = enableDenoise
	s.isRunning = true
	s.Send(WsResp{Status: "started"})
}

func (s *Session) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isRunning = false
	for _, ch := range s.channels {
		ch.reset()
	}
	s.Send(WsResp{Status: "stopped"})
}

func (s *Session) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

func (s *Session) Channel(name string) *ChannelState {
	return s.channels[name]
}

func (s *Session) SourceLang() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sourceLang
}

func (s *Session) SampleRate() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sampleRate
}

func (s *Session) IsEnableDenoise() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enableDenoise
}
