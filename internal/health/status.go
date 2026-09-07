package health

import (
	"sync"
	"sync/atomic"
	"time"
)

type Status struct {
	mu             sync.RWMutex
	brokerUp       atomic.Bool
	lastPoll       atomic.Int64
	lastPollErr    atomic.Value
	pollCount      atomic.Uint64
	errorCount     atomic.Uint64
}

func New() *Status {
	s := &Status{}
	s.lastPollErr.Store("")
	return s
}

func (s *Status) SetBroker(up bool)   { s.brokerUp.Store(up) }
func (s *Status) RecordPoll(err error) {
	s.pollCount.Add(1)
	if err != nil {
		s.errorCount.Add(1)
		s.lastPollErr.Store(err.Error())
	} else {
		s.lastPollErr.Store("")
	}
	s.lastPoll.Store(time.Now().Unix())
}

func (s *Status) Healthy(maxAge time.Duration) bool {
	last := s.lastPoll.Load()
	if last == 0 {
		return false
	}
	return time.Since(time.Unix(last, 0)) < maxAge && s.brokerUp.Load()
}

func (s *Status) Snapshot() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]interface{}{
		"broker_up":    s.brokerUp.Load(),
		"last_poll_ts": s.lastPoll.Load(),
		"poll_count":   s.pollCount.Load(),
		"error_count":  s.errorCount.Load(),
		"last_error":   s.lastPollErr.Load().(string),
	}
}
