package address

import (
	"common/appconstant"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"time"

	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	utilHttp "github.com/jabong/florest-core/src/common/utils/http"
	"github.com/jabong/florest-core/src/components/cache"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

//DebugInfo represents a message
type DebugInfo struct {
	Key   string
	Value string
}

//Debug represents response
type Debug struct {
	MessageStack []DebugInfo
}

//addDebugContents add debug contents in the workflow data
func addDebugContents(io workflow.WorkFlowData, debug *Debug) {
	for k, v := range debug.MessageStack {
		io.ExecContext.SetDebugMsg(v.Key, v.Value)
		logger.Info(fmt.Sprintf("params.DebugInfo %v- %v", k, v))
	}
}

//urlEncode url encode the given string with data
func urlEncode(reqURL string, data []string) string {
	var URL *url.URL
	URL, _ = url.Parse(reqURL)
	parameters := url.Values{}
	for _, v := range data {
		parameters.Add("q", v)
	}
	URL.RawQuery = parameters.Encode()
	reqURL = URL.String()
	return reqURL
}

//Decrypt to decrypt an encrypted string
func Decrypt(encryptedData []string, debugInfo *Debug) []string {
	p := profiler.NewProfiler()
	p.StartProfile("AddressHelper#Decrypt")

	defer func() {
		p.EndProfileWithMetric([]string{"AddressHelper#Decrypt"})
	}()

	var (
		err               error
		res               []byte
		partialData, data []string
	)
	batchSize := int(appconstant.BATCH_SIZE)
	length := len(encryptedData)

	if length > batchSize {
		modulo := length % batchSize
		loop := math.Ceil(float64(length) / appconstant.BATCH_SIZE)
		loops := int(loop)
		for i := 0; i < loops; i++ {
			if modulo != 0 && i == (loops-1) {
				partialData = encryptedData[i*batchSize : (i*batchSize)+modulo]
			} else {
				partialData = encryptedData[i*batchSize : (i*batchSize)+batchSize]
			}
			res, err = encryptServiceObj.DecryptData(partialData, debugInfo)
			if err != nil {
				logger.Error("Decrypt: PartialResponse:: Data Decryption Error ", err.Error())
				for k := 0; k < len(partialData); k++ {
					data = append(data, "")
				}
			} else {
				d, _ := getDataFromServiceResponse(res)
				data = append(data, d...)
			}

		}
	} else {
		res, err = encryptServiceObj.DecryptData(encryptedData, debugInfo)
		if err != nil {
			logger.Error("Decrypt: Data Decryption Error ", err.Error())
			return data
		}
		data, err = getDataFromServiceResponse(res)
		if err != nil {
			logger.Error(fmt.Sprintf("Decrypt: getDataFromServiceResponse() Error:: %+v", err))
			return data
		}
	}

	return data
}

//getDataFromServiceResponse to parse the encryption/decryption service response
func getDataFromServiceResponse(body []byte) (data []string, err error) {
	p := profiler.NewProfiler()
	p.StartProfile("AddressHelper#getDataFromServiceResponse")

	defer func() {
		p.EndProfileWithMetric([]string{"AddressHelper#getDataFromServiceResponse"})
	}()

	r := utilHttp.Response{}
	err = json.Unmarshal(body, &r)

	if err != nil {
		logger.Error(fmt.Sprintf("decodeJson error is = %v", err))
		return data, err
	}

	tempData, _ := r.Data.([]interface{})
	var res []string
	for _, val := range tempData {
		d, ok := val.(string)
		if d == "" {
			res = append(res, "0")
		} else {
			res = append(res, d)
		}
		if !ok {
			logger.Error(fmt.Sprintf("Type Assertion fail. Error: %v, Data: %v", err, data))
		}
	}
	return res, nil
}

func decryptEncryptedFields(ef []EncryptedFields, params *RequestParams, debug *Debug) ([]DecryptedFields, error) {
	var (
		encryptedPhoneString    []string
		encryptedAltPhoneString []string
		dp, dap                 string
	)
	for _, v := range ef {
		encryptedPhoneString = append(encryptedPhoneString, v.EncryptedPhone)
		encryptedAltPhoneString = append(encryptedAltPhoneString, v.EncryptedAlternatePhone)
	}
	decryptedPhone := Decrypt(encryptedPhoneString, debug)
	decryptedAltPhone := Decrypt(encryptedAltPhoneString, debug)
	res := make([]DecryptedFields, 0)
	if len(decryptedPhone) > 0 {
		for k, v := range ef {
			if decryptedPhone[k] != "" {
				dp = decryptedPhone[k]
			} else {
				dp = ""
			}

			if decryptedAltPhone[k] != "" {
				dap = decryptedAltPhone[k]
			} else {
				dap = ""
			}
			res = append(res, DecryptedFields{Id: v.Id, DecryptedPhone: dp, DecryptedAlternatePhone: dap})
		}
	}
	if len(res) == 0 {
		return nil, errors.New("Error in Decrypting Encryption Fields")
	}
	return res, nil
}

func mergeDecryptedFieldsWithAddressResult(ef []DecryptedFields, address *[]AddressResponse) {
	val := (*address)
	for k := range val {
		temp := &val[k]
		if temp.Id == ef[k].Id {
			temp.Phone = ef[k].DecryptedPhone
			if ef[k].DecryptedAlternatePhone == "0" {
				temp.AlternatePhone = ""
			} else {
				temp.AlternatePhone = ef[k].DecryptedAlternatePhone
			}
		}
	}
}

//GetAddressListCacheKey return the cache key to get/set user addresses
func GetAddressListCacheKey(userID string) string {
	return fmt.Sprintf(appconstant.ADDRESS_CACHE_KEY, userID)
}

//getAddressListFromCache get user's address list from cache
func getAddressListFromCache(userId string, params QueryParams, debugInfo *Debug) ([]AddressResponse, error) {
	p := profiler.NewProfiler()
	p.StartProfile("AddressHelper#getAddressListFromCache")

	defer func() {
		p.EndProfileWithMetric([]string{"AddressHelper#getAddressListFromCache"})
	}()

	var address []AddressResponse

	cacheObj, errG := cache.Get(cache.Redis)
	if errG != nil {
		msg := fmt.Sprintf("Redis Config Error - %v", errG)
		logger.Error(msg)
	}
	addressListCacheKey := GetAddressListCacheKey(userId)

	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "getAddressListFromCache:addressListCacheKey", Value: addressListCacheKey})
	result, err := cacheObj.Get(addressListCacheKey, false, false)
	if err != nil {
		logger.Error(fmt.Sprintf("Error while getting address list from cache %s", err))
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "getAddressListFromCache:Error", Value: "Error while getting address list from cache::" + err.Error()})
		return address, err
	}

	var addressList []AddressResponse
	data := result.Value.(string)
	byt := []byte(data)
	if err := json.Unmarshal(byt, &addressList); err != nil {
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "getAddressListFromCache.UnmarshalErr", Value: err.Error()})
		return address, err
	}
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "getAddressListFromCache:Result", Value: fmt.Sprintf("%+v", addressList)})

	return addressList, nil
}

