package muxtool

import (
	"encoding/xml"
	"net/http"
	"strconv"
)

// ProblemDetails [RFC9457] for HTTP APIs.
//
// [RFC9457]: https://datatracker.ietf.org/doc/html/rfc9457
type ProblemDetails struct {
	// XMLName 指定根标签名
	XMLName xml.Name `json:"-" xml:"problem"`

	// Status https://datatracker.ietf.org/doc/html/rfc9457#section-3.1.2
	//
	// 状态码
	Status int `json:"status" xml:"status"`

	// Title https://datatracker.ietf.org/doc/html/rfc9457#section-3.1.3
	//
	// 问题类型的简短、人类可读的摘要。
	Title string `json:"title,omitzero" xml:"title,omitzero"`

	// Detail https://datatracker.ietf.org/doc/html/rfc9457#section-3.1.4
	//
	// 问题的详细原因，应当专注于帮助客户端解决问题，而不是提供调试信息。
	Detail string `json:"detail,omitzero" xml:"detail,omitzero"`

	// Instance https://datatracker.ietf.org/doc/html/rfc9457#section-3.1.5
	//
	// 它可以是请求的 URI，也可以是它作为问题实例的唯一标识符，可能对服务器具有重要意义。
	Instance string `json:"instance,omitzero" xml:"instance,omitzero"`

	// Method 请求方法。
	//
	// 自定义扩展数据。
	Method string `json:"method,omitzero" xml:"method,omitzero"`

	// Host 请求主机头。
	//
	// 自定义扩展数据。
	Host string `json:"host,omitzero" xml:"host,omitzero"`
}

func (d ProblemDetails) String() string {
	return "problem details, host='" + d.Host + "'" +
		", method='" + d.Method + "'" +
		", instance='" + d.Instance + "'" +
		", status=" + strconv.Itoa(d.Status) +
		", title='" + d.Title + "'" +
		", detail='" + d.Detail + "'"
}

type ResponseError struct {
	Request *http.Request
	RawBody []byte
	Details *ProblemDetails
}

func (r *ResponseError) Error() string {
	if r.Details != nil {
		return r.Details.String()
	}

	req, res := r.Request, r.Request.Response

	return "response problem details, host='" + r.Request.Host + "'" +
		", method='" + req.Method + "'" +
		", instance='" + req.RequestURI + "'" +
		", status=" + strconv.Itoa(res.StatusCode) +
		", body='" + string(r.RawBody) + "'"
}
