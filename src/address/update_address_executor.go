package address

import (
	"common/appconstant"
	"fmt"

	constants "github.com/jabong/florest-core/src/common/constants"
	logger "github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/profiler"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

type UpdateAddressExecutor struct {
	id string
}

func (n *UpdateAddressExecutor) SetID(id string) {
	n.id = id
}

func (n UpdateAddressExecutor) GetID() (id string, err error) {
	return n.id, nil
}

func (a UpdateAddressExecutor) Name() string {
	return "AddAddressExecutor"
}

func (a UpdateAddressExecutor) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	prof := profiler.NewProfiler()
	prof.StartProfile("UpdateAddressExecutor")

	defer func() {
		prof.EndProfileWithMetric([]string{"UpdateAddressExecutor_execute"})
	}()

	rc, _ := io.ExecContext.Get(constants.RequestContext)
	logger.Info("UpdateAddressExecutor_rc")
	io.ExecContext.SetDebugMsg("Update Address Executor", "Update Address Executor-Execute")
	p, _ := io.IOData.Get(appconstant.IoRequestParams)
	params, pOk := p.(*RequestParams)
	if !pOk || params == nil {
		logger.Error("UpdateAddressExecutor. invalid type of params")
		return io, &constants.AppError{Code: constants.ParamsInValidErrorCode, Message: "invalid type of params"}
	}
	addressResult := new(AddressResult)
	debugInfo := new(Debug)
	var err error

	if params.QueryParams.Address.Id != 0 {
		addressResult, err = UpdateAddress(params, debugInfo)
		addDebugContents(io, debugInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("There is some error occured while updating the address %v", err), rc)
			return io, &constants.AppError{Code: constants.DbErrorCode, Message: err.Error()}
		}
	} else {
		addressResult, err = AddAddress(params, debugInfo)
		addDebugContents(io, debugInfo)
		if err != nil {
			logger.Error(fmt.Sprintf("There is some error occured while creating the new address %v", err), rc)
			return io, &constants.AppError{Code: constants.DbErrorCode, Message: err.Error()}
		}
	}
	derr := io.IOData.Set(appconstant.IoAddressResult, addressResult)
	if derr != nil {
		logger.Error(fmt.Sprintf("error in setting address result to workflow data- %v", derr), rc)
		return io, &constants.AppError{Code: constants.IncorrectDataErrorCode, Message: derr.Error()}
	}
	return io, nil
}