//saveDataInCache save data in cache
func saveDataInCache(id string, value interface{}) error {
	p := profiler.NewProfiler()
	p.StartProfile("AddressHelper#saveDataInCache")

	defer func() {
		p.EndProfileWithMetric([]string{"AddressHelper#saveDataInCache"})
	}()
	cacheObj, errG := cache.Get(cache.Redis)
	if errG != nil {
		msg := fmt.Sprintf("Redis Config Error - %v", errG)
		logger.Error(msg)
		return errG
	}

	var cacheKey string
	cacheKey = GetAddressListCacheKey(id)

	addressFiltered := orderSaveData(value) // To maintain the order of billing,shipping,other
	str, _ := json.Marshal(addressFiltered)

	item := cache.Item{}
	item.Key = cacheKey
	item.Value = string(str)

	err := cacheObj.SetWithTimeout(item, false, false, appconstant.EXPIRATION_TIME)
	if err != nil {
		return err
	}

	return nil
}

//udpateAddressInCache update/edit user's particular address in cache
func udpateAddressInCache(params *RequestParams, debugInfo *Debug) error {
	p := profiler.NewProfiler()
	p.StartProfile("AddressHelper#updateAddressInCache")

	defer func() {
		p.EndProfileWithMetric([]string{"AddressHelper#updateAddressInCache"})
	}()

	rc := params.RequestContext
	userID := rc.UserID
	address := params.QueryParams.Address

	addressList, err := getAddressListFromCache(userID, params.QueryParams, debugInfo)

	if err != nil {
		logger.Error(fmt.Sprintf("Error while fetching address list from Cache"), rc)
		return errors.New("Error while fetching address list from Cache")
	}

	var (
		index int
		flag  bool
	)
	addressID := fmt.Sprintf("%d", params.QueryParams.AddressId)
	for key, value := range addressList {
		if value.Id == addressID {
			index = key
			flag = true
			break
		}
	}
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "udpateAddressInCache:index", Value: fmt.Sprintf("%d", index)})
	if flag {
		addressList[index].Id = addressID
		addressList[index].IsOffice = params.QueryParams.AddressType
		addressList[index].FirstName = address.FirstName
		addressList[index].Phone = address.Phone
		addressList[index].Address1 = address.Address1
		addressList[index].City = address.City
		addressList[index].AddressRegion = address.AddressRegion
		addressList[index].PostCode = address.PostCode
		addressList[index].SmsOpt = address.SmsOpt

		if address.Country != "" {
			addressList[index].Country = address.Country
		}
		if address.RegionName != "" {
			addressList[index].RegionName = address.RegionName
		}
		if address.LastName != "" {
			addressList[index].LastName = address.LastName
		}
		if address.Address2 != "" {
			addressList[index].Address2 = address.Address2
		}
		if address.AlternatePhone != "" {
			addressList[index].AlternatePhone = address.AlternatePhone
		}
	}
	if index == 0 && addressList[index].IsDefaultShipping == "1" {
		addressList[1] = addressList[0]
	}
	err = saveDataInCache(userID, addressList)
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "saveDataInCache:cacheKey", Value: GetAddressListCacheKey(userID)})
	if err != nil {
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "udpateAddressInCache:saveDataInCache.Err", Value: err.Error()})
		logger.Error(fmt.Sprintf("udpateAddressInCache: Could not update address in Cache"), rc)
		return errors.New("Could not update address in Cache")
	}
	return nil
}

