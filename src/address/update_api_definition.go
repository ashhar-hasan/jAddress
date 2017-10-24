package address

import (
	"common/appconstant"
	"fmt"

	"github.com/jabong/florest-core/src/common/constants"
	"github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/ratelimiter"
	"github.com/jabong/florest-core/src/core/common/orchestrator"
	"github.com/jabong/florest-core/src/core/common/utils/healthcheck"
	"github.com/jabong/florest-core/src/core/common/versionmanager"
)

type UpdateAddressAPI struct {
}

func (a *UpdateAddressAPI) GetVersion() versionmanager.Version {
	return versionmanager.Version{
		Resource: "ADDRESS",
		Version:  "V1",
		Action:   "PUT",
		BucketID: constants.OrchestratorBucketDefaultValue, //todo - should it be a constant
		Path:     "{" + appconstant.URLPARAM_ADDRESSID + "}",
	}
}

func (a *UpdateAddressAPI) GetOrchestrator() orchestrator.Orchestrator {
	logger.Info("CreateAddress Pipeline Creation begin")

	updateAddressOrchestrator := new(orchestrator.Orchestrator)
	updateAddressWorkflow := new(orchestrator.WorkFlowDefinition)
	updateAddressWorkflow.Create()

	//Creation of the nodes in the workflow definition

	queryTermEnhancer := new(QueryTermEnhancer)
	queryTermEnhancer.SetID("1")
	err := updateAddressWorkflow.AddExecutionNode(queryTermEnhancer)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	requestValidator := new(QueryTermValidator)
	requestValidator.SetID("2")
	err = updateAddressWorkflow.AddExecutionNode(requestValidator)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}
	addressValidator := new(AddressValidator)
	addressValidator.SetID("3")
	err = updateAddressWorkflow.AddExecutionNode(addressValidator)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}
	//Encrypt the mobile no. and alternate phone no.
	addressDataEncryptor := new(DataEncryptor)
	addressDataEncryptor.SetID("4")
	err = updateAddressWorkflow.AddExecutionNode(addressDataEncryptor)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	addAddressExecutor := new(UpdateAddressExecutor)
	addAddressExecutor.SetID("5")
	err = updateAddressWorkflow.AddExecutionNode(addAddressExecutor)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	//Add the connection between the nodes
	err = updateAddressWorkflow.AddConnection(queryTermEnhancer, requestValidator)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}
	err = updateAddressWorkflow.AddConnection(requestValidator, addressValidator)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}
	err = updateAddressWorkflow.AddConnection(addressValidator, addressDataEncryptor)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	err = updateAddressWorkflow.AddConnection(addressDataEncryptor, addAddressExecutor)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}
	//Set start node for the search workflow
	updateAddressWorkflow.SetStartNode(queryTermEnhancer)

	//Assign the workflow definition to the Orchestrator
	updateAddressOrchestrator.Create(updateAddressWorkflow)

	logger.Info(updateAddressOrchestrator.String())
	logger.Info("CreateAddress Pipeline Created")
	return *updateAddressOrchestrator
}

func (a *UpdateAddressAPI) GetHealthCheck() healthcheck.HCInterface {
	return new(AddressHealthCheck)
}

func (a *UpdateAddressAPI) Init() {
	//api initialization should come here
}

func (a *UpdateAddressAPI) GetRateLimiter() ratelimiter.RateLimiter {
	return nil
}
