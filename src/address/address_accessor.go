package address

import (
	"common/appconfig"
	"common/appconstant"
	"fmt"

	"github.com/jabong/florest-core/src/common/config"
	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
)

var encryptServiceObj *EncryptionService

//Initialise initialises Address Accessor
func Initialise() {
	var err error
	c := config.GlobalAppConfig.ApplicationConfig
	appConfig, _ := c.(*appconfig.AddressServiceConfig)
	encConfHost := appConfig.EncryptionServiceConfig.Host
	encConfReqTimeOut := appConfig.EncryptionServiceConfig.ReqTimeout
	encryptServiceObj, err = InitEncryptionService(encConfHost, encConfReqTimeOut)
	if err != nil {
		panic("Failed to initialise Encryption Service" + err.Error())
	}

	logger.Info(fmt.Sprintf("Address Service Accessor Initialize"))
}

func GetAddressList(params *RequestParams, debugInfo *Debug) (*AddressResult, error) {
	Initialise()
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
	a.AddressList = addressResult
	a.Summery = AddressDetails{Count: len(addressResult), Type: addressType}
	return a, nil
}
