package address

import (
	"common/appconstant"
	"fmt"
	"testing"

	fconstants "github.com/jabong/florest-core/src/common/constants"

	gk "github.com/onsi/ginkgo"
	gm "github.com/onsi/gomega"
)

func TestAddress(t *testing.T) {
	gm.RegisterFailHandler(gk.Fail)
	gk.RunSpecs(t, "Address Suite")
}

var userID string = "1773895"
var sessionID string = "12345678901234567890"

var _ = gk.Describe("Address API", func() {
	InitializeTestService()

	apiName := "AddressService"
	apiVersion := "v1"

	gk.Describe("GET /"+apiName+"/healthcheck", func() {
		request := CreateTestRequest("GET", "/"+apiName+"/healthcheck")
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return api health status", func() {
				MatchHeaderStatus(response)
				responseBody, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				validateHealthCheckResponse(response.Body.String())
			})
		})
	})

	gk.Describe("GET /"+apiName+"/"+apiVersion+"/address", func() {
		request := CreateTestRequest("GET", "/"+apiName+"/"+apiVersion+"/address")
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should successfully get", func() {
				MatchHeaderStatus(response)
				responseBody, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchVersionableNotFound(responseBody)
			})
		})
	})

	baseURL := fmt.Sprintf("/%s/%s/address/", apiName, apiVersion)

	//Test case for missing X-Jabong-UserId
	allURL := baseURL + appconstant.ALL
	gk.Describe("GET"+allURL, func() {
		request := CreateTestRequest("GET", allURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return user id missing in headers", func() {
				responseBody, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchHTTPCode(responseBody, fconstants.HTTPStatusBadRequestCode)
				gm.Expect(responseBody.Status.Errors[0].Code).To(gm.Equal(fconstants.APPErrorCode(1401)))
				gm.Expect(responseBody.Status.Errors[0].Message).To(gm.Equal("UserId must be provided in request header"))
			})
		})
	})

	// Test case for missing X-Jabong-SessionId
	gk.Describe("GET"+allURL, func() {
		request := CreateTestRequest("GET", allURL)
		request.Header.Add("X-Jabong-UserId", userID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return session id missing in headers", func() {
				responseBody, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchHTTPCode(responseBody, fconstants.HTTPStatusBadRequestCode)
				gm.Expect(responseBody.Status.Errors[0].Code).To(gm.Equal(fconstants.APPErrorCode(1401)))
				gm.Expect(responseBody.Status.Errors[0].Message).To(gm.Equal("SessionId must be provided in request header"))
			})
		})
	})

	// Test case for GET /v1/address/all
	gk.Describe("GET"+allURL, func() {
		request := CreateTestRequest("GET", allURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 3 addresses", func() {
				responseBody, addressResult := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(3))
			})
		})
	})
})
