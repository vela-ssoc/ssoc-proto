package muxconn

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/vela-ssoc/vela-common-mba/smux"
)

type velaConn struct {
	parent *velaSession
	stream *smux.Stream
	limit  io.ReadWriter
	closed atomic.Bool
	cancel context.CancelCauseFunc
}

func (c *velaConn) Read(b []byte) (int, error) {
	n, err := c.limit.Read(b)
	c.parent.traffic.incrRX(n)

	return n, err
}

func (c *velaConn) Write(b []byte) (int, error) {
	n, err := c.limit.Write(b)
	c.parent.traffic.incrTX(n)

	return n, err
}

func (c *velaConn) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return net.ErrClosed
	}

	c.cancel(net.ErrClosed)
	err := c.stream.Close()
	c.parent.streams.closeOne()

	return err
}

func (c *velaConn) LocalAddr() net.Addr                { return c.stream.LocalAddr() }
func (c *velaConn) RemoteAddr() net.Addr               { return c.stream.RemoteAddr() }
func (c *velaConn) SetDeadline(t time.Time) error      { return c.stream.SetDeadline(t) }
func (c *velaConn) SetReadDeadline(t time.Time) error  { return c.stream.SetReadDeadline(t) }
func (c *velaConn) SetWriteDeadline(t time.Time) error { return c.stream.SetWriteDeadline(t) }
