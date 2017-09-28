package address

import (
	"common/appconstant"
	"encoding/json"

	utilhttp "github.com/jabong/florest-core/src/common/utils/http"
	gm "github.com/onsi/gomega"
)

//GetHTTPResponseAndAddressResult parses the responseBody to return pointers the http response and search result
func GetHTTPResponseAndAddressResult(responseBody string) (*utilhttp.Response, *AddressResult, *AddressResponse, map[string]*AddressResponse) {
	var responeBody utilhttp.Response
	err := json.Unmarshal([]byte(responseBody), &responeBody)
	gm.Expect(err).To(gm.BeNil())

	byteArray, errMar := json.Marshal(responeBody.Data)
	gm.Expect(errMar).To(gm.BeNil())

	var addressResult AddressResult
	errUnMar := json.Unmarshal(byteArray, &addressResult)
	gm.Expect(errUnMar).To(gm.BeNil())

	byteArray, errMar = json.Marshal(addressResult.AddressList)
	gm.Expect(errMar).To(gm.BeNil())

	var address AddressResponse
	var addressList map[string]*AddressResponse
	if addressResult.Summary.Type == appconstant.BILLING || addressResult.Summary.Type == appconstant.SHIPPING {
		errUnMar = json.Unmarshal(byteArray, &address)
		gm.Expect(errUnMar).To(gm.BeNil())
	} else {
		errUnMar = json.Unmarshal(byteArray, &addressList)
		gm.Expect(errUnMar).To(gm.BeNil())
	}
	return &responeBody, &addressResult, &address, addressList
}

func sanitizeAddress(address *AddressRequest) {
	address.Address1 = sanitize(address.Address1, false)
	address.Address2 = sanitize(address.Address2, false)
	address.City = sanitize(address.City, false)
	address.FirstName = sanitize(address.FirstName, true)
	address.LastName = sanitize(address.LastName, true)
}

func matchPayloadWithResponse(response *AddressResponse, payload AddressRequest) {
	sanitizeAddress(&payload)
	gm.Expect(response.Address1).To(gm.Equal(payload.Address1))
	gm.Expect(response.Address2).To(gm.Equal(payload.Address2))
	gm.Expect(response.AddressRegion).To(gm.Equal(payload.AddressRegion))
	gm.Expect(response.AlternatePhone).To(gm.Equal(payload.AlternatePhone))
	gm.Expect(response.City).To(gm.Equal(payload.City))
	gm.Expect(response.Country).To(gm.Equal(payload.Country))
	gm.Expect(response.FirstName).To(gm.Equal(payload.FirstName))
	gm.Expect(response.IsOffice).To(gm.Equal(payload.IsOffice))
	gm.Expect(response.LastName).To(gm.Equal(payload.LastName))
	gm.Expect(response.Phone).To(gm.Equal(payload.Phone))
}
