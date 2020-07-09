package requestparsing

import "bytes"

const (
	statusOK                 = "HTTP/1.1 200 OK"
	statusTooManyHostHeaders = "HTTP/1.1 400 Bad Request: too many Host headers"
	statusNotImplemented     = "HTTP/1.1 501 Not Implemented"
	statusBadRequest         = "HTTP/1.1 400 Bad Request"
	statusInvalidHeaderName  = "HTTP/1.1 400 Bad Request: invalid header name"
	statusInvalidHeaderValue = "HTTP/1.1 400 Bad Request: invalid header value"
)

func extractStatus(response []byte) string {
	position := bytes.IndexByte(response, '\r')
	if position == -1 {
		return ""
	}
	return string(response[:position])
}
