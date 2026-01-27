package muxconn

import (
	"context"
	"net"

	"github.com/hashicorp/yamux"
	"golang.org/x/time/rate"
)

func NewYaMUX(parent context.Context, conn net.Conn, cfg *yamux.Config, serverSide bool) (Muxer, error) {
	if parent == nil {
		parent = context.Background()
	}

	var err error
	mux := &yamuxSession{
		traffic: new(trafficStat),
		limiter: newUnlimit(),
		streams: new(streamStat),
		parent:  parent,
	}
	if serverSide {
		mux.session, err = yamux.Server(conn, cfg)
	} else {
		mux.session, err = yamux.Client(conn, cfg)
	}
	if err != nil {
		return nil, err
	}

	return mux, nil
}

type yamuxSession struct {
	session *yamux.Session
	traffic *trafficStat
	limiter *rateLimiter
	streams *streamStat
	parent  context.Context
}

func (m *yamuxSession) Open(context.Context) (net.Conn, error) {
	return m.newConn(m.session.OpenStream())
}

func (m *yamuxSession) Accept() (net.Conn, error)  { return m.newConn(m.session.AcceptStream()) }
func (m *yamuxSession) Close() error               { return m.session.Close() }
func (m *yamuxSession) Addr() net.Addr             { return m.session.LocalAddr() }
func (m *yamuxSession) RemoteAddr() net.Addr       { return m.session.RemoteAddr() }
func (m *yamuxSession) IsClosed() bool             { return m.session.IsClosed() }
func (m *yamuxSession) Limit() rate.Limit          { return m.limiter.Limit() }
func (m *yamuxSession) SetLimit(bps rate.Limit)    { m.limiter.SetLimit(bps) }
func (m *yamuxSession) NumStreams() (int64, int64) { return m.streams.Load() }
func (m *yamuxSession) Traffic() (uint64, uint64)  { return m.traffic.Load() }
func (m *yamuxSession) Library() (string, string)  { return "yamux", "github.com/hashicorp/yamux" }

func (m *yamuxSession) newConn(stm *yamux.Stream, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}

	m.streams.openOne()
	ctx, cancel := context.WithCancelCause(m.parent)
	limit := m.limiter.warpReadWriter(ctx, stm)

	conn := &yamuxConn{
		parent: m,
		stream: stm,
		limit:  limit,
		cancel: cancel,
	}

	return conn, nil
}
