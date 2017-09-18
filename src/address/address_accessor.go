package address

import (
	"common/appconfig"
	"common/appconstant"
	"errors"
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
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_accessor-GetAddressList")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_accessor-GetAddressList"})
	}()

	rc := params.RequestContext
	userID := rc.UserID
	a := new(AddressResult)

	var (
		addressType   string = appconstant.ALL
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
	start := params.QueryParams.Offset
	end := params.QueryParams.Offset + params.QueryParams.Limit
	addressFiltered := make([]AddressResponse, 0)

	if params.QueryParams.AddressType == "all" || params.QueryParams.AddressType == "" {
		if params.QueryParams.Limit != 0 {
			addressFiltered = addressResult //[start:end]
		}
	} else {
		for _, v := range addressResult {
			if params.QueryParams.AddressType == "billing" {
				if v.IsDefaultBilling == "1" {
					addressFiltered = append(addressFiltered, v)
				}
			} else if params.QueryParams.AddressType == "shipping" {
				if v.IsDefaultShipping == "1" {
					addressFiltered = append(addressFiltered, v)
				}
			} else if params.QueryParams.AddressType == "other" {
				if v.IsDefaultShipping == "0" && v.IsDefaultBilling == "0" {
					addressFiltered = append(addressFiltered, v)
				}
			}
		}
	}
	if end > len(addressFiltered) {
		end = len(addressFiltered)
	}
	if start > end {
		start = end
	}
	addressResult = addressFiltered[start:end]
	a.AddressList = addressResult
	a.Summery = AddressDetails{Count: len(addressResult), Type: addressType}
	return a, nil
}

func UpdateAddress(params *RequestParams, debugInfo *Debug) (*AddressResult, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_accessor-UpdateAddress")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_accessor-UpdateAddress"})
	}()

	rc := params.RequestContext

	cacheErr := udpateAddressInCache(params, debugInfo)
	if cacheErr != nil {
		cacheKey := GetAddressListCacheKey(params.RequestContext.UserID)
		err := invalidateCache(cacheKey)
		logger.Error(fmt.Sprintf("UpdateAddress: Error while invalidating the cache key %s, %v", cacheKey, err), rc)
	}
	go updateAddressInDb(params, debugInfo)
	a := new(AddressResult)
	return a, nil
}
func UpdateType(params *RequestParams, debugInfo *Debug) (*AddressResult, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_accessor-UpdateType")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_accessor-UpdateType"})
	}()
	cacheErr := updateTypeInCache(params, debugInfo)
	if cacheErr != nil {
		cacheKey := GetAddressListCacheKey(params.RequestContext.UserID)
		err := invalidateCache(cacheKey)
		logger.Error(fmt.Sprintf("UpdateAddress: Error while invalidating the cache key %s %v", cacheKey, err))
	}
	e := make(chan error, 0)
	go updateType(params, debugInfo, e)

	err := <-e
	if err != nil {
		return nil, err
	}
	a := new(AddressResult)
	return a, nil
}
func AddAddress(params *RequestParams, debugInfo *Debug) (*AddressResult, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_accessor-AddAddress")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_accessor-AddAddress"})
	}()
	a := new(AddressResult)

	rc := params.RequestContext
	userId := rc.UserID
	addressData := params.QueryParams.Address

	lastInsertedId, err := addAddress(userId, addressData, debugInfo)

	if err != nil {
		logger.Error(fmt.Sprintf("Error in adding new address in db - %v", err), rc)
		return a, errors.New("Some error occured while adding new address")
	}
	id := fmt.Sprintf("%d", lastInsertedId)
	addressResult, err := getAddressList(params, id, debugInfo)
	if err != nil {
		logger.Warning(fmt.Sprintf("Some error occured while getting address details after adding new address"), rc)
	}
	a.AddressList = addressResult

	go updateAddressListInCache(params, fmt.Sprintf("%d", lastInsertedId), debugInfo)

	return a, nil
}

func DeleteAddress(params *RequestParams, debugInfo *Debug) (*AddressResult, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_accessor-DeleteAddress")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_accessor-DeleteAddress"})
	}()
	a := new(AddressResult)
	addressResult, cacheErr := deleteAddressFromCache(params, debugInfo)
	if cacheErr != nil {
		rc := params.RequestContext
		cacheKey := GetAddressListCacheKey(params.RequestContext.UserID)
		err := invalidateCache(cacheKey)
		logger.Error(fmt.Sprintf("DeleteAddress: Error while invalidating the cache key %s, %v", cacheKey, err), rc)
	}
	e := make(chan error, 0)

	go deleteAddress(params, cacheErr, debugInfo, e) //Delete Adddress From DB

	err := <-e
	if err != nil {
		return nil, err
	}
	a.AddressList = addressResult
	return a, nil
}
