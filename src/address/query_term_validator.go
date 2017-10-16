package address

import (
	"common/appconstant"
	"fmt"

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
	p, _ := io.IOData.Get(appconstant.IO_REQUEST_PARAMS)
	params, pOk := p.(*RequestParams)
	if !pOk || params == nil {
		logger.Error("QueryTermValidator. invalid type of params")
		return io, &constants.AppError{Code: constants.ParamsInValidErrorCode, Message: "invalid type of params"}
	}
	rp, _ := io.IOData.Get(constants.Request)
	appHTTPReq, _ := rp.(*utilHttp.Request)
	httpReq := appHTTPReq.OriginalRequest
	if appHTTPReq.HTTPVerb == "DELETE" || appHTTPReq.HTTPVerb == "PUT" || appHTTPReq.HTTPVerb == "GET" {
		// Update default billing/shipping address case
		if len(*appHTTPReq.PathParameters) == 2 {
			validateAndSetParamsForUpdate(params, appHTTPReq)
		}
		derr1 := validateAndSetParams(params, appHTTPReq)
		if derr1 != nil {
			return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: derr1.Error()}
		}
	}
	err := validateAndSetURLParams(params, httpReq)
	if err != nil {
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: err.Error()}
	}
	logger.Info(fmt.Sprintf("params ----- > %+v", params), rc)
	return io, nil
}
