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
		case appconstant.FIRST_NAME:
			str, ok := value.(string)
			if !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be string type", appconstant.FIRST_NAME)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			address.FirstName = sanitize(str, true)
		case appconstant.LAST_NAME:
			str, ok := value.(string)
			if !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be string type", appconstant.LAST_NAME)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			address.LastName = sanitize(str, true)
		case appconstant.ADDRESS1:
			str, ok := value.(string)
			if !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be string type", appconstant.ADDRESS1)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			address.Address1 = sanitize(str, false)
		case appconstant.ADDRESS2:
			str, ok := value.(string)
			if !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be string type", appconstant.ADDRESS2)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			address.Address2 = sanitize(str, false)
		case appconstant.PHONE:
			mobile, ok := value.(string)
			if !isIntegral(mobile) || !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be integer type", appconstant.PHONE)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			validLen := 10
			if len(mobile) != validLen {
				msg := fmt.Sprintf("Invalid value for field '%s' - length should be %d", appconstant.PHONE, validLen)
				return errors.New(msg)
			}
			address.Phone = mobile
		case appconstant.ALTERNATE_PHONE:
			altPh, ok := value.(string)
			if !isIntegral(altPh) || !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be integer type", appconstant.ALTERNATE_PHONE)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			validLen := 10
			if len(altPh) != validLen {
				msg := fmt.Sprintf("Invalid value for field '%s' - length should be %d", appconstant.ALTERNATE_PHONE, validLen)
				return errors.New(msg)
			}
			address.AlternatePhone = altPh
		case appconstant.CITY:
			str, ok := value.(string)
			if !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be string type", appconstant.CITY)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			address.City = sanitize(str, false)
		case appconstant.REGION:
			str, ok := value.(string)
			if !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be string type", appconstant.REGION)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			address.RegionName = str
		case appconstant.ADDRESS_REGION:
			addressRegionID, ok := value.(string)
			if !isIntegral(addressRegionID) || !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be integer type", appconstant.ADDRESS_REGION)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			address.AddressRegion = addressRegionID
		case appconstant.POSTCODE:
			p, ok := value.(string)
			if !isIntegral(p) || !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be integer type", appconstant.POSTCODE)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			validLen := 6
			if len(p) != validLen {
				msg := fmt.Sprintf("Invalid value for field '%s' - length should be %d", appconstant.POSTCODE, validLen)
				return errors.New(msg)
			}
			address.PostCode = p
		case appconstant.SMS_OPT:
			sms, ok := value.(string)
			if !isIntegral(sms) || !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be integer type", appconstant.SMS_OPT)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			str, _ := strconv.Atoi(sms)
			if str != 0 && str != 1 {
				logger.Error(fmt.Sprintf("Invalid Value in 'sms_opt' field - should be 0 or 1"), params.RequestContext)
				return errors.New("Invalid Value in 'sms_opt' field - should be 0 or 1")
			}
			address.SmsOpt = sms
		case appconstant.IS_OFFICE:
			isOffice, ok := value.(string)
			if !isIntegral(isOffice) || !ok {
				msg := fmt.Sprintf("Field name '%s' is expected to be integer type", appconstant.IS_OFFICE)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
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
				msg := fmt.Sprintf("Field name '%s' is expected to be integer type", appconstant.COUNTRY)
				logger.Error(msg, params.RequestContext)
				return errors.New(msg)
			}
			address.Country = country
		default:
			break
		}
	}
	if httpVerb == "PUT" || httpVerb == "POST" {
		// TODO: Tell what params are missing
		if address.FirstName == "" ||
			address.Address1 == "" ||
			address.City == "" ||
			address.PostCode == "" ||
			address.AddressRegion == "" {
			return fmt.Errorf("Required parameters are missing: %s=%s, %s=%s, %s=%s, %s=%s, %s=%s", appconstant.FIRST_NAME, address.FirstName, appconstant.ADDRESS1, address.Address1, appconstant.CITY, address.City, appconstant.POSTCODE, address.PostCode, appconstant.ADDRESS_REGION, address.AddressRegion)
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
