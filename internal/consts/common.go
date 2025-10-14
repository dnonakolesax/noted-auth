package consts

const EmptyString = ""

const (
	ATCookieKey = "NTD-DNAnAT"
	RTCookieKey = "NTD-DNART"
)

type ContextKey string

const (
	TraceContextKey ContextKey = "trace"
	TraceLoggerKey  string     = "trace-id"
)

const (
	ErrorLoggerKey = "error"
)

const (
	HTTPHeaderXRequestID = "X-Request-Id"
)
