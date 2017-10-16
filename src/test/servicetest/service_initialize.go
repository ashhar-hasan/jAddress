package servicetest

import (
	"common/appconstant"

	"github.com/jabong/florest-core/src/core/service"
)

func InitializeTestService() {
	service.RegisterHTTPErrors(appconstant.APPErrorCodeToHTTPCodeMap)

	initTestLogger()

	initTestConfig()

	service.InitMonitor()

	service.InitVersionManager()

	initialiseTestWebServer()

	service.InitHealthCheck()

}

func PurgeTestService() {

}
