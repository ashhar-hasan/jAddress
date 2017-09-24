package utils

import (
	"net/http"
	"net/http/httptest"

	fconstants "github.com/jabong/florest-core/src/common/constants"
	utilhttp "github.com/jabong/florest-core/src/common/utils/http"
	gm "github.com/onsi/gomega"
)

//MatchHeaderStatus verified the Content-Type and Cache-Control headers
func MatchHeaderStatus(responseRecorder *httptest.ResponseRecorder) {
	gm.Expect(responseRecorder.HeaderMap.Get("Content-Type")).To(gm.Equal("application/json"))
	gm.Expect(responseRecorder.HeaderMap.Get("Cache-Control")).To(gm.Equal(""))
}

//MatchSuccessResponseStatus verifies status for successful response
func MatchSuccessResponseStatus(responseBody *utilhttp.Response) {
	gm.Expect(responseBody.Status.HTTPStatusCode).To(gm.Equal(fconstants.HTTPCode(200)))
	gm.Expect(responseBody.Status.Success).To(gm.Equal(true))
}

//MatchVersionableNotFound verifies status for versionable not found
func MatchVersionableNotFound(responseBody *utilhttp.Response) {
	gm.Expect(responseBody.Status.HTTPStatusCode).To(gm.Equal(fconstants.HTTPCode(http.StatusNotFound)))
	gm.Expect(responseBody.Status.Errors[0].Code).To(gm.Equal(fconstants.APPErrorCode(1601)))
	gm.Expect(responseBody.Status.Errors[0].Message).To(gm.Equal("Versionable not found in version manager"))
	//TODO
	//gm.Expect(responeBody.DebugData).To(gm.Equal(""))
}

//MatchNotFoundResponseStatus verifies status for Not Found response
func MatchNotFoundResponseStatus(responseBody *utilhttp.Response) {
	gm.Expect(responseBody.Status.HTTPStatusCode).To(gm.Equal(fconstants.HTTPCode(404)))
	gm.Expect(responseBody.Status.Success).To(gm.Equal(false))
}

//MatchHTTPCode verifies status for given HTTPCode
func MatchHTTPCode(responseBody *utilhttp.Response, httpCode fconstants.HTTPCode) {
	gm.Expect(responseBody.Status.HTTPStatusCode).To(gm.Equal(fconstants.HTTPStatusBadRequestCode))
	gm.Expect(responseBody.Status.Success).To(gm.Equal(false))
}