//updateAddressListInCache re-populate address list in cache
func updateAddressListInCache(params *RequestParams, addressID string, debug *Debug) []AddressResponse {
	p := profiler.NewProfiler()
	p.StartProfile("AddressHelper#updateAddressListInCache")

	defer func() {
		p.EndProfileWithMetric([]string{"AddressHelper#updateAddressListInCache"})
	}()

	rc := params.RequestContext
	userID := rc.UserID
	address, _ := getAddressList(params, addressID, debug)

	addressList, err := getAddressListFromCache(userID, params.QueryParams, debug)
	if err != nil {
		logger.Error(fmt.Sprintf("updateAddressListInCache::Error while fetching address list from Cache"), rc)
	}
	var (
		index int
		flag  bool
	)
	for key, value := range addressList {
		addID := fmt.Sprintf("%d", value.Id)
		if addID == addressID {
			index = key
			flag = true
			break
		}
	}
	if flag {
		debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "updateAddressListInCache:Deleting an index", Value: fmt.Sprintf("%d", index)})
		addressList = append(addressList[:index], addressList[index+1:]...)
	}
	addressList = append(addressList, address...)
	err = saveDataInCache(userID, addressList)

	debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "saveDataInCache:cacheKey", Value: GetAddressListCacheKey(userID)})

	if err != nil {
		debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "updateAddressListInCache.saveDataInCache:Err", Value: err.Error()})
		logger.Error(fmt.Sprintf("Could not update addressList in cache. %s", err.Error()), rc)
	}
	return address
}

func deleteAddressFromCache(params *RequestParams, debugInfo *Debug) (address []AddressResponse, err error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("address-address_accessor-DeleteAddress")
	defer func() {
		prof.EndProfileWithMetric([]string{"address-address_accessor-DeleteAddress"})
	}()
	rc := params.RequestContext
	userId := rc.UserID
	addressId := fmt.Sprintf("%d", params.QueryParams.AddressId)

	addressList, err := getAddressListFromCache(userId, params.QueryParams, debugInfo)
	if err != nil {
		logger.Error(fmt.Sprintf("deleteAddressFromCache: Could not retrieve address list from Cache"), rc)
		return address, errors.New("Could not retrieve address list from Cache")
	}

	var (
		index int
		flag  bool
	)
	for key, value := range addressList {
		if value.Id == addressId {
			index = key
			flag = true
			break
		}
	}
	if flag {
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "deleteAddressFromCache:index", Value: fmt.Sprintf("%d", index)})
		addressList = append(addressList[:index], addressList[index+1:]...)
		err = saveDataInCache(userId, addressList)
		if err != nil {
			logger.Error(fmt.Sprintf("deleteAddressFromCache: Could not update address list in Cache while deleting. "+err.Error()), rc)
			return address, errors.New("Could not update address list in Cache while deleting. " + err.Error())
		}
	}
	return addressList, nil
}

