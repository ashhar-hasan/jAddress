package address

import (
	"common/appconstant"
	"fmt"

	"github.com/jabong/florest-core/src/common/constants"
	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

//ListAddressExecutor is responsible for retreiving the list of addresses
//associated with a particular user
type ListAddressExecutor struct {
	id string
}

func (a *ListAddressExecutor) SetID(id string) {
	a.id = id
}

func (a ListAddressExecutor) GetID() (id string, err error) {
	return a.id, nil
}

func (a ListAddressExecutor) Name() string {
	return "ListAddressExecutor"
}

//Execute sets the retreived addresses of a user into the workflow data
func (a ListAddressExecutor) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	profiler := profiler.NewProfiler()
	profiler.StartProfile("Address#ListAddressExecutor")

	defer func() {
		profiler.EndProfileWithMetric([]string{"ListAddressExecutor#Execute"})
	}()

	rc, _ := io.ExecContext.Get(constants.RequestContext)
	logger.Info("Entered "+a.Name(), rc)
	io.ExecContext.SetDebugMsg("List Address Executor", "List Address Executor Execute")

	p, _ := io.IOData.Get(appconstant.IO_REQUEST_PARAMS)
	params, pOk := p.(*RequestParams)
	if !pOk || params == nil {
		logger.Error("ListAddressExecutor.invalid type of params", rc)
		return io, &constants.AppError{Code: constants.ParamsInValidErrorCode, Message: "Invalid type of params"}
	}

	debugInfo := new(Debug)
	addressListResult, err := GetAddressList(params, debugInfo)
	if err != nil {
		logger.Error(fmt.Println("unable to get address"))
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: err.Error()}
	}
	addDebugContents(io, debugInfo)
	derr := io.IOData.Set(appconstant.IO_ADDRESS_RESULT, addressListResult)
	if derr != nil {
		logger.Error(fmt.Sprintf("Error in setting address list result to workflow data- %v", derr), rc)
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: derr.Error()}
	}

	return io, nil
}
