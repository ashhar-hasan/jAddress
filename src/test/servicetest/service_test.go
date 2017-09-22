package servicetest

import (
	fconstants "github.com/jabong/florest-core/src/common/constants"
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
				gm.Expect(response.HeaderMap.Get("Content-Type")).To(gm.Equal("application/json"))
				gm.Expect(response.HeaderMap.Get("Cache-Control")).To(gm.Equal(""))
				gm.Expect(response.Code).To(gm.Equal(int(fconstants.HTTPStatusSuccessCode)))
				validateHealthCheckResponse(response.Body.String())
			})
		})
	})

	gk.Describe("GET /"+apiName+"/"+version+"/address", func() {

		request := testUtil.CreateTestRequest("GET", "/"+apiName+"/"+version+"/address")
		response := GetResponse(request)
		gk.Context("then the response", func() {

			gk.It("should successfully get", func() {
				gm.Expect(response.HeaderMap.Get("Content-Type")).To(gm.Equal("application/json"))
				gm.Expect(response.HeaderMap.Get("Cache-Control")).To(gm.Equal(""))
				gm.Expect(response.Code).To(gm.Equal(int(fconstants.HTTPStatusNotFound)))
			})
		})
	})

})
