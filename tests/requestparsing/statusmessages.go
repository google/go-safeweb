package requestparsing

const (
	statusOK                 = "HTTP/1.1 200 OK\r\n"
	statusTooManyHostHeaders = "HTTP/1.1 400 Bad Request: too many Host headers\r\n"
	statusNotImplemented     = "HTTP/1.1 501 Not Implemented\r\n"
	statusBadRequest         = "HTTP/1.1 400 Bad Request\r\n"
)
