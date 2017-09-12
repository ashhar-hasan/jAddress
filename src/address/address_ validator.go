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

func (n *AddressValidator) SetID(id string) {
	n.id = id
}

func (n AddressValidator) GetID() (string, error) {
	return n.id, nil
}

func (n AddressValidator) Name() string {
	return "AddressValidator"
}

func (a AddressValidator) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("AddressValidator")

	defer func() {
		prof.EndProfileWithMetric([]string{"AddressValidator_execute"})
	}()

	rc, _ := io.ExecContext.Get(constants.RequestContext)
	logger.Info("AddressValidator_rc")
	io.ExecContext.SetDebugMsg("Address Validator", "Address Validator-Execute")
	p, _ := io.IOData.Get(appconstant.IoRequestParams)
	params, pOk := p.(*RequestParams)
	if !pOk || params == nil {
		logger.Error("AddressValidator. invalid type of params")
		return io, &constants.AppError{Code: constants.ParamsInValidErrorCode, Message: "invalid type of params"}
	}
	rp, _ := io.IOData.Get(constants.Request)
	appHTTPReq, _ := rp.(*utilHttp.Request)
	err := validateAddressParams(params, appHTTPReq.HTTPVerb, io)
	if err != nil {
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: err.Error()}
	}
	logger.Info(fmt.Sprintf("params ----- > %+v", params), rc)
	return io, nil
}

