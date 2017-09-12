package address

import (
	"fmt"

	"github.com/jabong/florest-core/src/common/constants"
	"github.com/jabong/florest-core/src/common/logger"
	"github.com/jabong/florest-core/src/common/ratelimiter"
	"github.com/jabong/florest-core/src/core/common/orchestrator"
	"github.com/jabong/florest-core/src/core/common/utils/healthcheck"
	"github.com/jabong/florest-core/src/core/common/versionmanager"
)

type CreateAddressAPI struct {
}

func (a *CreateAddressAPI) GetVersion() versionmanager.Version {
	return versionmanager.Version{
		Resource: "ADDRESS",
		Version:  "V1",
		Action:   "POST",
		BucketID: constants.OrchestratorBucketDefaultValue, //todo - should it be a constant
		Path:     "",
	}
}

func (a *CreateAddressAPI) GetOrchestrator() orchestrator.Orchestrator {
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

	addressValidator := new(AddressValidator)
	addressValidator.SetID("2")
	err = updateAddressWorkflow.AddExecutionNode(addressValidator)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}
	//Encrypt the mobile no. and alternate phone no.
	addressDataEncryptor := new(DataEncryptor)
	addressDataEncryptor.SetID("3")
	err = updateAddressWorkflow.AddExecutionNode(addressDataEncryptor)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	addAddressExecutor := new(UpdateAddressExecutor)
	addAddressExecutor.SetID("4")
	err = updateAddressWorkflow.AddExecutionNode(addAddressExecutor)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	//Add the connection between the nodes
	err = updateAddressWorkflow.AddConnection(queryTermEnhancer, addressValidator)
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

func (a *CreateAddressAPI) GetHealthCheck() healthcheck.HCInterface {
	return new(AddressHealthCheck)
}

func (a *CreateAddressAPI) Init() {
	//api initialization should come here
}

func (a *CreateAddressAPI) GetRateLimiter() ratelimiter.RateLimiter {
	return nil
}