//invalidateCache invalidate cache key
func invalidateCache(key string) error {
	cacheObj, errG := cache.Get(cache.Redis)
	if errG != nil {
		msg := fmt.Sprintf("Redis Config Error - %v", errG)
		logger.Error(msg)
	}

	err := cacheObj.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func updateTypeInCache(params *RequestParams, debugInfo *Debug) error {
	p := profiler.NewProfiler()
	p.StartProfile("AddressHelper#updateTypeInCache")

	defer func() {
		p.EndProfileWithMetric([]string{"AddressHelper#updateTypeInCache"})
	}()

	rc := params.RequestContext
	userID := rc.UserID
	addressID := fmt.Sprintf("%d", params.QueryParams.AddressId)
	addressList, err := getAddressListFromCache(userID, params.QueryParams, debugInfo)

	if err != nil {
		msg := "Error while fetching address list from cache"
		logger.Error(msg)
		return errors.New(msg)
	}
	var (
		index int
		flag  bool
	)

	for key, value := range addressList {
		if value.Id == addressID {
			index = key
			flag = true
			break
		}
	}
	if addressList[index].IsDefaultBilling == "1" && params.QueryParams.AddressType == appconstant.BILLING {
		return nil
	} else if addressList[index].IsDefaultShipping == "1" && params.QueryParams.AddressType == appconstant.SHIPPING {
		return nil
	}
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "udpateTypeInCache:index", Value: fmt.Sprintf("%d", index)})
	if flag {
		// addressList[index].AddressType = params.QueryParams.Address.AddressType
		if params.QueryParams.AddressType == appconstant.BILLING {
			for k, _ := range addressList {
				if addressList[k].IsDefaultBilling == "1" {
					addressList[k].IsDefaultBilling = "0"
					addressList[k].UpdatedAt = time.Now().Format(appconstant.DATETIME_FORMAT)
				}
			}
			addressList[index].UpdatedAt = time.Now().Format(appconstant.DATETIME_FORMAT)
			addressList[index].IsDefaultBilling = "1"
		} else {
			for k, _ := range addressList {
				if addressList[k].IsDefaultShipping == "1" {
					addressList[k].IsDefaultShipping = "0"
					addressList[k].UpdatedAt = time.Now().Format(appconstant.DATETIME_FORMAT)
				}
			}
			addressList[index].UpdatedAt = time.Now().Format(appconstant.DATETIME_FORMAT)
			addressList[index].IsDefaultShipping = "1"
		}
	}
	err = saveDataInCache(userID, addressList)
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "saveDataInCache:cacheKey", Value: GetAddressListCacheKey(userID)})
	if err != nil {
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "udpateTypeInCache:saveDataInCache.Err", Value: err.Error()})
		logger.Error(fmt.Sprintf("udpateTypeInCache: Could not update address in Cache"), rc)
		return errors.New("Could not update address type in Cache")
	}
	return nil
}

func getFilteredListFromCache(userId string, params QueryParams, debugInfo *Debug) ([]interface{}, error) {
	p := profiler.NewProfiler()
	p.StartProfile("AddressHelper#getFilteredListFromCache")

	defer func() {
		p.EndProfileWithMetric([]string{"AddressHelper#getFilteredListFromCache"})
	}()

	var address []interface{}
	cacheObj, errG := cache.Get(cache.Redis)
	if errG != nil {
		msg := fmt.Sprintf("Redis Config Error - %v", errG)
		logger.Error(msg)
	}
	addressListCacheKey := GetAddressListCacheKey(userId)

	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "getAddressListFromCache:addressListCacheKey", Value: addressListCacheKey})
	result, err := cacheObj.Get(addressListCacheKey, false, false)
	if err != nil {
		logger.Error(fmt.Sprintf("Error while getting address list from cache %s", err))
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "getAddressListFromCache:Error", Value: "Error while getting address list from cache::" + err.Error()})
		return address, err
	}

	var addressList []interface{}
	data := result.Value.(string)
	byt := []byte(data)
	if err := json.Unmarshal(byt, &addressList); err != nil {
		debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "getAddressListFromCache.UnmarshalErr", Value: err.Error()})
		return address, err
	}
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "getAddressListFromCache:Result", Value: fmt.Sprintf("%+v", addressList)})
	return addressList, nil
}

