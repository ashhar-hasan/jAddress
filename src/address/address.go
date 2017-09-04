package address

import (
	"common/appconstant"
	"fmt"

	"github.com/jabong/florest-core/src/common/constants"
	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

type ListAddressExecutor struct {
	id string
}

func (n *ListAddressExecutor) SetID(id string) {
	n.id = id
}

func (n ListAddressExecutor) GetID() (id string, err error) {
	return n.id, nil
}

func (a ListAddressExecutor) Name() string {
	return "ListAddressExecutor"
}

func (a ListAddressExecutor) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	//Business Logic
	profiler := profiler.NewProfiler()
	profiler.StartProfile("Address#ListAddressExecutor")

	defer func() {
		profiler.EndProfileWithMetric([]string{"ListAddressExecutor#Execute"})
	}()

	rc, _ := io.ExecContext.Get(constants.RequestContext)
	logger.Info("Entered "+a.Name(), rc)
	io.ExecContext.SetDebugMsg("List Address Executor", "List Address Executor Execute")

	p, _ := io.IOData.Get(appconstant.IoRequestParams)
	params, pOk := p.(*RequestParams)
	if !pOk || params == nil {
		logger.Error("ListAddressExecutor.invalid type of params", rc)
		return io, &constants.AppError{Code: constants.ParamsInValidErrorCode, Message: "Invalid type of params"}
	}

	debugInfo := new(Debug)
	addressListResult, _ := GetAddressList(params, debugInfo)
	addDebugContents(io, debugInfo)
	derr := io.IOData.Set(appconstant.IoAddressResult, addressListResult)
	if derr != nil {
		logger.Error(fmt.Sprintf("Error in setting address list result to workflow data- %v", derr), rc)
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: derr.Error()}
	}

	return io, nil
}
