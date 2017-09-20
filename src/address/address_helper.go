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

func mergeDecryptedFieldsWithAddressResult(ef []DecryptedFields, address *map[string]*AddressResponse) {
	val := (*address)
	for i := 0; i < len(ef); i++ {
		val[ef[i].Id].Phone = ef[i].DecryptedPhone
		val[ef[i].Id].AlternatePhone = ef[i].DecryptedAlternatePhone
	}
}

//GetAddressListCacheKey return the cache key to get/set user addresses
func GetAddressListCacheKey(userID string) string {
	return fmt.Sprintf(appconstant.ADDRESS_CACHE_KEY, userID)
}

//getAddressListFromCache get user's address list from cache
func getAddressListFromCache(userId string, params QueryParams, debugInfo *Debug) (map[string]*AddressResponse, error) {
	p := profiler.NewProfiler()
	p.StartProfile("AddressHelper#getAddressListFromCache")

	defer func() {
		p.EndProfileWithMetric([]string{"AddressHelper#getAddressListFromCache"})
	}()

	addressList := make(map[string]*AddressResponse)
	var address map[string]*AddressResponse
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
	}

	var cacheKey string
	cacheKey = GetAddressListCacheKey(id)
	str, _ := json.Marshal(value)

	item := cache.Item{}
	item.Key = cacheKey
	item.Value = string(str)

	err := cacheObj.Set(item, false, false)
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
	index := fmt.Sprintf("%d", params.QueryParams.AddressId)
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "udpateAddressInCache:id", Value: fmt.Sprintf("%d", index)})
	addressList[index].IsOffice = params.QueryParams.Address.IsOffice
	addressList[index].FirstName = address.FirstName
	addressList[index].Phone = address.Phone
	addressList[index].Address1 = address.Address1
	addressList[index].City = address.City
	addressList[index].AddressRegion = address.AddressRegion
	addressList[index].PostCode = address.PostCode
	addressList[index].SmsOpt = address.SmsOpt
	addressList[index].UpdatedAt = time.Now().Format(appconstant.DATETIME_FORMAT)

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
func updateAddressListInCache(params *RequestParams, addressID string, debug *Debug) {
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
	addressList[addressID] = address[addressID]
	err = saveDataInCache(userID, addressList)

	debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "saveDataInCache:cacheKey", Value: GetAddressListCacheKey(userID)})

	if err != nil {
		debug.MessageStack = append(debug.MessageStack, DebugInfo{Key: "updateAddressListInCache.saveDataInCache:Err", Value: err.Error()})
		logger.Error(fmt.Sprintf("Could not update addressList in cache. %s", err.Error()), rc)
	}
}

func deleteAddressFromCache(params *RequestParams, debugInfo *Debug) (address map[string]*AddressResponse, err error) {
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
	delete(addressList, addressId)
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "deleteAddressFromCache:id", Value: addressId})
	err = saveDataInCache(userId, addressList)
	if err != nil {
		logger.Error(fmt.Sprintf("deleteAddressFromCache: Could not update address list in Cache while deleting. "+err.Error()), rc)
		return address, errors.New("Could not update address list in Cache while deleting. " + err.Error())
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
	debugInfo.MessageStack = append(debugInfo.MessageStack, DebugInfo{Key: "udpateTypeInCache:id", Value: addressID})
	if params.QueryParams.AddressType == appconstant.BILLING {
		addressList[addressID].IsDefaultBilling = "1"
	} else {
		addressList[addressID].IsDefaultShipping = "1"
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
