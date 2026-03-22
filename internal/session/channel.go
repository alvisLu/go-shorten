package session

import (
	"sync"
	"time"
)

const InterimDebounce = 200 * time.Millisecond

type PendingFinal struct {
	ID   string
	Data [][]float32
}

type ChannelState struct {
	mu            sync.Mutex
	StreamBuffer  [][]float32
	InterimTimer  *time.Timer
	CurrentSegID  string
	PendingFinals []PendingFinal
	Processing    bool
}

func newChannelState() *ChannelState {
	return &ChannelState{}
}

func (c *ChannelState) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.InterimTimer != nil {
		c.InterimTimer.Stop()
		c.InterimTimer = nil
	}
	c.StreamBuffer = nil
	c.CurrentSegID = ""
	c.PendingFinals = nil
	c.Processing = false
}

func (c *ChannelState) Lock()   { c.mu.Lock() }
func (c *ChannelState) Unlock() { c.mu.Unlock() }
