package datasets

import (
	"net/http"
	"testing"

	"github.com/ONSdigital/dis-bundle-api/utils"
	. "github.com/smartystreets/goconvey/convey"
)

func createTestRequest(florenceToken, authToken *string) *http.Request {
	request := http.Request{
		Header: make(http.Header),
	}

	if florenceToken != nil {
		request.Header.Set(utils.HeaderFlorenceToken, *florenceToken)
	}
	if authToken != nil {
		request.Header.Set(utils.HeaderAuthorization, *authToken)
	}
	return &request
}

func TestCreateAuthHeaders(t *testing.T) {
	florenceToken := "florence-test-token"
	authToken := "auth-test-token"

	Convey("CreateAuthHeaders", t, func() {
		Convey("Should use florence token if existing", func() {
			request := createTestRequest(&florenceToken, &authToken)
			headers := CreateAuthHeaders(request)

			So(headers.ServiceToken, ShouldEqual, florenceToken)
		})

		Convey("Should use authorization token if florence token not set", func() {
			request := createTestRequest(nil, &authToken)

			headers := CreateAuthHeaders(request)

			So(headers.ServiceToken, ShouldEqual, authToken)
		})

		Convey("Should not set anything otherwise", func() {
			request := createTestRequest(nil, nil)

			headers := CreateAuthHeaders(request)

			So(headers.ServiceToken, ShouldEqual, "")
		})
	})
}
