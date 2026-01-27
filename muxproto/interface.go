package muxproto

import (
	"net"

	"github.com/vela-ssoc/ssoc-proto/muxconn"
)

type ClientHooker interface {
	// Disconnected 通道掉线。
	Disconnected(mux muxconn.Muxer, err error)

	// Reconnected 通道掉线后重连成功。
	Reconnected(mux muxconn.Muxer)

	// OnExit 连接通道遇到不可重试的错误无法继续保持连接，
	// 通常原因是 context 取消。
	OnExit(err error)
}

func Outbound(laddr net.Addr) net.IP {
	if ta, ok := laddr.(*net.TCPAddr); ok {
		return ta.IP
	}

	return net.IPv4zero
}
