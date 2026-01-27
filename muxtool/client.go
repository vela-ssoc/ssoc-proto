package muxtool

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/ssoc-proto/muxproto"
)

type Client struct {
	dial muxproto.Dialer
	wsd  *websocket.Dialer
	cli  *http.Client
	log  *slog.Logger
}

// NewClient 创建一个客户端。
func NewClient(dial muxproto.Dialer, log *slog.Logger) Client {
	trp := newTransport(dial, log)
	cli := &http.Client{Transport: trp}

	wsd := &websocket.Dialer{
		NetDialContext:    dial.DialContext,
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: true,
	}

	return Client{
		dial: dial,
		wsd:  wsd,
		cli:  cli,
		log:  log,
	}
}

// HTTPClient 某些三方库需要，不建议业务开发者拿到该返回值后调用业务接口。
func (c Client) HTTPClient() *http.Client {
	return c.cli
}

// Transport 某些三方库需要，不建议业务开发者拿到该返回值后调用业务接口。
func (c Client) Transport() http.RoundTripper {
	return c.cli.Transport
}

// Do 某些三方库需要，不建议业务开发者拿到该返回值后调用业务接口。
func (c Client) Do(r *http.Request) (*http.Response, error) {
	return c.cli.Do(r)
}

// JSON 请求时无 body，但是响应数据是 JSON 报文。
// result 可以为 nil，表示不关心响应数据内容。
func (c Client) JSON(ctx context.Context, method, rawURL string, result any) error {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	res, err := c.send(req)
	if err != nil {
		return err
	}

	return c.unmarshalJSON(res.Body, result)
}

// SendJSON 请求携带 JSON body，响应数据是 JSON 报文。
// result 可以为 nil，表示不关心响应数据内容。
func (c Client) SendJSON(ctx context.Context, method, rawURL string, body, result any) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf8")

	res, err := c.send(req)
	if err != nil {
		return err
	}

	return c.unmarshalJSON(res.Body, result)
}

// Websocket 开启一个 websocket 双向流，如果走的是内部通道，则协议必须是 ws，不可以是 wss。
func (c Client) Websocket(ctx context.Context, rawURL string, header http.Header) (*websocket.Conn, error) {
	ws, _, err := c.wsd.DialContext(ctx, rawURL, header)

	return ws, err
}

//goland:noinspection GoUnhandledErrorResult
func (Client) unmarshalJSON(rc io.ReadCloser, result any) error {
	defer rc.Close()

	if result == nil || rc == http.NoBody {
		return nil
	}

	return json.NewDecoder(rc).Decode(result)
}

func (c Client) send(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", "ssoc-mux-client/1.0")
	res, err := c.cli.Do(r)
	if err != nil {
		return nil, err
	}

	code := res.StatusCode
	if c.is2xxStatusCode(code) || c.is3xxStatusCode(code) {
		return res, nil
	}

	body := res.Body
	//goland:noinspection GoUnhandledErrorResult
	defer body.Close()

	respErr := &ResponseError{Request: r}
	raw, err := io.ReadAll(io.LimitReader(body, 4096)) // 最多取 4K 响应报文，避免出现大的响应报文。
	if err != nil {
		return nil, respErr
	}
	respErr.RawBody = raw

	if c.isApplicationJSON(res.Header.Get("Content-Type")) {
		berr := new(BusinessErrorBody)
		if err = json.Unmarshal(raw, berr); err == nil {
			respErr.BusinessError = berr
		}
	}

	return nil, respErr
}

func (Client) is2xxStatusCode(code int) bool { return code/100 == 2 }
func (Client) is3xxStatusCode(code int) bool { return code/100 == 3 }

func (Client) isApplicationJSON(contentType string) bool {
	before, _, _ := strings.Cut(contentType, ";")
	before = strings.ToLower(strings.TrimSpace(before))

	return before == "application/json"
}
