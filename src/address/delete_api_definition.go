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

type DeleteAddressAPI struct {
}

func (a *DeleteAddressAPI) GetVersion() versionmanager.Version {
	return versionmanager.Version{
		Resource: "ADDRESS",
		Version:  "V1",
		Action:   "DELETE",
		BucketID: constants.OrchestratorBucketDefaultValue, //todo - should it be a constant
		Path:     "",
	}
}

func (a *DeleteAddressAPI) GetOrchestrator() orchestrator.Orchestrator {
	logger.Info("Delete Address Pipeline Creation begin")

	deleteAddressOrchestrator := new(orchestrator.Orchestrator)
	deleteAddressWorkflow := new(orchestrator.WorkFlowDefinition)
	deleteAddressWorkflow.Create()

	//Creation of the nodes in the search workflow definition
	queryTermEnhancer := new(QueryTermEnhancer)
	queryTermEnhancer.SetID("1")
	err := deleteAddressWorkflow.AddExecutionNode(queryTermEnhancer)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	deleteAddressExecutor := new(DeleteAddressExecutor)
	deleteAddressExecutor.SetID("2")
	err = deleteAddressWorkflow.AddExecutionNode(deleteAddressExecutor)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	//Add the connections between the nodes
	err = deleteAddressWorkflow.AddConnection(queryTermEnhancer, deleteAddressExecutor)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	//Set start node for the suggest workflow
	deleteAddressWorkflow.SetStartNode(queryTermEnhancer)

	//Assign the workflow definition to the Orchestrator
	deleteAddressOrchestrator.Create(deleteAddressWorkflow)

	logger.Info(deleteAddressOrchestrator.String())
	logger.Info("Delete Address Pipeline Created")

	return *deleteAddressOrchestrator
}

func (a *DeleteAddressAPI) GetHealthCheck() healthcheck.HCInterface {
	return new(AddressHealthCheck)
}

func (a *DeleteAddressAPI) Init() {
	//api initialization should come here
}

func (a *DeleteAddressAPI) GetRateLimiter() ratelimiter.RateLimiter {
	return nil
}
