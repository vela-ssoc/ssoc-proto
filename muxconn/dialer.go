package muxconn

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type DialConfig struct {
	// Protocols 连接协议：
	// - smux
	// - yamux
	//
	// 不填写默认 smux 协议。
	Protocol string

	// Dialer websocket 拨号器。
	Dialer *websocket.Dialer

	// Path websocket 连接。
	Path string

	// PerTimeout 每次连接的超时时间。
	PerTimeout time.Duration

	Logger *slog.Logger
}

func (dc DialConfig) DialContext(parent context.Context, addresses []string) (Muxer, error) {
	addresses = dc.deduplicate(addresses)
	attrs := []any{"addresses", addresses}
	dc.log().Debug("准备连接服务端", attrs...)

	var errs []error
	for _, addr := range addresses {
		msgs := append(attrs, "address", addr)
		mux, err := dc.dialContext(parent, addr)
		if err == nil {
			dc.log().Info("连接服务成功", msgs...)
			return mux, nil
		}
		errs = append(errs, err)
		msgs = append(msgs, "error", err)
		dc.log().Warn("连接服务端出错", msgs...)
	}
	err := errors.Join(errs...)
	if err == nil {
		err = errors.New("连接地址不能为空")
	}
	attrs = append(attrs, "error", err)
	dc.log().Error("本轮全部连接失败", attrs...)

	return nil, err
}

func (dc DialConfig) dialContext(parent context.Context, address string) (Muxer, error) {
	proto := dc.Protocol
	if proto != "yamux" {
		proto = "smux"
	}
	quires := make(url.Values, 4)
	quires.Set("protocol", proto)

	reqURL := &url.URL{Scheme: "wss", Host: address, Path: dc.Path, RawQuery: quires.Encode()}
	if reqURL.Path == "" {
		reqURL.Path = "/api/v1/tunnel"
	}

	strURL := reqURL.String()
	d := dc.websocketDialer()
	ctx, cancel := dc.perContext(parent)
	defer cancel()

	ws, _, err := d.DialContext(ctx, strURL, nil)
	if err != nil {
		return nil, err
	}

	var mux Muxer
	conn := ws.NetConn()
	if proto == "yamux" {
		mux, err = NewYaMUX(parent, conn, nil, false)
	} else {
		mux, err = NewSMUX(parent, conn, nil, false)
	}
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return mux, nil
}

func (dc DialConfig) websocketDialer() *websocket.Dialer {
	if d := dc.Dialer; d != nil {
		return d
	}

	return &websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func (dc DialConfig) deduplicate(addresses []string) []string {
	uniq := make(map[string]struct{}, 8)
	rets := make([]string, 0, 10)
	for _, addr := range addresses {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}

		if _, exists := uniq[addr]; !exists {
			uniq[addr] = struct{}{}
			rets = append(rets, addr)
		}
	}

	return rets
}

func (dc DialConfig) perTimeout() time.Duration {
	if d := dc.PerTimeout; d > 0 {
		return d
	}

	return 5 * time.Second
}

func (dc DialConfig) perContext(ctx context.Context) (context.Context, context.CancelFunc) {
	d := dc.perTimeout()

	return context.WithTimeout(ctx, d)
}

func (dc DialConfig) log() *slog.Logger {
	if l := dc.Logger; l != nil {
		return l
	}

	return slog.Default()
}
