package address

import (
	"fmt"

	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"

	"github.com/jabong/florest-core/src/common/constants"
	"github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	utilHttp "github.com/jabong/florest-core/src/common/utils/http"
)

//QueryTermEnhancer parses, validates and sets default input parameters
type QueryTermEnhancer struct {
	id string
}

func (n *QueryTermEnhancer) SetID(id string) {
	n.id = id
}

func (n QueryTermEnhancer) GetID() (id string, err error) {
	return n.id, nil
}

func (a QueryTermEnhancer) Name() string {
	return "QueryTermEnhancer"
}

func (a QueryTermEnhancer) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	p := profiler.NewProfiler()
	p.StartProfile("Address#QueryTermEnhancer")

	defer func() {
		p.EndProfileWithMetric([]string{"QueryTermEnhancer#Execute"})
	}()

	hrc, _ := io.ExecContext.Get(constants.RequestContext)
	rc, _ := hrc.(utilHttp.RequestContext)
	logger.Info("Entered "+a.Name(), rc)

	io.ExecContext.SetDebugMsg("Query Term Enhancer", "Query Term Enhancer Execute")
	rp, _ := io.IOData.Get(constants.Request)
	logger.Debug(fmt.Sprintf("Request : %v", rp), rc)
	appHttpReq, pOk := rp.(*utilHttp.Request)
	logger.Debug(fmt.Sprintf("HTTP Request : %+v", appHttpReq), rc)

	if !pOk || appHttpReq == nil {
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: "Invalid request params"}
	}
	httpReq := appHttpReq.OriginalRequest
	sessionId, err := io.ExecContext.Get("SESSION_ID")
	if sessionId == "" || err != nil {
		return io, &constants.AppError{Code: constants.ParamsInSufficientErrorCode, Message: "SessionId must be provided in request header"}
	}

	m, _ := io.IOData.Get(constants.ResponseMetaData)
	md, _ := m.(*utilHttp.ResponseMetaData)
	if md == nil {
		md = utilHttp.NewResponseMetaData()
		io.IOData.Set(constants.ResponseMetaData, md)
	}

	//create new request params
	params := RequestParams{}
	resource, _ := io.IOData.Get(constants.Resource)
	logger.Debug(fmt.Sprintf("QueryParams : %+v", params), rc)
	if derr := io.IOData.Set("QUERYPARAMS", &params); derr != nil {
		return io, derr
	}
	return io, nil
}
