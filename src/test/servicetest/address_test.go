package servicetest

import (
	"common/appconstant"
	"fmt"
	testUtils "test/utils"
	"testing"

	fconstants "github.com/jabong/florest-core/src/common/constants"

	gk "github.com/onsi/ginkgo"
	gm "github.com/onsi/gomega"
)

func TestAddress(t *testing.T) {
	gm.RegisterFailHandler(gk.Fail)
	gk.RunSpecs(t, "Address Suite")
}

var _ = gk.Describe("Address API", func() {
	InitializeTestService()

	apiName := "AddressService"
	apiVersion := "v1"
	baseURL := fmt.Sprintf("/%s/%s/address/", apiName, apiVersion)

	//Test case for missing X-Jabong-UserId
	allURL := baseURL + appconstant.ALL
	gk.Describe("GET"+allURL, func() {
		request := testUtils.CreateTestRequest("GET", allURL)
		request.Header.Add("X-Jabong-SessionId", "12345678901234567890")
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return user id missing in headers", func() {
				responseBody, _ := testUtils.GetHTTPResponseAndAddressResult(response.Body.String())
				testUtils.MatchHTTPCode(responseBody, fconstants.HTTPStatusBadRequestCode)
				gm.Expect(responseBody.Status.Errors[0].Code).To(gm.Equal(fconstants.APPErrorCode(1401)))
				gm.Expect(responseBody.Status.Errors[0].Message).To(gm.Equal("UserId must be provided in request header"))
			})
		})
	})

	// Test case for missing X-Jabong-SessionId
	gk.Describe("GET"+allURL, func() {
		request := testUtils.CreateTestRequest("GET", allURL)
		request.Header.Add("X-Jabong-UserId", "1")
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return session id missing in headers", func() {
				responseBody, _ := testUtils.GetHTTPResponseAndAddressResult(response.Body.String())
				testUtils.MatchHTTPCode(responseBody, fconstants.HTTPStatusBadRequestCode)
				gm.Expect(responseBody.Status.Errors[0].Code).To(gm.Equal(fconstants.APPErrorCode(1401)))
				gm.Expect(responseBody.Status.Errors[0].Message).To(gm.Equal("SessionId must be provided in request header"))
			})
		})
	})
})
