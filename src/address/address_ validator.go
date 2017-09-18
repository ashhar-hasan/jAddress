package address

import (
	"common/appconstant"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	constants "github.com/jabong/florest-core/src/common/constants"
	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	utilHttp "github.com/jabong/florest-core/src/common/utils/http"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

type AddressValidator struct {
	id string
}

func (a *AddressValidator) SetID(id string) {
	a.id = id
}

func (a AddressValidator) GetID() (string, error) {
	return a.id, nil
}

func (a AddressValidator) Name() string {
	return "AddressValidator"
}

func (a AddressValidator) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("AddressValidator#Execute")

	defer func() {
		prof.EndProfileWithMetric([]string{"AddressValidator#Execute"})
	}()

	rc, _ := io.ExecContext.Get(constants.RequestContext)
	io.ExecContext.SetDebugMsg("Address Validator", "Address Validator#Execute")
	p, _ := io.IOData.Get(appconstant.IO_REQUEST_PARAMS)
	params, pOk := p.(*RequestParams)
	if !pOk || params == nil {
		logger.Error("AddressValidator:\tInvalid type of params.")
		return io, &constants.AppError{Code: constants.ParamsInValidErrorCode, Message: "invalid type of params"}
	}
	rp, _ := io.IOData.Get(constants.Request)
	appHTTPReq, _ := rp.(*utilHttp.Request)
	err := validateAddressParams(params, appHTTPReq.HTTPVerb, io)
	if err != nil {
		logger.Error("Address Validator:\tRequest params validation failed." + err.Error())
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: err.Error()}
	}
	logger.Info(fmt.Sprintf("AddressValidator#Execute received the request params: %+v", params), rc)
	return io, nil
}

