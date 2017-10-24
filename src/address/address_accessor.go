package address

import (
	"common/appconfig"
	"common/appconstant"
	"errors"
	"fmt"
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
		addressResult []AddressResponse
		cacheResult   []interface{}
		err           error
	)

	if params.QueryParams.AddressType != "" {
		addressType = params.QueryParams.AddressType
	}
	var cache = true
	cacheResult, err = getFilteredListFromCache(userID, params.QueryParams, debugInfo)
	if len(cacheResult) == 0 || cacheResult == nil || err != nil {
		cache = false
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "GetAddressList.Err", Value: err.Error()})
		logger.Error(fmt.Sprintf("Error in getting addresslist from cache. Error::" + err.Error()))
		addressResult, err = getAddressList(params, "", debugInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("error in getting the address list - %v", err))
			return a, err
		}
		billing, shipping := false, false
		for k, v := range addressResult {
			if billing && shipping {
				break
			}
			if v.IsDefaultBilling == "1" && v.IsDefaultShipping == "1" {
				addressResult[0], addressResult[k] = addressResult[k], addressResult[0]
				if len(addressResult) == 1 {
					addressResult = append(addressResult, addressResult[0])
				} else {
					temp := addressResult[1]
					addressResult[1] = addressResult[0]
					addressResult = append(addressResult, temp)
				}
				both = true
				break
			}
			if !shipping && v.IsDefaultShipping == "1" {
				addressResult[1], addressResult[k] = addressResult[k], addressResult[1]
				shipping = true
				both = false
			}
			if !billing && v.IsDefaultBilling == "1" {
				addressResult[0], addressResult[k] = addressResult[k], addressResult[0]
				billing = true
				both = false
			}
		}
	}
	var result interface{}
	var length int
	//Because type of cacheResult and DB result is different
	if cache {
		result, length = filterCacheResult(cacheResult, params)
	} else {
		result, length = filterDBResult(addressResult, params)
	}
	a.AddressList = result
	a.Summary = AddressDetails{Count: length, Type: addressType}
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

	lastInsertedId, isFirst, err := addAddress(userId, addressData, debugInfo)
	params.QueryParams.AddressId = int(lastInsertedId)
	if err != nil {
		logger.Error(fmt.Sprintf("Error in adding new address in db - %v", err), rc)
		return a, errors.New("Some error occured while adding new address")
	}
	id := fmt.Sprintf("%d", lastInsertedId)
	addressResult := updateAddressListInCache(params, id, debugInfo)
	if params.QueryParams.Default == 1 && isFirst == false {
		_, err := UpdateType(params, debugInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("There is some error occured while updating the type %v", err), rc)
			return a, errors.New("Some error occurred while setting the default address")
		}
		addressResult, _ = getAddressList(params, id, debugInfo)
	}
	a.AddressList = addressResult
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

//checks the address to be deleted is default or not
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
	index := -1
	flag := false
	for k, v := range addressResult {
		if v.Id == addID {
			index = k
			flag = true
			break
		}
	}
	if flag {
		if addressResult[index].IsDefaultBilling == "1" {
			return 1, nil
		} else if addressResult[index].IsDefaultShipping == "1" {
			return 2, nil
		}
	}
	return 0, nil
}

func filterCacheResult(addressResult []interface{}, params *RequestParams) (interface{}, int) {
	length := len(addressResult)
	if length == 0 {
		return nil, 0
	}
	start := params.QueryParams.Offset
	end := params.QueryParams.Offset + params.QueryParams.Limit
	addressType := params.QueryParams.AddressType
	if addressType == "all" || addressType == "" {
		if both {
			if length > 2 {
				addressResult = append(addressResult[:1], addressResult[2:]...)
			} else {
				addressResult = addressResult[0:1]
			}
		}
	} else if addressType == "billing" {
		addressResult = addressResult[0:1]
	} else if addressType == "shipping" {
		if length > 2 {
			addressResult = addressResult[1:2]
		} else {
			addressResult = addressResult[1:]
		}
	} else if addressType == "other" {
		if length > 2 {
			addressResult = addressResult[2:]
		} else {
			addressResult = nil
		}
	}
	if end > len(addressResult) {
		end = len(addressResult)
	}
	if start > end {
		start = end
	}
	addressResult = addressResult[start:end]
	return addressResult, len(addressResult)
}

func filterDBResult(addressResult []AddressResponse, params *RequestParams) (interface{}, int) {
	length := len(addressResult)
	if length == 0 {
		return nil, 0
	}
	start := params.QueryParams.Offset
	end := params.QueryParams.Offset + params.QueryParams.Limit
	addressType := params.QueryParams.AddressType
	if addressType == "all" || addressType == "" {
		if both {
			if length > 2 {
				addressResult = append(addressResult[:1], addressResult[2:]...)
			} else {
				addressResult = addressResult[0:1]
			}
		}
	} else if addressType == "billing" {
		addressResult = addressResult[0:1]
	} else if addressType == "shipping" {
		if length > 2 {
			addressResult = addressResult[1:2]
		} else {
			addressResult = addressResult[1:]
		}
	} else if addressType == "other" {
		if length > 2 {
			addressResult = addressResult[2:]
		} else {
			addressResult = nil
		}
	}
	if end > len(addressResult) {
		end = len(addressResult)
	}
	if start > end {
		start = end
	}
	addressResult = addressResult[start:end]
	return addressResult, len(addressResult)
}
