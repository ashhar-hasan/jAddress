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

type AddressAPI struct {
}

func (a *AddressAPI) GetVersion() versionmanager.Version {
	return versionmanager.Version{
		Resource: "ADDRESS",
		Version:  "V1",
		Action:   "GET",
		BucketID: constants.OrchestratorBucketDefaultValue, //todo - should it be a constant
		Path:     "",
	}
}

func (a *AddressAPI) GetOrchestrator() orchestrator.Orchestrator {
	logger.Info("Address Pipeline Creation begin")

	addressOrchestrator := new(orchestrator.Orchestrator)
	listAddressWorkflow := new(orchestrator.WorkFlowDefinition)
	listAddressWorkflow.Create()

	//Creation of the nodes in the workflow definition

	queryTermEnhancer := new(QueryTermEnhancer)
	queryTermEnhancer.SetID("1")
	err := listAddressWorkflow.AddExecutionNode(queryTermEnhancer)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
		err = nil
	}

	queryTermValidator := new(QueryTermValidator)
	queryTermValidator.SetID("2")
	err = listAddressWorkflow.AddExecutionNode(queryTermValidator)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
		err = nil
	}
	listAddressExecutor := new(ListAddressExecutor)
	listAddressExecutor.SetID("3")
	err = listAddressWorkflow.AddExecutionNode(listAddressExecutor)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
		err = nil
	}

	//Add the connection between the nodes
	err = listAddressWorkflow.AddConnection(queryTermEnhancer, queryTermValidator)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
		err = nil
	}

	err = listAddressWorkflow.AddConnection(queryTermValidator, listAddressExecutor)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
		err = nil
	}

	//Set start node for the search workflow
	listAddressWorkflow.SetStartNode(queryTermEnhancer)

	//Assign the workflow definition to the Orchestrator
	addressOrchestrator.Create(listAddressWorkflow)

	logger.Info(addressOrchestrator.String())
	logger.Info("Address Pipeline Created")
	logger.Info("Address Pipeline Created")
	return *addressOrchestrator
}

func (a *AddressAPI) GetHealthCheck() healthcheck.HCInterface {
	return new(AddressHealthCheck)
}

func (a *AddressAPI) Init() {
	//api initialization should come here
}

func (a *AddressAPI) GetRateLimiter() ratelimiter.RateLimiter {
	return nil
}
