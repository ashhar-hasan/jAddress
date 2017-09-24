package address

import (
	"common/appconfig"
	"common/appconstant"

	"github.com/jabong/florest-core/src/common/logger"
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

	Initialise()
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
	service.RegisterAPI(new(ListAddressAPI))
	service.RegisterAPI(new(CreateAddressAPI))
	service.RegisterAPI(new(UpdateAddressAPI))
	service.RegisterAPI(new(DeleteAddressAPI))
	service.RegisterAPI(new(UpdateTypeAPI))
}

func initTestConfig() {
	cm := new(service.ConfigManager)
	cm.InitializeGlobalConfig("../../config/testdata/testconf.json")
}

func initTestLogger() {
	logger.Initialise("../../config/testdata/testloggerAsync.json")
}
