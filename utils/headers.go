package utils

// Header key constants
const (
	HeaderAuthorization = "Authorization"
	HeaderCacheControl  = "Cache-Control"
	HeaderLocation      = "Location"
	HeaderContentType   = "Content-Type"
	HeaderETag          = "ETag"
	HeaderFlorenceToken = "X-Florence-Token"
	HeaderIfMatch       = "If-Match"
)

// CacheControl represents a value for the Cache-Control header
type CacheControl string

const (
	CacheControlNoStore CacheControl = "no-store"
)

func (c CacheControl) String() string {
	return string(c)
}
