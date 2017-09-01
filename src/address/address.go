package address

import (
	errors "common/appconstant"

	florest_constants "github.com/jabong/florest-core/src/common/constants"
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
	return io, &florest_constants.AppError{Code: errors.FunctionalityNotImplementedErrorCode, Message: "invalid request"}
}
