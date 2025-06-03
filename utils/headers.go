package utils

const (
	HeaderCacheControl = "Cache-Control"
	HeaderLocation     = "Location"
	HeaderContentType  = "Content-Type"
	HeaderIfMatch      = "If-Match"
)

const (
	ContentTypeJson = "application/json"
)

type CacheControl string

const (
	CacheControlNoStore CacheControl = "no-store"
)

func (c CacheControl) String() string {
	return string(c)
}
