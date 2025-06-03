package utils

const (
	HeaderAuthorization = "Authorization"
	HeaderCacheControl  = "Cache-Control"
	HeaderLocation      = "Location"
	HeaderContentType   = "Content-Type"
	HeaderETag          = "ETag"
	HeaderFlorenceToken = "X-Florence-Token"
	HeaderIfMatch       = "If-Match"
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
