package address

import (
	"common/appconstant"
	"errors"
	"fmt"
	"strconv"

	"github.com/jabong/florest-core/src/common/constants"
	"github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	utilHttp "github.com/jabong/florest-core/src/common/utils/http"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

//QueryTermEnhancer parses, validates and sets default input parameters
type QueryTermEnhancer struct {
	id string
}

func (a *QueryTermEnhancer) SetID(id string) {
	a.id = id
}

func (a QueryTermEnhancer) GetID() (id string, err error) {
	return a.id, nil
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
	appHTTPReq, pOk := rp.(*utilHttp.Request)
	logger.Debug(fmt.Sprintf("HTTP Request : %+v", appHTTPReq), rc)

	if !pOk || appHTTPReq == nil {
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: "Invalid request params"}
	}
	sessionID, err := io.ExecContext.Get(appconstant.SESSION_ID)
	if sessionID == "" || err != nil {
		return io, &constants.AppError{Code: constants.ParamsInSufficientErrorCode, Message: "SessionId must be provided in request header"}
	}
	userID, rerr := io.ExecContext.Get(appconstant.USER_ID)
	if userID == "" || rerr != nil {
		return io, &constants.AppError{Code: constants.ParamsInSufficientErrorCode, Message: "UserId must be provided in request header"}
	}
	m, _ := io.IOData.Get(constants.ResponseMetaData)
	md, _ := m.(*utilHttp.ResponseMetaData)
	if md == nil {
		md = utilHttp.NewResponseMetaData()
		io.IOData.Set(constants.ResponseMetaData, md)
	}

	//create new request params
	params := RequestParams{}

	updateParamsWithBuckets(&params, io)
	updateParamsWithRequestContext(&params, io)
	params.QueryParams.Default, err = utilHttp.GetIntParamFields(appHTTPReq.OriginalRequest, appconstant.URLPARAM_DEFAULT)
	if err != nil {
		params.QueryParams.Default = 0
	}
	if params.QueryParams.Default == 1 {
		params.QueryParams.AddressType = appconstant.SHIPPING
	}

	if appHTTPReq.HTTPVerb == "DELETE" || appHTTPReq.HTTPVerb == "PUT" || appHTTPReq.HTTPVerb == "GET" {
		// Update default billing/shipping address case
		if len(*appHTTPReq.PathParameters) == 2 {
			validateAndSetParamsForUpdate(&params, appHTTPReq)
		}
		derr1 := validateAndSetParams(&params, appHTTPReq)
		if derr1 != nil {
			return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: derr1.Error()}
		}
	}

	logger.Debug(fmt.Sprintf("QueryParams : %+v", params), rc)
	if derr := io.IOData.Set(appconstant.IO_REQUEST_PARAMS, &params); derr != nil {
		return io, derr
	}
	return io, nil
}

//updateParamsWithBuckets updates buckets to params
func updateParamsWithBuckets(params *RequestParams, io workflow.WorkFlowData) {
	rc, _ := io.ExecContext.Get(constants.RequestContext)
	bucketMap, err := io.ExecContext.GetBuckets()
	if err != nil { //no need to return error as its not fatal issue
		logger.Warning(fmt.Sprintf("err in retrieving buckets : %v", err), rc)
	}
	params.Buckets = bucketMap
}

//updateParamsWithRequestContext updates request context to params
func updateParamsWithRequestContext(params *RequestParams, io workflow.WorkFlowData) {
	rc, err := io.ExecContext.Get(constants.RequestContext)
	if err != nil { //no need to return error as its not fatal issue
		logger.Info(fmt.Sprintf("err in retrieving request context : %v", err), rc)
	}
	if v, ok := rc.(utilHttp.RequestContext); ok {
		params.RequestContext = v
	}
}

func validateAndSetParams(params *RequestParams, httpReq *utilHttp.Request) error {
	if httpReq.HTTPVerb == "GET" {
		val := httpReq.GetPathParameter(appconstant.URLPARAM_ADDRESSTYPE)
		if val == "" {
			val = appconstant.ALL
		}
		params.QueryParams.AddressType = val
		return nil
	}
	val := httpReq.GetPathParameter(appconstant.URLPARAM_ADDRESSID)
	addressID, err := strconv.Atoi(val)
	if err != nil {
		return errors.New("Id is missing or not a number")
	}
	params.QueryParams.AddressId = addressID
	return nil
}

func validateAndSetParamsForUpdate(params *RequestParams, httpReq *utilHttp.Request) error {
	val := httpReq.GetPathParameter(appconstant.URLPARAM_ADDRESSID)
	addressID, err := strconv.Atoi(val)
	if err != nil {
		return errors.New("Id is missing or not a number")
	}
	params.QueryParams.AddressId = addressID
	val = httpReq.GetPathParameter(appconstant.URLPARAM_ADDRESSTYPE)
	if val == "" {
		return errors.New("Address Type is missing")
	}
	addressType, _ := validateAddressType(val)
	if addressType == appconstant.ALL || addressType == appconstant.OTHER {
		return errors.New("Address Type can be only be billing or shipping")
	}
	params.QueryParams.AddressType = addressType
	return nil
}
