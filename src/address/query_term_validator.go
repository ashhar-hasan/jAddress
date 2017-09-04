package address

import (
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

func (n *QueryTermValidator) SetID(id string) {
	n.id = id
}

func (n QueryTermValidator) GetID() (string, error) {
	return n.id, nil
}

func (n QueryTermValidator) Name() string {
	return "QueryTermValidator"
}

func (a QueryTermValidator) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	profiler := profiler.NewProfiler()
	profiler.StartProfile("QueryTermValidator")

	defer func() {
		profiler.EndProfileWithMetric([]string{"QueryTermValidator_execute"})
	}()

	rc, _ := io.ExecContext.Get(constants.RequestContext)
	logger.Info("QueryTermValidator_rc")
	io.ExecContext.SetDebugMsg("Query Term Validator", "Query Term Validator-Execute")
	p, _ := io.IOData.Get("QUERYPARAMS")
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
		limit  int   = DEFAULT_LIMIT
		offset int   = DEFAULT_OFFSET
		err    error = nil
	)
	if httpReq.FormValue("limit") != "" {
		limit, err = utilHttp.GetIntParamFields(httpReq, "limit")
		if err != nil {
			return errors.New("Limit must be a valid number")
		}
	}
	if limit > MAX_LIMIT {
		limit = DEFAULT_LIMIT
	}
	params.QueryParams.Limit = limit
	if httpReq.FormValue(URL_PARAM_OFFSET) != "" {
		offset, err = utilHttp.GetIntParamFields(httpReq, URL_PARAM_OFFSET)
		if err != nil {
			return errors.New("Offset must be a number")
		}
	}
	params.QueryParams.Offset = offset
	if httpReq.FormValue(URL_PARAM_ADDRESS_TYPE) != "" {
		addressType := utilHttp.GetStringParamFields(httpReq, URL_PARAM_ADDRESS_TYPE)
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
	if str == BILLING {
		addressType = BILLING
	} else if str == SHIPPING {
		addressType = SHIPPING
	} else if str == OTHER {
		addressType = OTHER
	} else if str == ALL {
		addressType = ALL
	} else {
		return addressType, errors.New("Invalid address type. Possible types are billing, shipping")
	}
	return addressType, err
}