func validateAddressParams(params *RequestParams, httpVerb utilHttp.Method, io workflow.WorkFlowData) error {
	rp, _ := io.IOData.Get(appconstant.IO_HTTP_REQUEST)
	appHTTPReq, _ := rp.(*utilHttp.Request)
	bodyParam, err := appHTTPReq.GetBodyParameter()
	byteArr := []byte(bodyParam)
	if err != nil {
		logger.Error(fmt.Sprintf("Invalid Body Param: %v", err), params.RequestContext)
		return err
	}

	var r interface{}
	err = json.Unmarshal(byteArr, &r)
	if err != nil {
		logger.Error(fmt.Sprintf("decodeJson error: %v", err), params.RequestContext)
		return err
	}

	valMap, rOk := r.(map[string]interface{})
	if !rOk {
		err = errors.New("couldn't resolve json to map")
		logger.Error(err.Error(), params.RequestContext)
		return err
	}

	address := AddressRequest{}
	for key, value := range valMap {
		switch key {
		case (appconstant.ADDRESS_ID):
			id, ok := value.(float64)
			if !isIntegral(id) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'id' is excpected to be integer type"), params.RequestContext)
				return errors.New("Field Name 'id' is excpected to be integer type")
			}
			address.Id = strconv.FormatInt(int64(id), 10)
		case appconstant.FIRST_NAME:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'firstname' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'firstname' is excpected to be string type")
			}
			address.FirstName = sanitize(str, true)
		case appconstant.LAST_NAME:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'lastname' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'lastname' is excpected to be string type")
			}
			address.LastName = sanitize(str, true)
		case appconstant.ADDRESS1:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'address1' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'address1' is excpected to be string type")
			}
			address.Address1 = sanitize(str, false)
		case appconstant.ADDRESS2:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'address2' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'address2' is excpected to be string type")
			}
			address.Address2 = sanitize(str, false)
		case appconstant.PHONE:
			mobile, ok := value.(float64)
			if !isIntegral(mobile) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'phone' is excpected to be integer type"), params.RequestContext)
				return errors.New("Field Name 'phone' is excpected to be integer type")
			}
			s := fmt.Sprintf("%d", uint64(mobile))
			if len(s) != 10 {
				return errors.New("invalid phone number - length should be 10 digits")
			}

			address.Phone = strconv.FormatInt(int64(mobile), 10)

		case appconstant.ALTERNATE_PHONE:
			altPh, ok := value.(float64)
			if !isIntegral(altPh) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'alt_phone' is excpected to be integer type"), params.RequestContext)
				return errors.New("Field Name 'alt_phone' is excpected to be integer type")
			}
			s := fmt.Sprintf("%d", uint64(altPh))
			if len(s) != 10 {
				return errors.New("invalid alternate phone - length should be 10 digits")
			}
			address.AlternatePhone = strconv.FormatInt(int64(altPh), 10)
		case appconstant.CITY:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'city' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'city' is excpected to be string type")
			}
			address.City = sanitize(str, false)
		case appconstant.REGION:
			address.RegionName = value.(string)
		case appconstant.ADDRESS_REGION:
			addressRegionID, ok := value.(float64)
			if !isIntegral(addressRegionID) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'address_region' is excpected to be integer type"), params.RequestContext)
				return errors.New("Field Name 'address_region' is excpected to be integer type")
			}
			logger.Info(fmt.Sprintf("AddressRegion Id after converting to uint32 is: %d", uint32(addressRegionID)), params.RequestContext)
			address.AddressRegion = strconv.Itoa(int(addressRegionID))
		case appconstant.POSTCODE:
			p, ok := value.(float64)
			if !isIntegral(p) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'postcode' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'postcode' is excpected to be int type")
			}
			pc := fmt.Sprintf("%d", int(p))
			if len(pc) != 6 {
				return errors.New("Invalid postcode")
			}
			address.PostCode = strconv.Itoa(int(p))
		case appconstant.SMS_OPT:
			sms, ok := value.(float64)
			if !isIntegral(sms) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'sms_opt' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'sms_opt' is excpected to be int type")
			}
			str := int(sms)
			if str != 0 && str != 1 {
				logger.Error(fmt.Sprintf("Invalid Value in 'sms_opt' field - should be 0 or 1"), params.RequestContext)
				return errors.New("Invalid Value in 'sms_opt' field - should be 0 or 1")
			}
			address.SmsOpt = strconv.Itoa(str)
		case appconstant.IS_OFFICE:
			address.IsOffice = "0"
			isOffice, ok := value.(float64)
			if !isIntegral(isOffice) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'is_office' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'is_office' is excpected to be int type")
			}
			str := int(isOffice)
			if str != 0 && str != 1 {
				logger.Error(fmt.Sprintf("Invalid Value in 'is_office' field - should be 0 or 1"), params.RequestContext)
				return errors.New("Invalid Value in 'is_office' field - should be 0 or 1")
			}
			address.IsOffice = strconv.Itoa(str)
		case appconstant.COUNTRY:
			country, ok := value.(float64)
			if !isIntegral(country) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'country' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'country' is excpected to be int type")
			}
			address.Country = strconv.Itoa(int(country))
		case appconstant.ADDRESS_TYPE:
			addressType, err := validateAddressType(value)
			if err != nil {
				logger.Error(fmt.Sprintf(err.Error()), params.RequestContext)
				return err
			}
			address.AddressType = addressType
		case appconstant.PARAM_REQ:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'Req' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'Req' is excpected to be string type")
			}
			if str != "" && str != appconstant.UPDATE_TYPE {
				logger.Error(fmt.Sprintf("Incorrect value in 'Req' field - it should either be \"\" or \"update_type\""), params.RequestContext)
				return errors.New("Incorrect value in 'Req' field - it should either be \"\" or \"update_type\"")
			}
			address.Req = str

		default:
			break
		}
	}
	if httpVerb == "PUT" {
		if address.Req != "" {
			if address.Id == "" || address.AddressType == "" {
				return errors.New("Required parameters are missing")
			}
		} else {
			// TODO: Tell what params are missing
			if address.Id == "" ||
				address.FirstName == "" ||
				address.Address1 == "" ||
				address.City == "" ||
				address.PostCode == "" ||
				address.AddressRegion == "" {
				return errors.New("Required parameters are missing")
			}
		}
	} else if httpVerb == "POST" {
		// TODO: Tell what params are missing
		if address.FirstName == "" ||
			address.Address1 == "" ||
			address.City == "" ||
			address.PostCode == "" ||
			address.AddressRegion == "" {
			return errors.New("Required parameters are missing")
		}
	}
	params.QueryParams.Address = address

	return nil
}

func isIntegral(val float64) bool {
	return val == float64(int(val))
}
