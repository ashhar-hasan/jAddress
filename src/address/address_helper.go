package address

import (
	"fmt"

	logger "github.com/jabong/florest-core/src/common/logger"
	workflow "github.com/jabong/florest-core/src/core/common/orchestrator"
)

//DebugInfo represents a message
type DebugInfo struct {
	Key   string
	Value string
}

//Debug represents response
type Debug struct {
	MessageStack []DebugInfo
}

//addDebugContents add debug contents in the workflow data
func addDebugContents(io workflow.WorkFlowData, debug *Debug) {
	for k, v := range debug.MessageStack {
		io.ExecContext.SetDebugMsg(v.Key, v.Value)
		logger.Info(fmt.Sprintf("params.DebugInfo %v- %v", k, v))
	}
}
