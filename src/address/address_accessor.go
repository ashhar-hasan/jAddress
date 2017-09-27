package address

import (
	"common/appconfig"
	"common/appconstant"
	"errors"
	"fmt"
	"sort"
	"strconv"

	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	"github.com/jabong/florest-core/src/components/cache"
	"github.com/jabong/florest-core/src/components/sqldb"
)

var encryptServiceObj *EncryptionService

//Initialise initialises Address Accessor
func Initialise() {
	var err error
	appConfig, _ := appconfig.GetAddressServiceConfig()
	encryptServiceObj, err = InitEncryptionService(appConfig.EncryptionServiceConfig.Host, appConfig.EncryptionServiceConfig.ReqTimeout)
	if err != nil {
		panic("Failed to initialise Encryption Service" + err.Error())
	}
	if err = sqldb.Set("mysdb", appConfig.MySqlConfig.MySqlMaster, new(sqldb.MysqlDriver)); err != nil {
		logger.Error(err)
	}
	if err = cache.Set(cache.Redis, appConfig.Cache.Redis, new(cache.RedisClientAdapter)); err != nil {
		logger.Error(err)
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
		addressResult map[string]*AddressResponse
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
	addressFiltered := make(map[string]*AddressResponse, 0)

	if addressType == "all" || params.QueryParams.AddressType == "" {
		if params.QueryParams.Limit != 0 {
			addressFiltered = addressResult //[start:end]
		}
	} else {
		for k, v := range addressResult {
			if addressType == "other" {
				if v.IsDefaultShipping == "0" && v.IsDefaultBilling == "0" {
					addressFiltered[k] = v
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
	keys := make([]string, 0)
	for k, _ := range addressFiltered {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	temp := make(map[string]*AddressResponse)
	for i := start; i < end; i++ {
		temp[keys[i]] = addressFiltered[keys[i]]
	}
	a.AddressList = temp
	a.Summary = AddressDetails{Count: len(temp), Type: addressType}
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

	a := new(AddressResult)
	// Set as default shipping address
	if params.QueryParams.Default == 1 {
		_, err := UpdateType(params, debugInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("There is some error occured while updating the type %v", err), rc)
			return a, errors.New("Some error occurred while setting the default address")
		}
	}

	go updateAddressInDb(params, debugInfo)
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

	isFirst, _ := isFirstAddress(userId, debugInfo)
	lastInsertedId, err := addAddress(userId, addressData, debugInfo)
	params.QueryParams.AddressId = int(lastInsertedId)

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

	// Set as default shipping address
	if params.QueryParams.Default == 1 && isFirst == false {
		_, err := UpdateType(params, debugInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("There is some error occured while updating the type %v", err), rc)
			return a, errors.New("Some error occurred while setting the default address")
		}
	}

	return a, nil
}

func DeleteAddress(params *RequestParams, debugInfo *Debug) (*AddressResult, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_accessor-DeleteAddress")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_accessor-DeleteAddress"})
	}()
	a := new(AddressResult)
	flag, err1 := checkDefaultAddress(params, debugInfo)
	if err1 != nil {
		return nil, err1
	}
	if flag == 1 {
		return nil, errors.New("Cannot delete default billing address")
	} else if flag == 2 {
		return nil, errors.New("Select a different default delivery address first.")
	} else {
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
}

func GetAddressTypeList(params *RequestParams, debugInfo *Debug) (*AddressResult, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_accessor-GetAddressTypeList")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_accessor-GetAddressTypeList"})
	}()

	rc := params.RequestContext
	userID := rc.UserID
	a := new(AddressResult)
	addressType := params.QueryParams.AddressType
	addressResult, err := getAddressListFromCache(userID, params.QueryParams, debugInfo)
	if len(addressResult) == 0 || addressResult == nil || err != nil {
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "GetAddressList.Err", Value: err.Error()})
		logger.Error(fmt.Sprintf("Error in getting addresslist from cache. Error::" + err.Error()))
		addressResult, err = getAddressList(params, "", debugInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("error in getting the address list - %v", err))
			return a, err
		}
	}
	var index string
	for k, v := range addressResult {
		if addressType == "billing" && v.IsDefaultBilling == "1" {
			index = k
			break
		} else if addressType == "shipping" && v.IsDefaultShipping == "1" {
			index = k
			break
		}
	}

	a.AddressList = addressResult[index]
	if addressResult[index] != nil {
		a.Summary = AddressDetails{Count: 1, Type: addressType}
	} else {
		a.Summary = AddressDetails{Count: 0, Type: addressType}
	}
	return a, nil

}

func checkDefaultAddress(params *RequestParams, debugInfo *Debug) (int, error) {

	userID := params.RequestContext.UserID
	addressID := params.QueryParams.AddressId
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "CheckDefaultAddress", Value: "CheckDefaultAddress Execute"})
	addressResult, err := getAddressListFromCache(userID, params.QueryParams, debugInfo)
	if err != nil || len(addressResult) == 0 {
		logger.Info(fmt.Sprintf("Address not found in cache for addressID: %d", params.QueryParams.AddressId))
		val, err1 := checkDefaultAddressInDB(addressID, userID, debugInfo)
		if err1 != nil {
			return 0, err1
		}
		return val, nil
	}
	addID := strconv.Itoa(addressID)
	if addressResult[addID].IsDefaultBilling == "1" {
		return 1, nil
	} else if addressResult[addID].IsDefaultShipping == "1" {
		return 2, nil
	} else {
		return 0, nil
	}

}
