package address

import (
	"common/appconstant"
	"fmt"

	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
)

func GetAddressList(params *RequestParams, debugInfo *Debug) (*AddressResult, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_accessor-GetAddressList")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_accessor-GetAddressList"})
	}()

	rc := params.RequestContext
	userID := rc.UserID
	a := new(AddressResult)

	var (
		addressType   string = appconstant.All
		addressResult []AddressResponse
		err           error
	)

	if params.QueryParams.AddressType != "" {
		addressType = params.QueryParams.AddressType
	}
	addressResult, err = getAddressListFromCache(userID, params.QueryParams, debugInfo)
	if len(addressResult) == 0 || addressResult == nil || err != nil {
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "GetAddressList.Err", Value: err.Error()})
		logger.Error(fmt.Sprintf("Error in getting addresslist from cache. Error::" + err.Error()))
		addressResult, err = getAddressList(params, "", debugInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("error in getting the address list - %v", err))
			return a, err
		}
	}
	//Filter Result as per given limit and offset [Start]
	// start := params.QueryParams.Offset
	// end := params.QueryParams.Offset + params.QueryParams.Limit
	// addressFiltered := make([]AddressResponse, 0)

	// if params.QueryParams.AddressType == ALL || params.QueryParams.AddressType == "" {
	// 	if params.QueryParams.Limit != 0 {
	// 		addressFiltered = addressResult //[start:end]
	// 	}
	// } else {
	// 	for _, v := range addressResult {
	// 		if v.AddressType == params.QueryParams.AddressType {
	// 			addressFiltered = append(addressFiltered, v)
	// 		}
	// 	}
	// }
	// if end > len(addressFiltered) {
	// 	end = len(addressFiltered)
	// }
	// if start > end {
	// 	start = end
	// }
	// addressResult = addressFiltered[start:end]

	// //[closed]

	a.AddressList = addressResult
	a.Summery = AddressDetails{Count: len(addressResult), Type: addressType}
	return a, nil
}
