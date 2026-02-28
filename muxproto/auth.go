package muxproto

import (
	"encoding/json"
	"io"
)

// authEOF 认证报文结束标记符，因为 JSON 字符串不能包含这种字符。
const authEOF = 0x00

// WriteAuth 写入认证报文。
func WriteAuth(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return err
	}
	_, err := w.Write([]byte{authEOF})

	return err
}

// ReadAuth 读取认证报文
func ReadAuth(r io.Reader, result any) error {
	dr := &delimReader{raw: r, dem: authEOF}
	return json.NewDecoder(dr).Decode(result)
}

type delimReader struct {
	raw io.Reader // 原始流
	dem byte      // 结束标志
	err error     // 错误信息
}

func (dr *delimReader) Read(p []byte) (int, error) {
	if dr.err != nil {
		return 0, dr.err
	}

	n, err := dr.raw.Read(p)
	for i := 0; i < n; i++ {
		if p[i] == dr.dem {
			dr.err = io.EOF // 标记下次读取返回 EOF
			return i, nil   // 返回分隔符之前的数据
		}
	}

	return n, err
}

type BrokerAuthRequest struct {
	Secret     string   `json:"secret"   validate:"required"` // broker 密钥
	Semver     string   `json:"semver"   validate:"required"` // broker 版本号
	Inet       string   `json:"inet"     validate:"required"`
	Goos       string   `json:"goos"     validate:"required"`
	Goarch     string   `json:"goarch"   validate:"required"`
	PID        int      `json:"pid"`
	Hostname   string   `json:"hostname"`
	Workdir    string   `json:"workdir"`
	Executable string   `json:"executable"`
	Args       []string `json:"args"`
}

type BrokerAuthResponse struct {
	Code   int               `json:"code"`
	Text   string            `json:"text"`
	Config *BrokerBootConfig `json:"config"`
}

func (r BrokerAuthResponse) Err() error {
	if r.Code/100 == 2 {
		return nil
	}

	return &AuthError{Code: r.Code, Text: r.Text}
}

type BrokerBootConfig struct {
	URI string `json:"uri" validate:"mongodb_connection_string"`
}

type AgentBootConfig struct{}

type AuthError struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func (a *AuthError) Error() string {
	return a.Text
}
