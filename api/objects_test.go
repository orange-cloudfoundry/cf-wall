package api_test


import (
	"net/url"
	"errors"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	. "github.com/orange-cloudfoundry/cf-wall/api"
	"github.com/orange-cloudfoundry/cf-wall/core"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry-community/go-cfclient"

)

type FakeCli struct {
	Error error
}

func (self *FakeCli) ListSpaces() ([]cfclient.Space, error) {
	return []cfclient.Space{
		cfclient.Space{
			Name: "space-1",
			Guid: "fc95a4c6-b07f-4b9b-871e-1f5d67b06071",
			OrganizationGuid: "ddddb760-ef75-40f4-9d52-1bb557b61af8",
		},
		cfclient.Space{
			Name: "space-2",
			Guid: "4ca50e06-2a3c-49d5-a1e6-dffa3cf6a7b4",
			OrganizationGuid: "ddddb760-ef75-40f4-9d52-1bb557b61af8",
		},
		cfclient.Space{
			Name: "space-3",
			Guid: "ed6b2ad1-e7ca-4f94-8f12-46d5857aa571",
			OrganizationGuid: "91c1503b-9b4f-4ecf-a02c-4241fe217522",
		},
	}, self.Error
}

func (self *FakeCli) ListUsersByQuery(url.Values) (cfclient.Users, error) {
	return cfclient.Users{}, self.Error
}

func (self *FakeCli) ListServices() ([]cfclient.Service, error) {
	return []cfclient.Service{
		cfclient.Service{
			Label: "service-1",
			Guid: "2790478e-f500-43dd-93cf-6fa3c97292b9",
		},
		cfclient.Service{
			Label: "service-2",
			Guid: "e7289425-e9c6-42ab-a6fc-a1729d0849ca",
		},
	}, self.Error
}

func (self *FakeCli) ListBuildpacks() ([]cfclient.Buildpack, error) {
	return []cfclient.Buildpack{}, self.Error
}

func (self *FakeCli) ListUsers() (cfclient.Users, error) {
	return cfclient.Users{}, self.Error
}

func (self *FakeCli) ListOrgs() ([]cfclient.Org, error) {
	return []cfclient.Org{
		cfclient.Org{
			Name: "org-1",
			Guid: "ddddb760-ef75-40f4-9d52-1bb557b61af8",
		},
		cfclient.Org{
			Name: "org-2",
			Guid: "91c1503b-9b4f-4ecf-a02c-4241fe217522",
		},
	}, self.Error
}

func (self *FakeCli) ListSpacesByQuery(url.Values) ([]cfclient.Space, error) {
	return []cfclient.Space{}, self.Error
}

func (self *FakeCli) ListAppsByQuery(url.Values) ([]cfclient.App, error) {
	return []cfclient.App{}, self.Error
}

func (self *FakeCli) ListServiceInstancesByQuery(query url.Values) ([]cfclient.ServiceInstance, error) {
	return []cfclient.ServiceInstance{}, self.Error
}

func (self *FakeCli) ListServiceBindingsByQuery(query url.Values) ([]cfclient.ServiceBinding, error) {
	return []cfclient.ServiceBinding{}, self.Error
}

func assertJson(pRes *http.Response, pStruct interface{}) {
	lVal := pRes.Header.Get("Content-Type")
	Expect(lVal).To(Equal("application/json"), "with application/json")
	lDecoder := json.NewDecoder(pRes.Body)
	lErr := lDecoder.Decode(&pStruct)
	Expect(lErr).To(BeNil(), "with well structured json payload")
}

func assertStatusOk(pRes *http.Response, pErr error) {
	Expect(pRes.StatusCode).To(Equal(200),  "with status 200")
	Expect(pErr)           .To(BeNil(),     "with nil error")
}

func assertStatusKo(pRes *http.Response, pErr error, pStatus int) {
	Expect(pRes.StatusCode).To(Equal(pStatus),  "with error status")
	Expect(pErr)           .To(BeNil(),         "with nil error")
}

