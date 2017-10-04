package address

import (
	"common/appconstant"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	fconstants "github.com/jabong/florest-core/src/common/constants"
	gk "github.com/onsi/ginkgo"
	gm "github.com/onsi/gomega"
)

func TestAddress(t *testing.T) {
	gm.RegisterFailHandler(gk.Fail)
	gk.RunSpecs(t, "Address Suite")
}

const (
	userID              = "1773895"
	sessionID           = "12345678901234567890"
	invalidUserID       = "1"
	invalidSessionID    = "abcd"
	updateAddressID     = "35495082"
	oldDefaultAddressID = "35495058"
)

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
				gm.Expect(responseBody.Status.Errors[0].Code).To(gm.Equal(fconstants.ParamsInSufficientErrorCode))
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
				gm.Expect(responseBody.Status.Errors[0].Code).To(gm.Equal(fconstants.ParamsInSufficientErrorCode))
				gm.Expect(responseBody.Status.Errors[0].Message).To(gm.Equal("SessionId must be provided in request header"))
			})
		})
	})

	// Test case for invalid X-Jabong-SessionId
	gk.Describe("GET"+allURL, func() {
		request := CreateTestRequest("GET", allURL)
		request.Header.Add("X-Jabong-SessionId", invalidSessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return invalid session id", func() {
				responseBody, _, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchHTTPCode(responseBody, fconstants.HTTPStatusBadRequestCode)
				gm.Expect(responseBody.Status.Errors[0].Code).To(gm.Equal(fconstants.ParamsInValidErrorCode))
				gm.Expect(responseBody.Status.Errors[0].Message).To(gm.Equal("SessionId is invalid"))
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
				responseBody, addressResult, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(3))
			})
		})
	})

	// Test case for GET /v1/address/all?limit=1
	gk.Describe("GET"+allURL+"?limit=1", func() {
		request := CreateTestRequest("GET", allURL+"?limit=1")
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 1 address", func() {
				responseBody, addressResult, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(1))
			})
		})
	})

	// Test case for GET /v1/address/all?offset=1
	gk.Describe("GET"+allURL+"?offset=1", func() {
		request := CreateTestRequest("GET", allURL+"?offset=1")
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 2 addresses", func() {
				responseBody, addressResult, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(2))
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
				fmt.Printf("%+v,\n%s", addressList, response.Body.String())
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
		request.Header.Add("X-Jabong-UserId", invalidUserID)
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
		request.Header.Add("X-Jabong-UserId", invalidUserID)
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
		request.Header.Add("X-Jabong-UserId", invalidUserID)
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
		request.Header.Add("X-Jabong-UserId", invalidUserID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return 0 addresses", func() {
				responseBody, addressResult, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				gm.Expect(addressResult.Summary.Count).To(gm.Equal(0))
			})
		})
	})

	// Test case for PUT /v1/address with missing addressId in path params
	gk.Describe("PUT"+baseURL, func() {
		request := CreateTestRequest("PUT", baseURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return versionable not found", func() {
				MatchHeaderStatus(response)
				responseBody, _, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchVersionableNotFound(responseBody)
			})
		})
	})

	// Test case for PUT /v1/address with missing request body
	putURL := baseURL + updateAddressID
	gk.Describe("PUT"+putURL, func() {
		request := CreateTestRequest("PUT", putURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		request.Body = ioutil.NopCloser(strings.NewReader(""))
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return missing request body", func() {
				responseBody, _, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchHTTPCode(responseBody, fconstants.HTTPStatusBadRequestCode)
				gm.Expect(responseBody.Status.Errors[0].Code).To(gm.Equal(fconstants.IncorrectDataErrorCode))
				gm.Expect(responseBody.Status.Errors[0].Message).To(gm.Equal("unexpected end of JSON input"))
			})
		})
	})

	// Test case for PUT /v1/address
	gk.Describe("PUT"+putURL, func() {
		request := CreateTestRequest("PUT", putURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		payload, _ := ioutil.ReadFile("../../config/testdata/put.json")
		request.Body = ioutil.NopCloser(strings.NewReader(string(payload)))
		response := GetResponse(request)
		var expectedResponse AddressRequest
		json.Unmarshal(payload, &expectedResponse)

		gk.Context("then the response", func() {
			gk.It("should return successs", func() {
				responseBody, _, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)

				// Create a new GET request to check if the address was updated successfully
				request = CreateTestRequest("GET", allURL)
				request.Header.Add("X-Jabong-SessionId", sessionID)
				request.Header.Add("X-Jabong-UserId", userID)
				response = GetResponse(request)
				_, _, _, addressList := GetHTTPResponseAndAddressResult(response.Body.String())
				for _, addressResponse := range addressList {
					if addressResponse.Id == updateAddressID {
						matchPayloadWithResponse(addressResponse, expectedResponse)
					}
				}
			})
		})
	})

	// Test case for PUT /v1/address?default=1
	gk.Describe("PUT"+putURL+"?default=1", func() {
		request := CreateTestRequest("PUT", putURL+"?default=1")
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		payload, _ := ioutil.ReadFile("../../config/testdata/put.json")
		request.Body = ioutil.NopCloser(strings.NewReader(string(payload)))
		response := GetResponse(request)
		var expectedResponse AddressRequest
		json.Unmarshal(payload, &expectedResponse)

		gk.Context("then the response", func() {
			gk.It("should return successs", func() {
				responseBody, _, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)

				// Create a new GET request to check if the address was updated successfully
				request = CreateTestRequest("GET", allURL)
				request.Header.Add("X-Jabong-SessionId", sessionID)
				request.Header.Add("X-Jabong-UserId", userID)
				response = GetResponse(request)
				_, _, _, addressList := GetHTTPResponseAndAddressResult(response.Body.String())
				for _, addressResponse := range addressList {
					if addressResponse.Id == updateAddressID {
						matchPayloadWithResponse(addressResponse, expectedResponse)
					}
				}
				// Check that the new address is set as the default
				gm.Expect(addressList[updateAddressID].IsDefaultShipping).To(gm.Equal("1"))
				gm.Expect(addressList[updateAddressID].IsDefaultBilling).To(gm.Equal("1"))
				// Also ensure that the older default shipping address was reset
				gm.Expect(addressList[oldDefaultAddressID].IsDefaultShipping).To(gm.Equal("0"))
				gm.Expect(addressList[oldDefaultAddressID].IsDefaultBilling).To(gm.Equal("0"))
			})
		})
	})

	// Test case for POST with missing body
	postURL := baseURL
	gk.Describe("POST"+postURL, func() {
		request := CreateTestRequest("POST", postURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		request.Body = ioutil.NopCloser(strings.NewReader(""))
		response := GetResponse(request)

		gk.Context("then the response", func() {
			gk.It("should return missing request body", func() {
				responseBody, _, _, _ := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchHTTPCode(responseBody, fconstants.HTTPStatusBadRequestCode)
				gm.Expect(responseBody.Status.Errors[0].Code).To(gm.Equal(fconstants.IncorrectDataErrorCode))
				gm.Expect(responseBody.Status.Errors[0].Message).To(gm.Equal("unexpected end of JSON input"))
			})
		})
	})

	// Test case for POST
	gk.Describe("POST"+postURL, func() {
		request := CreateTestRequest("POST", postURL)
		request.Header.Add("X-Jabong-SessionId", sessionID)
		request.Header.Add("X-Jabong-UserId", userID)
		payload, _ := ioutil.ReadFile("../../config/testdata/post.json")
		request.Body = ioutil.NopCloser(strings.NewReader(string(payload)))
		response := GetResponse(request)
		var expectedResponse AddressRequest
		json.Unmarshal(payload, &expectedResponse)

		gk.Context("then the response", func() {
			gk.It("should return successs", func() {
				responseBody, _, _, addressList := GetHTTPResponseAndAddressResult(response.Body.String())
				MatchSuccessResponseStatus(responseBody)
				// Check that the new address is not default and that values are same as input values
				for _, addressResponse := range addressList {
					gm.Expect(addressResponse.IsDefaultBilling).To(gm.Equal("0"))
					gm.Expect(addressResponse.IsDefaultShipping).To(gm.Equal("0"))
					matchPayloadWithResponse(addressResponse, expectedResponse)
				}
			})
		})
	})

	// Test case for POST /v1/address?default=1
})
