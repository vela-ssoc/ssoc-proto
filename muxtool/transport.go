package muxtool

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/vela-ssoc/ssoc-proto/muxproto"
)

type muxTransport struct {
	trp http.RoundTripper
	log *slog.Logger
}

func newTransport(dial muxproto.Dialer, log *slog.Logger) *muxTransport {
	trp := &http.Transport{
		DialContext:         dial.DialContext,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     time.Minute,
	}

	return &muxTransport{
		trp: trp,
		log: log,
	}
}

func (mt *muxTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	mt.log.Debug("HTTP 请求", "method", r.Method, "url", r.URL)
	return mt.trp.RoundTrip(r)
}