func getRequest(pUrl string, pPath string, pToken bool) (*http.Response, error) {
	lUrl, _  := url.Parse(pUrl + pPath)
	lHeaders := http.Header{}
	if pToken {
		lHeaders.Add("Authorization", "bearer secret")
	}
	lReq := &http.Request{
		Method: "GET",
		URL:    lUrl,
		Header: lHeaders,
	}
	lClient := http.Client{}
	lRes, lErr := lClient.Do(lReq)
	Expect(lErr).To   (BeNil(), "no error")
	Expect(lReq).NotTo(BeNil(), "valid http response")
	return lRes, lErr
}


func cliGen(pError error) (func(pReq *http.Request) core.CFClient) {
	return func(pReq *http.Request) core.CFClient {
		lOk := pReq.Header.Get("Authorization")
		if "" == lOk {
			panic(core.NewHttpError(errors.New("authorization header not found"), 400, 10))
		}
		return &FakeCli{ Error: pError }
	}
}

var _ = Describe("Objects", func() {
	var lHandler *ObjectHandler
	var lApiMock *httptest.Server


	BeforeEach(func() {
		lRouter  := mux.NewRouter()
		lHandler  = NewObjectHandler(&core.AppConfig{}, lRouter)
		lApiMock = httptest.NewServer(lRouter)
	})


	Context("With error CC api", func() {
		BeforeEach(func() {
			lHandler.CCCreator = cliGen(errors.New("internal api error"))
		})

		Context("With authorization header token", func() {

			It("get spaces", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/spaces", true)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 500)
				assertJson(lRes, &lDescr)
				Expect(lDescr.Code).To(Equal(50), "user code 10")
			})

			It("get orgs", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/orgs", true)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 500)
				assertJson(lRes, &lDescr)
				Expect(lDescr.Code).To(Equal(50), "user code 10")
			})

			It("get users", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/users", true)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 500)
				assertJson(lRes, &lDescr)
				Expect(lDescr.Code).To(Equal(50), "user code 10")
			})

			It("get services", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/services", true)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 500)
				assertJson(lRes, &lDescr)
				Expect(lDescr.Code).To(Equal(50), "user code 10")
			})

			It("get buildpacks", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/buildpacks", true)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 500)
				assertJson(lRes, &lDescr)
				Expect(lDescr.Code).To(Equal(50), "user code 10")
			})

			It("get org spaces", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/orgs/ddddb760-ef75-40f4-9d52-1bb557b61af8/spaces", true)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 500)
				assertJson(lRes, &lDescr)
				Expect(lDescr.Code).To(Equal(50), "user code 10")
			})
		})
	})

	Context("With running CC api", func() {
		BeforeEach(func() {
			lHandler.CCCreator = cliGen(nil)
		})

		Context("Without authorization header token", func() {
			It("get spaces", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/spaces", false)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 400)
				assertJson(lRes, &lDescr)
			})
			It("get orgs", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/orgs", false)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 400)
				assertJson(lRes, &lDescr)
			})
			It("get services", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/services", false)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 400)
				assertJson(lRes, &lDescr)
			})
			It("get buildpacks", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/buildpacks", false)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 400)
				assertJson(lRes, &lDescr)
			})
			It("get users", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/users", false)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 400)
				assertJson(lRes, &lDescr)
			})
			It("get org spaces", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/orgs/ddddb760-ef75-40f4-9d52-1bb557b61af8/spaces", false)
				lDescr := struct {
					Code  int    `json:"code"`
					Error string `json:"error"`
				}{}
				assertStatusKo(lRes, lErr, 400)
				assertJson(lRes, &lDescr)
			})
		})

		Context("With authorization header token", func() {
			It("should not fail", func() {
				lRes, lErr := getRequest(lApiMock.URL, "/v1/spaces", true)
				var lData []Space
				assertStatusOk(lRes, lErr)
				assertJson(lRes, &lData)
				Expect(lData).Should(HaveLen(3))
			})
		})
	})


	AfterEach(func() {
		if lApiMock != nil {
			lApiMock.Close()
		}
	})
})
