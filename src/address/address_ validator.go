package address

import (
	"common/appconstant"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

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
			mobile, ok := value.(string)
			if !isIntegral(strings.TrimPrefix(mobile, "+")) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'phone' is excpected to be of the form +digits"), params.RequestContext)
				return errors.New("Field Name 'phone' is excpected to be of the form +digits")
			}
			if len(mobile) != 13 {
				return errors.New("invalid phone number - length should be 10 digits excluding country code")
			}
			address.Phone = mobile
		case appconstant.ALTERNATE_PHONE:
			altPh, ok := value.(string)
			if !isIntegral(strings.TrimPrefix(altPh, "+")) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'alt_phone' is excpected to be to be of the form +digits"), params.RequestContext)
				return errors.New("Field Name 'alt_phone' is excpected to be to be of the form +digits")
			}
			if len(altPh) != 13 {
				return errors.New("invalid alternate phone - length should be 10 digits excluding country code")
			}
			address.AlternatePhone = altPh
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
			addressRegionID, ok := value.(string)
			if !isIntegral(addressRegionID) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'address_region' is excpected to be integer type"), params.RequestContext)
				return errors.New("Field Name 'address_region' is excpected to be integer type")
			}
			address.AddressRegion = addressRegionID
		case appconstant.POSTCODE:
			p, ok := value.(string)
			if !isIntegral(p) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'postcode' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'postcode' is excpected to be int type")
			}
			if len(p) != 6 {
				return errors.New("Invalid postcode")
			}
			address.PostCode = p
		case appconstant.SMS_OPT:
			sms, ok := value.(string)
			if !isIntegral(sms) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'sms_opt' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'sms_opt' is excpected to be int type")
			}
			str, _ := strconv.Atoi(sms)
			if str != 0 && str != 1 {
				logger.Error(fmt.Sprintf("Invalid Value in 'sms_opt' field - should be 0 or 1"), params.RequestContext)
				return errors.New("Invalid Value in 'sms_opt' field - should be 0 or 1")
			}
			address.SmsOpt = sms
		case appconstant.IS_OFFICE:
			address.IsOffice = "0"
			isOffice, ok := value.(string)
			if !isIntegral(isOffice) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'is_office' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'is_office' is excpected to be int type")
			}
			str, _ := strconv.Atoi(isOffice)
			if str != 0 && str != 1 {
				logger.Error(fmt.Sprintf("Invalid Value in 'is_office' field - should be 0 or 1"), params.RequestContext)
				return errors.New("Invalid Value in 'is_office' field - should be 0 or 1")
			}
			address.IsOffice = isOffice
		case appconstant.COUNTRY:
			country, ok := value.(string)
			if !isIntegral(country) || !ok {
				logger.Error(fmt.Sprintf("Field Name 'country' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'country' is excpected to be int type")
			}
			address.Country = country
		default:
			break
		}
	}
	if httpVerb == "PUT" {
		// TODO: Tell what params are missing
		if address.FirstName == "" ||
			address.Address1 == "" ||
			address.City == "" ||
			address.PostCode == "" ||
			address.AddressRegion == "" {
			return errors.New("Required parameters are missing")
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

func isIntegral(val string) bool {
	fl, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return false
	}
	return fl == float64(int(fl))
}
