package address

import (
	errors "common/appconstant"

	florest_constants "github.com/jabong/florest-core/src/common/constants"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

type Address struct {
	id string
}

func (n *Address) SetID(id string) {
	n.id = id
}

func (n Address) GetID() (id string, err error) {
	return n.id, nil
}

func (a Address) Name() string {
	return "Address"
}

func (a Address) Execute(io workflow.WorkFlowData) (workflow.WorkFlowData, error) {
	//Business Logic
	return io, &florest_constants.AppError{Code: errors.FunctionalityNotImplementedErrorCode, Message: "invalid request"}
}
