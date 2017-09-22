package servicetest

import (
	"address"
	"common/appconstant"

	"github.com/jabong/florest-core/src/core/service"
)

func InitializeTestService() {
	service.RegisterHTTPErrors(appconstant.APPErrorCodeToHTTPCodeMap)
	registerAllAPIs()
	initTestLogger()

	initTestConfig()

	service.InitMonitor()

	service.InitVersionManager()

	initialiseTestWebServer()

	service.InitHealthCheck()

}

func PurgeTestService() {

}

func registerAllAPIs() {
	service.RegisterAPI(new(address.ListAddressAPI))
	service.RegisterAPI(new(address.CreateAddressAPI))
	service.RegisterAPI(new(address.UpdateAddressAPI))
	service.RegisterAPI(new(address.DeleteAddressAPI))
	service.RegisterAPI(new(address.UpdateTypeAPI))
}