func validateAddressParams(params *RequestParams, httpVerb utilHttp.Method, io workflow.WorkFlowData) error {
	rp, _ := io.IOData.Get(appconstant.IoHttpRequest)
	appHttpReq, _ := rp.(*utilHttp.Request)
	bodyParam, err := appHttpReq.GetBodyParameter()
	byteArr := []byte(bodyParam)
	if err != nil {
		logger.Error(fmt.Sprintf("Invalid Body Param = %v", err), params.RequestContext)
		return err
	}

	var r interface{}
	err = json.Unmarshal(byteArr, &r)
	if err != nil {
		logger.Error(fmt.Sprintf("decodeJson error is = %v", err), params.RequestContext)
		return err
	}

	valMap, rOk := r.(map[string]interface{})
	if !rOk {
		logger.Error(fmt.Sprintf("couldn't resolve json to map"), params.RequestContext)
		err = errors.New("couldn't resolve json to map")
		return err
	}

	address := AddressRequest{}
	for key, value := range valMap {
		switch key {
		case (appconstant.AddressId):
			id, ok := value.(float64)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'id' is excpected to be integer type"), params.RequestContext)
				return errors.New("Field Name 'id' is excpected to be integer type")
			}
			address.Id = uint32(id)
		case appconstant.FirstName:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'FirstName' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'FirstName' is excpected to be string type")
			}
			address.FirstName = sanitize(str, true)
		case appconstant.LastName:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'LastName' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'LastName' is excpected to be string type")
			}
			address.LastName = sanitize(str, true)
		case appconstant.Address1:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'Address1' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'Address1' is excpected to be string type")
			}
			address.Address1 = sanitize(str, false)
		case appconstant.Address2:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'Address2' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'Address2' is excpected to be string type")
			}
			address.Address2 = sanitize(str, false)
		case appconstant.Phone:
			mobile, ok := value.(float64)
			s := fmt.Sprintf("%d", uint64(mobile))
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'Phone' is excpected to be integer type"), params.RequestContext)
				return errors.New("Field Name 'Phone' is excpected to be integer type")
			}
			if len(s) != 10 {
				return errors.New("Invalid phone number")
			}

			address.Phone = int64(mobile) //fmt.Sprintf("%d", uint64(mobile))

		case appconstant.AlternatePhone:
			altPh, ok := value.(float64)
			s := fmt.Sprintf("%d", uint64(altPh))
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'Alternate Phone' is excpected to be integer type"), params.RequestContext)
				return errors.New("Field Name 'Alternate Phone' is excpected to be integer type")
			}
			if len(s) != 10 {
				return errors.New("Invalid alternate phone")
			}
			address.AlternatePhone = int64(altPh)
		case appconstant.City:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'City' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'City' is excpected to be string type")
			}
			address.City = sanitize(str, false)
		case appconstant.Region:
			address.RegionName = value.(string)
		case appconstant.AddressRegion:
			addressRegionId, ok := value.(float64)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'Address_Region' is excpected to be integer type"), params.RequestContext)
				return errors.New("Field Name 'Address_Region' is excpected to be integer type")
			}
			logger.Info(fmt.Sprintf("AddressRegion Id after converting to uint32 is: %d", uint32(addressRegionId)), params.RequestContext)
			address.AddressRegion = uint32(addressRegionId)
		case appconstant.Postcode:
			p, ok := value.(float64)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'Postcode' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'Postcode' is excpected to be int type")
			}
			pc := fmt.Sprintf("%d", int(p))
			postcode, err := strconv.Atoi(pc)
			if err != nil {
				logger.Error(fmt.Sprintf("Field Name 'Postcode' is excpected to be integer type"), params.RequestContext)
				return errors.New("Field Name 'Postcode' is excpected to be integer type")
			}
			if len(pc) != 6 {
				return errors.New("Invalid postcode")
			}
			address.PostCode = postcode
		case appconstant.SmsOpt:
			sms, ok := value.(float64)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'sms_opt' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'sms_opt' is excpected to be int type")
			}
			sms_opt := fmt.Sprintf("%d", int(sms))
			str, err := strconv.Atoi(sms_opt)
			if err != nil {
				logger.Error(fmt.Sprintf("Field Name 'sms_opt' is excpected to be integer type"), params.RequestContext)
				return errors.New("Field Name 'sms_opt' is excpected to be integer type")
			}
			if str != 0 && str != 1 {
				logger.Error(fmt.Sprintf("Invalid Value in 'sms_opt' Field"), params.RequestContext)
				return errors.New("Invalid Value in 'sms_opt' Field")
			}
			address.SmsOpt = str
		case appconstant.IsOffice:
			address.IsOffice = 0
			isOffice, ok := value.(float64)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'is_office' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'is_office' is excpected to be int type")
			}
			isOff := fmt.Sprintf("%d", int(isOffice))
			str, _ := strconv.Atoi(isOff)
			if str != 0 && str != 1 {
				logger.Error(fmt.Sprintf("Invalid Value in 'is_office' Field"), params.RequestContext)
				return errors.New("Invalid Value in 'is_office' Field")
			}
			address.IsOffice = str
		case appconstant.Country:
			country, ok := value.(float64)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'country' is excpected to be int type"), params.RequestContext)
				return errors.New("Field Name 'country' is excpected to be int type")
			}
			address.Country = uint32(country)
			fmt.Println(address.Country)
		case appconstant.AddressType:
			addressType, err := validateAddressType(value)
			if err != nil {
				logger.Error(fmt.Sprintf(err.Error()), params.RequestContext)
				return err
			}
			address.AddressType = addressType
		case appconstant.ParamReq:
			str, ok := value.(string)
			if !ok {
				logger.Error(fmt.Sprintf("Field Name 'Req' is excpected to be string type"), params.RequestContext)
				return errors.New("Field Name 'Req' is excpected to be string type")
			}
			if str != "" && str != appconstant.UpdateType {
				logger.Error(fmt.Sprintf("Incorrect value in 'Req' field"), params.RequestContext)
				return errors.New("Incorrect value in 'Req' field")
			}
			address.Req = str

		default:
			break
		}
	}
	if httpVerb == "PUT" {
		if address.Req != "" {
			if address.Id == 0 || address.AddressType == "" {
				return errors.New("Required parameters are missing")
			}
		} else {
			// TODO: Tell what params are missing
			if address.Id == 0 ||
				address.FirstName == "" ||
				address.Address1 == "" ||
				address.City == "" ||
				address.PostCode == 0 ||
				address.AddressRegion == 0 {
				return errors.New("Required parameters are missing")
			}
		}
	} else if httpVerb == "POST" {
		// TODO: Tell what params are missing
		if address.FirstName == "" ||
			address.Address1 == "" ||
			address.City == "" ||
			address.PostCode == 0 ||
			address.AddressRegion == 0 {
			return errors.New("Required parameters are missing")
		}
	}
	params.QueryParams.Address = address

	return nil
}
