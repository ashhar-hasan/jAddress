package address

import (
	"common/appconfig"
	"fmt"

	"github.com/jabong/florest-core/src/common/config"
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
	addressWorkflow := new(orchestrator.WorkFlowDefinition)
	addressWorkflow.Create()

	//Creation of the nodes in the workflow definition
	addressNode := new(Address)
	addressNode.SetID("address node 1")
	eerr := addressWorkflow.AddExecutionNode(addressNode)
	if eerr != nil {
		logger.Error(fmt.Sprintln(eerr))
	}

	//Set start node for the search workflow
	addressWorkflow.SetStartNode(addressNode)

	//Assign the workflow definition to the Orchestrator
	addressOrchestrator.Create(addressWorkflow)

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
	c := config.GlobalAppConfig.ApplicationConfig
	appConfig, _ := c.(*appconfig.AddressServiceConfig)
	fmt.Println(appConfig.MySqlConfig.MySqlMaster.Username)
}

func (a *AddressAPI) GetRateLimiter() ratelimiter.RateLimiter {
	return nil
}
