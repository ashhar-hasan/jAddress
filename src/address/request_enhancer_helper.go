package address

import (
	"common/appconstant"
	"fmt"
	"regexp"
	"strings"

	"github.com/jabong/florest-core/src/common/constants"
	"github.com/jabong/florest-core/src/common/logger"
	utilHttp "github.com/jabong/florest-core/src/common/utils/http"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

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

func validateSession(sessionID *string) bool {
	session := strings.Trim(*sessionID, " ")
	ret := true
	if len(session) <= 0 {
		ret = false
	} else {
		r, _ := regexp.Compile("^[A-Za-z0-9-]{20,}$")
		if r.MatchString(session) == false {
			ret = false
		}
	}
	return ret
}

func updateDefaultInParams(params *RequestParams, appHTTPReq *utilHttp.Request) {
	var err error
	params.QueryParams.Default, err = utilHttp.GetIntParamFields(appHTTPReq.OriginalRequest, appconstant.URLPARAM_DEFAULT)
	if err != nil {
		params.QueryParams.Default = 0
	}
	if params.QueryParams.Default == 1 {
		params.QueryParams.AddressType = appconstant.SHIPPING
	}
}
