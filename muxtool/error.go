package muxtool

import (
	"net/http"
	"strconv"
)

// ProblemDetails is Problem Details for HTTP APIs.
// RFC7807 https://www.rfc-editor.org/rfc/rfc7807
type ProblemDetails struct {
	Host     string `json:"host"     xml:"host"`
	Method   string `json:"method"   xml:"method"`
	Instance string `json:"instance" xml:"instance"`
	Status   int    `json:"status"   xml:"status"`
	Title    string `json:"title"    xml:"title"`
	Detail   string `json:"detail"   xml:"detail"`
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
