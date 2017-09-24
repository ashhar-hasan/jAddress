package servicetest

import (
	"github.com/jabong/florest-core/src/core/service"
)

func initTestConfig() {
	cm := new(service.ConfigManager)
	cm.InitializeGlobalConfig("../../../config/testdata/testconf.json")
}
