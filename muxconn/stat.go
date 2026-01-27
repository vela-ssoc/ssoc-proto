package muxconn

import "sync/atomic"

type trafficStat struct {
	rx, tx atomic.Uint64
}

func (s *trafficStat) Load() (rx, tx uint64) {
	return s.rx.Load(), s.tx.Load()
}

func (s *trafficStat) incrRX(n int) {
	if n > 0 {
		s.rx.Add(uint64(n))
	}
}

func (s *trafficStat) incrTX(n int) {
	if n > 0 {
		s.tx.Add(uint64(n))
	}
}

type streamStat struct {
	cumulative, active atomic.Int64
}

func (s *streamStat) Load() (cumulative, active int64) {
	return s.cumulative.Load(), s.active.Load()
}

func (s *streamStat) openOne() {
	s.cumulative.Add(1)
	s.active.Add(1)
}

func (s *streamStat) closeOne() {
	s.active.Add(-1)
}
