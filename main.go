package main

import (
	"address"
	"common/appconfig"
	"common/appconstant"
	"fmt"

	"github.com/jabong/florest-core/src/common/utils/http"
	"github.com/jabong/florest-core/src/core/service"
)

//main is the entry point of the florest web service
func main() {
	fmt.Println("APPLICATION BEGIN")
	webserver := new(service.Webserver)
	registerConfig()
	registerErrors()
	registerAllApis()
	registerInitFunc()
	overrideConfByEnvVariables()
	webserver.Start()
}

func registerAllApis() {
	service.RegisterAPI(new(address.ListAddressAPI))
	service.RegisterAPI(new(address.CreateAddressAPI))
	service.RegisterAPI(new(address.UpdateAddressAPI))
	service.RegisterAPI(new(address.DeleteAddressAPI))
	service.RegisterAPI(new(address.UpdateTypeAPI))
}

func registerConfig() {
	service.RegisterConfig(new(appconfig.AddressServiceConfig))
}

func registerErrors() {
	service.RegisterHTTPErrors(appconstant.APPErrorCodeToHTTPCodeMap)
}

func registerInitFunc() {
	service.RegisterCustomAPIInitFunc(func() {
		appHeaderMap := map[http.CustomHeader]string{
			http.UserID:        "X-Jabong-UserId",
			http.SessionID:     "X-Jabong-SessionId",
			http.RequestID:     "X-Jabong-Reqid",
			http.TransactionID: "X-Jabong-Tid",
			http.TokenID:       "X-Jabong-Token",
			http.AppID:         "X-Jabong-Appid",
			http.DebugFlag:     "X-Jabong-Debug",
		}
		http.RegisterCustomHeader(appHeaderMap)
		address.Initialise()
	})
}

func overrideConfByEnvVariables() {
	service.RegisterGlobalEnvUpdateMap(appconfig.MapEnvVariables())
}
