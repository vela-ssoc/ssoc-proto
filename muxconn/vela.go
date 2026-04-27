package muxconn

import (
	"context"
	"net"

	"github.com/vela-ssoc/vela-common-mba/smux"
	"golang.org/x/time/rate"
)

// NewVela 实现旧版 SSOC tunnel 逻辑。
// 该通道实现属于自维护项目，由于缺乏维护人力且仅在 SSOC 项目中使用，
// 缺少全面的 BUG 反馈和使用案例，后续将逐步迁移并替换为开源 mux 组件实现。
//
// Deprecated: 推荐使用 NewSMUX 或 NewYaMUX 替代。
func NewVela(parent context.Context, conn net.Conn, cfg *smux.Config, serverSide bool) Muxer {
	if parent == nil {
		parent = context.Background()
	}

	mux := &velaSession{
		stats:   newMUXStreamStats(),
		limiter: newUnlimit(),
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
	stats   *muxStreamStats
	limiter *rateLimiter
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
func (m *velaSession) NumStreams() (int64, int64) { return m.stats.numStreams() }
func (m *velaSession) Traffic() (uint64, uint64)  { return m.stats.mux.load() }
func (m *velaSession) Streams() []Streamer        { return m.stats.actives() }

func (m *velaSession) Library() (string, string) {
	return "vela", "github.com/vela-ssoc/vela-common-mba/smux"
}

func (m *velaSession) newConn(stm *smux.Stream, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancelCause(m.parent)
	limit := m.limiter.warpReadWriter(ctx, stm)

	conn := &velaConn{
		parent: m,
		stream: stm,
		limit:  limit,
		cancel: cancel,
	}
	m.stats.putConn(conn)

	return conn, nil
}
