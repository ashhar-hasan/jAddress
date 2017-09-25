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
