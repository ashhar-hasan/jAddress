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

type UpdateTypeAPI struct {
}

func (a *UpdateTypeAPI) GetVersion() versionmanager.Version {
	return versionmanager.Version{
		Resource: "UPDATETYPE",
		Version:  "V1",
		Action:   "PUT",
		BucketID: constants.OrchestratorBucketDefaultValue,
		Path:     "",
	}
}

func (a *UpdateTypeAPI) GetOrchestrator() orchestrator.Orchestrator {
	logger.Info("Update Type Pipeline Creation Begin")

	updateTypeOrchestrator := new(orchestrator.Orchestrator)
	updateTypeWorkflow := new(orchestrator.WorkFlowDefinition)
	updateTypeWorkflow.Create()

	queryTermEnhancer := new(QueryTermEnhancer)
	queryTermEnhancer.SetID("1")
	err := updateTypeWorkflow.AddExecutionNode(queryTermEnhancer)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	queryTermValidator := new(QueryTermValidator)
	queryTermValidator.SetID("2")
	err = updateTypeWorkflow.AddExecutionNode(queryTermValidator)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	updateTypeExecutor := new(UpdateTypeExecutor)
	updateTypeExecutor.SetID("3")
	err = updateTypeWorkflow.AddExecutionNode(updateTypeExecutor)
	if err != nil {
		logger.Error(fmt.Sprintln(err))
	}

	updateTypeWorkflow.SetStartNode(queryTermEnhancer)
	updateTypeOrchestrator.Create(updateTypeWorkflow)
	logger.Info(updateTypeOrchestrator.String())
	logger.Info("Update Type Pipeline Created")
	return *updateTypeOrchestrator
}

func (a *UpdateTypeAPI) GetHealthCheck() healthcheck.HCInterface {
	return new(AddressHealthCheck)
}

func (a *UpdateTypeAPI) Init() {
	//api initialization should come here
}

func (a *UpdateTypeAPI) GetRateLimiter() ratelimiter.RateLimiter {
	return nil
}
