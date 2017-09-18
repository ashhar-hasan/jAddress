package address

import (
	"common/appconstant"
	"fmt"

	"github.com/jabong/florest-core/src/common/constants"
	"github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

type DeleteAddressExecutor struct {
	id string
}

func (n *DeleteAddressExecutor) SetID(id string) {
	n.id = id
}

func (n DeleteAddressExecutor) GetID() (id string, err error) {
	return n.id, nil
}

func (a DeleteAddressExecutor) Name() string {
	return "DeleteAddressExecutor"
}

func (a DeleteAddressExecutor) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("DeleteAddressExecutor")

	defer func() {
		prof.EndProfileWithMetric([]string{"DeleteAddressExecutor_execute"})
	}()

	rc, _ := io.ExecContext.Get(constants.RequestContext)
	logger.Info("DeleteAddressExecutor_rc")
	io.ExecContext.SetDebugMsg("Delete Address Executor", "Delete Address Executor-Execute")
	p, _ := io.IOData.Get(appconstant.IO_REQUEST_PARAMS)
	params, pOk := p.(*RequestParams)
	if !pOk || params == nil {
		logger.Error("DeleteAddressExecutor. invalid type of params")
		return io, &constants.AppError{Code: constants.ParamsInValidErrorCode, Message: "invalid type of params"}
	}

	debugInfo := new(Debug)
	_, err := DeleteAddress(params, debugInfo)
	addDebugContents(io, debugInfo)
	if err != nil {
		logger.Error(fmt.Sprintf("There is some error occured while deleting the address %v", err), rc)
		return io, &constants.AppError{Code: constants.DbErrorCode, Message: err.Error()}
	}
	err = io.IOData.Set(appconstant.IO_ADDRESS_RESULT, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("error in setting add address result to workflow data- %v", err), rc)
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: err.Error()}
	}
	return io, nil

}
