package muxconn

import (
	"context"
	"net"

	"github.com/vela-ssoc/vela-common-mba/smux"
	"golang.org/x/time/rate"
)

func NewVela(parent context.Context, conn net.Conn, cfg *smux.Config, serverSide bool) Muxer {
	if parent == nil {
		parent = context.Background()
	}

	mux := &velaSession{
		traffic: new(trafficStat),
		limiter: newUnlimit(),
		streams: new(streamStat),
		parent:  parent,
	}
	if serverSide {
		mux.session = smux.Server(conn, cfg)
	} else {
		mux.session = smux.Client(conn, cfg)
	}

	return mux
}

type velaSession struct {
	session *smux.Session
	traffic *trafficStat
	limiter *rateLimiter
	streams *streamStat
	parent  context.Context
}

func (m *velaSession) Open(context.Context) (net.Conn, error) {
	return m.newConn(m.session.OpenStream())
}

func (m *velaSession) Accept() (net.Conn, error)  { return m.newConn(m.session.AcceptStream()) }
func (m *velaSession) Close() error               { return m.session.Close() }
func (m *velaSession) Addr() net.Addr             { return m.session.LocalAddr() }
func (m *velaSession) RemoteAddr() net.Addr       { return m.session.RemoteAddr() }
func (m *velaSession) IsClosed() bool             { return m.session.IsClosed() }
func (m *velaSession) Limit() rate.Limit          { return m.limiter.Limit() }
func (m *velaSession) SetLimit(bps rate.Limit)    { m.limiter.SetLimit(bps) }
func (m *velaSession) NumStreams() (int64, int64) { return m.streams.Load() }
func (m *velaSession) Traffic() (uint64, uint64)  { return m.traffic.Load() }

func (m *velaSession) Library() (string, string) {
	return "vela", "github.com/vela-ssoc/vela-common-mba/smux"
}

func (m *velaSession) newConn(stm *smux.Stream, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}

	m.streams.openOne()
	ctx, cancel := context.WithCancelCause(m.parent)
	limit := m.limiter.warpReadWriter(ctx, stm)

	conn := &velaConn{
		parent: m,
		stream: stm,
		limit:  limit,
		cancel: cancel,
	}

	return conn, nil
}
