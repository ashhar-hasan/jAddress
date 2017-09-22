package servicetest

import (
	gk "github.com/onsi/ginkgo"
	gm "github.com/onsi/gomega"

	testUtil "test/utils"
	"testing"
)

func TestSearch(t *testing.T) {
	gm.RegisterFailHandler(gk.Fail)
	gk.RunSpecs(t, "Service Suite")
}

var _ = gk.Describe("Test my api :)", func() {
	InitializeTestService()

	apiName := "AddressService"
	version := "v1"

	gk.Describe("GET /"+apiName+"/healthcheck", func() {

		request := testUtil.CreateTestRequest("GET", "/"+apiName+"/healthcheck")
		response := GetResponse(request)
		gk.Context("then the response", func() {

			gk.It("should return api health status", func() {
				testUtil.MatchHeaderStatus(response)
				responseBody, _ := testUtil.GetHTTPResponseAndAddressResult(response.Body.String())
				testUtil.MatchSuccessResponseStatus(responseBody)
				validateHealthCheckResponse(response.Body.String())
			})
		})
	})

	gk.Describe("GET /"+apiName+"/"+version+"/address", func() {

		request := testUtil.CreateTestRequest("GET", "/"+apiName+"/"+version+"/address")
		response := GetResponse(request)
		gk.Context("then the response", func() {

			gk.It("should successfully get", func() {
				testUtil.MatchHeaderStatus(response)
				responseBody, _ := testUtil.GetHTTPResponseAndAddressResult(response.Body.String())
				testUtil.MatchVersionableNotFound(responseBody)
			})
		})
	})

})
