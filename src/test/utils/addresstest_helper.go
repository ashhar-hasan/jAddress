package utils

import (
	"address"
	"encoding/json"

	utilhttp "github.com/jabong/florest-core/src/common/utils/http"
	gm "github.com/onsi/gomega"
)

//GetHTTPResponseAndAddressResult parses the responseBody to return pointers the http response and search result
func GetHTTPResponseAndAddressResult(responseBody string) (*utilhttp.Response, *address.AddressResult) {
	var responeBody utilhttp.Response
	err := json.Unmarshal([]byte(responseBody), &responeBody)
	gm.Expect(err).To(gm.BeNil())

	byteArray, errMar := json.Marshal(responeBody.Data)
	gm.Expect(errMar).To(gm.BeNil())

	var addressResult address.AddressResult
	errUnMar := json.Unmarshal(byteArray, &addressResult)
	gm.Expect(errUnMar).To(gm.BeNil())
	return &responeBody, &addressResult
}
