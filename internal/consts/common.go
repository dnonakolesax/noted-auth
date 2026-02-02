package consts

const EmptyString = ""

const (
	ATCookieKey  = "NTD-DNAnAT"
	RTCookieKey  = "NTD-DNART"
	IDTCookieKey = "NTD-DNALT"
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

const (
	ApplicationJSONContentType = "application/json"
)

const (
	CtxUserIDKey = "user_id"
)

const (
	IdentifierID = "ID"
	IdentifierLogin = "LOGIN"
)