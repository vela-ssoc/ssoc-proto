package muxconn

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/xtaci/smux"
)

type smuxConn struct {
	parent *smuxSession
	stream *smux.Stream
	limit  io.ReadWriter
	closed atomic.Bool
	cancel context.CancelCauseFunc
}

func (c *smuxConn) Read(b []byte) (int, error) {
	n, err := c.limit.Read(b)
	c.parent.traffic.incrRX(n)

	return n, err
}

func (c *smuxConn) Write(b []byte) (int, error) {
	n, err := c.limit.Write(b)
	c.parent.traffic.incrTX(n)

	return n, err
}

func (c *smuxConn) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return net.ErrClosed
	}

	c.cancel(net.ErrClosed)
	err := c.stream.Close()
	c.parent.streams.closeOne()

	return err
}

func (c *smuxConn) LocalAddr() net.Addr                { return c.stream.LocalAddr() }
func (c *smuxConn) RemoteAddr() net.Addr               { return c.stream.RemoteAddr() }
func (c *smuxConn) SetDeadline(t time.Time) error      { return c.stream.SetDeadline(t) }
func (c *smuxConn) SetReadDeadline(t time.Time) error  { return c.stream.SetReadDeadline(t) }
func (c *smuxConn) SetWriteDeadline(t time.Time) error { return c.stream.SetWriteDeadline(t) }
