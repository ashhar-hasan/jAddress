package address

import (
	"common/appconstant"
	"errors"
	"fmt"
	"net/http"

	constants "github.com/jabong/florest-core/src/common/constants"
	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	utilHttp "github.com/jabong/florest-core/src/common/utils/http"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

type QueryTermValidator struct {
	id string
}

func (a *QueryTermValidator) SetID(id string) {
	a.id = id
}

func (a QueryTermValidator) GetID() (string, error) {
	return a.id, nil
}

func (a QueryTermValidator) Name() string {
	return "QueryTermValidator"
}

func (a QueryTermValidator) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("QueryTermValidator")

	defer func() {
		prof.EndProfileWithMetric([]string{"QueryTermValidator_execute"})
	}()

	rc, _ := io.ExecContext.Get(constants.RequestContext)
	logger.Info("QueryTermValidator_rc")
	io.ExecContext.SetDebugMsg("Query Term Validator", "Query Term Validator-Execute")
	p, _ := io.IOData.Get(appconstant.IoRequestParams)
	params, pOk := p.(*RequestParams)
	if !pOk || params == nil {
		logger.Error("QueryTermValidator. invalid type of params")
		return io, &constants.AppError{Code: constants.ParamsInValidErrorCode, Message: "invalid type of params"}
	}
	rp, _ := io.IOData.Get(constants.Request)
	appHTTPReq, _ := rp.(*utilHttp.Request)
	httpReq := appHTTPReq.OriginalRequest
	err := validateAndSetURLParams(params, httpReq)
	if err != nil {
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: err.Error()}
	}
	logger.Info(fmt.Sprintf("params ----- > %+v", params), rc)
	return io, nil
}

func validateAndSetURLParams(params *RequestParams, httpReq *http.Request) error {
	var (
		limit  = appconstant.DefaultLimit
		offset = appconstant.DefaultOffset
		err    error
	)
	if httpReq.FormValue("limit") != "" {
		limit, err = utilHttp.GetIntParamFields(httpReq, appconstant.UrlParamLimit)
		if err != nil {
			return errors.New("Limit must be a valid number")
		}
	}
	if limit > appconstant.MaxLimit {
		limit = appconstant.DefaultLimit
	}
	params.QueryParams.Limit = limit
	if httpReq.FormValue(appconstant.UrlParamOffset) != "" {
		offset, err = utilHttp.GetIntParamFields(httpReq, appconstant.UrlParamOffset)
		if err != nil {
			return errors.New("Offset must be a number")
		}
	}
	params.QueryParams.Offset = offset
	if httpReq.FormValue(appconstant.UrlParamAddressType) != "" {
		addressType := utilHttp.GetStringParamFields(httpReq, appconstant.UrlParamAddressType)
		res, err := validateAddressType(addressType)
		if err != nil {
			logger.Error(fmt.Sprintf("Invalid address type. Possible types are all, billiing, shipping, other"), params.RequestContext)
			return err
		}
		params.QueryParams.AddressType = res
	}
	return nil
}

func validateAddressType(ty interface{}) (addressType string, err error) {
	str, ok := ty.(string)
	if !ok {
		return addressType, errors.New("Field Name 'addressType' is expected to be string")
	}
	if str == appconstant.Billing {
		addressType = appconstant.Billing
	} else if str == appconstant.Shipping {
		addressType = appconstant.Shipping
	} else if str == appconstant.Other {
		addressType = appconstant.Other
	} else if str == appconstant.All {
		addressType = appconstant.All
	} else {
		return addressType, errors.New("Invalid address type. Possible types are billing, shipping, other, all")
	}
	return addressType, err
}