func orderSaveData(value interface{}) []AddressResponse {
	data, _ := value.([]AddressResponse)

	if len(data) == 1 {
		data = append(data, data[0])
		both = true
		return data
	} else {
		firstBilling, firstShipping := data[0].IsDefaultBilling, data[0].IsDefaultShipping
		secondBilling, secondShipping := data[1].IsDefaultBilling, data[1].IsDefaultShipping
		if firstBilling == "1" && firstShipping == "1" && secondBilling == "1" && secondShipping == "1" {
			both = true
			return data
		} else if firstBilling == "1" && firstShipping == "1" && secondBilling == "0" && secondShipping == "0" {
			temp := data[1]
			data[1] = data[0]
			data = append(data, temp)
			both = true
			return data
		} else if firstBilling == "1" && firstShipping == "0" && secondBilling == "1" && secondShipping == "0" {
			flag := false
			index := -1
			for i := 2; i < len(data); i++ {
				if data[i].IsDefaultShipping == "1" {
					flag = true
					index = i
					break
				}

			}
			if flag {
				data[1] = data[index]
				data = deleteIndexFromData(data, index)
			}
		} else if firstBilling == "0" && firstShipping == "0" && secondBilling == "1" && secondShipping == "1" {
			temp := data[0]
			data[0] = data[1]
			data = append(data, temp)
			both = true
			return data
		} else if firstBilling == "0" && firstShipping == "1" && secondBilling == "0" && secondShipping == "1" {
			flag := false
			index := -1
			for i := 2; i < len(data); i++ {
				if data[i].IsDefaultBilling == "1" {
					flag = true
					index = i
					break
				}
			}
			if flag {
				data[0] = data[index]
				data = deleteIndexFromData(data, index)
			}
		} else if firstBilling == "1" && firstShipping == "0" && secondBilling == "0" && secondShipping == "0" {
			index := -1
			for i := 2; i < len(data); i++ {
				if data[i].IsDefaultShipping == "1" {
					index = i
					break
				}
			}
			data[1], data[index] = data[index], data[1]
		} else if firstBilling == "0" && firstShipping == "0" && secondBilling == "0" && secondShipping == "1" {
			index := -1
			for i := 2; i < len(data); i++ {
				if data[i].IsDefaultBilling == "1" {
					index = i
					break
				}
			}
			data[0], data[index] = data[index], data[0]
		} else if firstBilling == "0" && firstShipping == "0" && secondBilling == "1" && secondShipping == "0" {
			data[0], data[1] = data[1], data[0]
			index := -1
			for i := 2; i < len(data); i++ {
				if data[i].IsDefaultShipping == "1" {
					index = i
					break
				}
			}
			data[1], data[index] = data[index], data[1]
		} else if firstBilling == "0" && firstShipping == "0" && secondBilling == "0" && secondShipping == "0" {
			index1 := -1
			index2 := -1
			for i := 2; i < len(data); i++ {
				if data[i].IsDefaultBilling == "1" {
					index1 = i
				}
				if data[i].IsDefaultShipping == "1" {
					index2 = i
				}
			}
			data[0], data[index1] = data[index1], data[0]
			data[1], data[index2] = data[index2], data[1]
			if index1 == index2 {
				temp := data[1]
				data[1] = data[0]
				data = append(data, temp)
				both = true
				return data
			}
		} else if firstBilling == "0" && firstShipping == "1" && secondBilling == "0" && secondShipping == "0" {
			data[1], data[0] = data[0], data[1]
			index := -1
			for i := 2; i < len(data); i++ {
				if data[i].IsDefaultBilling == "1" {
					index = i
					break
				}
			}
			data[0], data[index] = data[index], data[0]
		} else if firstBilling == "0" && firstShipping == "1" && secondBilling == "1" && secondShipping == "0" {
			data[1], data[0] = data[0], data[1]
		}
	}
	both = false
	return data
}

func deleteIndexFromData(data []AddressResponse, index int) []AddressResponse {
	if index == len(data)-1 {
		data = data[:(len(data) - 1)]
		return data
	}
	data = append(data[:index], data[index+1:]...)
	return data
}
