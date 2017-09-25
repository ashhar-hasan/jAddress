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

var userID = "1773895"
var sessionID = "12345678901234567890"

var _ = gk.Describe("Address API", func() {
	InitializeTestService()

	apiName := "AddressService"
	apiVersion := "v1"

	// Test case for healthcheck
	gk.Describe("GET /"+apiName+"/healthcheck", func() {
		request := CreateTestRequest("GET", "/"+apiName+"/healthcheck")
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return api health status", func() {
				MatchHeaderStatus(response)
				responseBody, _, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				validateHealthCheckResponse(response.Body.String())
			})
		})
	})

	// Test case for versionable not found
	gk.Describe("GET /"+apiName+"/"+apiVersion+"/address", func() {
		request := CreateTestRequest("GET", "/"+apiName+"/"+apiVersion+"/address")
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should successfully get", func() {
				MatchHeaderStatus(response)
				responseBody, _, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchVersionableNotFound(responseBody)
			})
		})
	})

	baseURL := fmt.Sprintf("/%s/%s/address/", apiName, apiVersion)

	// Test case for missing X-Jabong-UserId
	allURL := baseURL + appconstant.ALL
	gk.Describe("GET"+allURL, func() {
		request := CreateTestRequest("GET", allURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return user id missing in headers", func() {
				responseBody, _, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
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
				responseBody, _, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchHTTPCode(responseBody, fconstants.HTTPStatusBadRequestCode)
				gm.Expect(responseBody.Status.Errors[0].Code).To(gm.Equal(fconstants.APPErrorCode(1401)))
				gm.Expect(responseBody.Status.Errors[0].Message).To(gm.Equal("SessionId must be provided in request header"))
			})
		})
	})
})

var _ = gk.Describe("GET Address API", func() {
	InitializeTestService()

	apiName := "AddressService"
	apiVersion := "v1"
	baseURL := fmt.Sprintf("/%s/%s/address/", apiName, apiVersion)
	allURL := baseURL + appconstant.ALL

	// Test case for GET /v1/address/all
	gk.Describe("GET"+allURL, func() {
		request := CreateTestRequest("GET", allURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 3 addresses", func() {
				responseBody, addressResult, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(3))
			})
		})
	})

	// Test case for GET /v1/address/shipping
	shippingURL := baseURL + appconstant.SHIPPING
	gk.Describe("GET"+shippingURL, func() {
		request := CreateTestRequest("GET", shippingURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 1 default shipping addresses", func() {
				responseBody, addressResult, address, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(1))
				gm.Expect(address.IsDefaultBilling).To(gm.Equal("0"))
				gm.Expect(address.IsDefaultShipping).To(gm.Equal("1"))
				gm.Expect(address.Id).To(gm.Equal("35495058"))
			})
		})
	})

	// Test case for GET /v1/address/billing
	billingURL := baseURL + appconstant.BILLING
	gk.Describe("GET"+billingURL, func() {
		request := CreateTestRequest("GET", billingURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 1 default billing addresses", func() {
				responseBody, addressResult, address, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(1))
				gm.Expect(address.IsDefaultBilling).To(gm.Equal("1"))
				gm.Expect(address.IsDefaultShipping).To(gm.Equal("0"))
				gm.Expect(address.Id).To(gm.Equal("35495082"))
			})
		})
	})

	// Test case for GET /v1/address/other
	otherURL := baseURL + appconstant.OTHER
	gk.Describe("GET"+otherURL, func() {
		request := CreateTestRequest("GET", otherURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 1 other addresses", func() {
				responseBody, addressResult, _, addressList := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(1))
				gm.Expect(addressList["1268408"].IsDefaultBilling).To(gm.Equal("0"))
				gm.Expect(addressList["1268408"].IsDefaultShipping).To(gm.Equal("0"))
				gm.Expect(addressList["1268408"].Id).To(gm.Equal("1268408"))
			})
		})
	})

	// Test case for GET /v1/address/all user not found
	gk.Describe("GET"+allURL, func() {
		request := CreateTestRequest("GET", allURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", "1")
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 0 addresses", func() {
				responseBody, addressResult, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(0))
			})
		})
	})

	// Test case for GET /v1/address/shipping user not found
	gk.Describe("GET"+shippingURL, func() {
		request := CreateTestRequest("GET", shippingURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", "1")
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 0 addresses", func() {
				responseBody, addressResult, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(0))
			})
		})
	})

	// Test case for GET /v1/address/billing user not found
	gk.Describe("GET"+billingURL, func() {
		request := CreateTestRequest("GET", billingURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", "1")
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 0 addresses", func() {
				responseBody, addressResult, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(0))
			})
		})
	})

	// Test case for GET /v1/address/other user not found
	gk.Describe("GET"+otherURL, func() {
		request := CreateTestRequest("GET", otherURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", "1")
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 0 addresses", func() {
				responseBody, addressResult, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(0))
			})
		})
	})
})
