package main

import (
	"common/appconfig"
	"common/appconstant"
	"fmt"
	"github.com/jabong/florest-core/src/common/utils/http"
	"github.com/jabong/florest-core/src/core/service"
	"hello"
)

//main is the entry point of the florest web service
func main() {
	fmt.Println("APPLICATION BEGIN")
	webserver := new(service.Webserver)
	registerConfig()
	registerErrors()
	registerAllApis()
	registerInitFunc()
	webserver.Start()
}

func registerAllApis() {
	service.RegisterAPI(new(hello.HelloAPI))
}

func registerConfig() {
	service.RegisterConfig(new(appconfig.AppConfig))
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
	})
}
