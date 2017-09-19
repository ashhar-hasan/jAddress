package servicetest

import (
	"address"
	"common/appconstant"

	"github.com/jabong/florest-core/src/core/service"
)

func InitializeTestService() {
	service.RegisterHTTPErrors(appconstant.APPErrorCodeToHTTPCodeMap)
	service.RegisterAPI(new(address.ListAllAddressAPI))
	initTestLogger()

	initTestConfig()

	service.InitMonitor()

	service.InitVersionManager()

	initialiseTestWebServer()

	service.InitHealthCheck()

}

func PurgeTestService() {

}
