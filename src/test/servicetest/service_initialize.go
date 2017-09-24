package servicetest

import (
	"address"
	"common/appconfig"
	"common/appconstant"

	utilHttp "github.com/jabong/florest-core/src/common/utils/http"
	"github.com/jabong/florest-core/src/core/service"
)

func InitializeTestService() {
	registerCustomHTTPHeaders()
	service.RegisterHTTPErrors(appconstant.APPErrorCodeToHTTPCodeMap)
	service.RegisterConfig(new(appconfig.AddressServiceConfig))
	registerAllAPIs()
	initTestLogger()
	initTestConfig()
	service.InitMonitor()
	service.InitVersionManager()
	initialiseTestWebServer()
	service.InitHealthCheck()

	address.Initialise()
}

func PurgeTestService() {

}

func registerCustomHTTPHeaders() {
	appHeaderMap := map[utilHttp.CustomHeader]string{
		utilHttp.UserID:        "X-Jabong-UserId",
		utilHttp.SessionID:     "X-Jabong-SessionId",
		utilHttp.RequestID:     "X-Jabong-Reqid",
		utilHttp.TransactionID: "X-Jabong-Tid",
		utilHttp.TokenID:       "X-Jabong-Token",
		utilHttp.AppID:         "X-Jabong-Appid",
		utilHttp.DebugFlag:     "X-Jabong-Debug",
	}
	utilHttp.RegisterCustomHeader(appHeaderMap)
}

func registerAllAPIs() {
	service.RegisterAPI(new(address.ListAddressAPI))
	service.RegisterAPI(new(address.CreateAddressAPI))
	service.RegisterAPI(new(address.UpdateAddressAPI))
	service.RegisterAPI(new(address.DeleteAddressAPI))
	service.RegisterAPI(new(address.UpdateTypeAPI))
}
