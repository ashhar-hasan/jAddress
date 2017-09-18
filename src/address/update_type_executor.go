package address

import (
	"common/appconstant"
	"fmt"

	"github.com/jabong/florest-core/src/common/constants"
	"github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

type UpdateTypeExecutor struct {
	id string
}

func (n *UpdateTypeExecutor) SetID(id string) {
	n.id = id
}

func (n UpdateTypeExecutor) GetID() (string, error) {
	return n.id, nil
}

func (n UpdateTypeExecutor) Name() string {
	return "UpdateTypeExecutor"
}

func (n UpdateTypeExecutor) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("UpdateTypeExecutor")

	defer func() {
		prof.EndProfileWithMetric([]string{"UpdateTypeExecutor_execute"})
	}()
	rc, _ := io.ExecContext.Get(constants.RequestContext)
	logger.Info("UpdateTypeExecutor_rc")
	io.ExecContext.SetDebugMsg("Update Type Executor", "Update Type Executor-Execute")
	p, _ := io.IOData.Get(appconstant.IO_REQUEST_PARAMS)
	params, pOk := p.(*RequestParams)
	if !pOk || params == nil {
		logger.Error("UpdateTypeExecutor. invalid type of params")
		return io, &constants.AppError{Code: constants.ParamsInValidErrorCode, Message: "invalid type of params"}
	}

	debugInfo := new(Debug)
	addressResult := new(AddressResult)
	addressResult, err := UpdateType(params, debugInfo)
	addDebugContents(io, debugInfo)
	if err != nil {
		logger.Error(fmt.Sprintf("There is some error occured while updating the type %v", err), rc)
		return io, &constants.AppError{Code: constants.DbErrorCode, Message: err.Error()}
	}
	err = io.IOData.Set(appconstant.IO_ADDRESS_RESULT, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("error in setting update type  result to workflow data- %v", err), rc)
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: err.Error()}
	}
	return io, nil

}
