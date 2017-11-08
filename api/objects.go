package api

import "fmt"
import "net/url"
import "net/http"
import "errors"
import "github.com/gorilla/mux"
import log "github.com/sirupsen/logrus"
import "github.com/orange-cloudfoundry/cf-wall/core"

type Org struct {
	Name string `json:"name"`
	Id   string `json:"guid"`
}

type Space struct {
	Name  string `json:"name"`
	Id    string `json:"guid"`
	OrgId string `json:"org_guid"`
}

type User struct {
	Name string `json:"name"`
	Id   string `json:"guid"`
}

type Buildpack struct {
	Name string `json:"name"`
	Id   string `json:"guid"`
}

type Service struct {
	Name string `json:"name"`
	Id   string `json:"guid"`
}

type ObjectHandler struct {
	Config    *core.AppConfig
	CCCreator func(*http.Request) core.CFClient
}

func NewObjectHandler(pConf *core.AppConfig, pRouter *mux.Router) *ObjectHandler {
	lObj := ObjectHandler{
		Config: pConf,
	}
	lObj.CCCreator = lObj.createCli

	pRouter.Path("/v1/orgs").
		HandlerFunc(core.DecorateHandler(lObj.getOrgs))
	pRouter.Path("/v1/orgs/{guid}/spaces").
		HandlerFunc(core.DecorateHandler(lObj.getOrgSpaces))
	pRouter.Path("/v1/spaces").
		HandlerFunc(core.DecorateHandler(lObj.getSpaces))
	pRouter.Path("/v1/users").
		HandlerFunc(core.DecorateHandler(lObj.getUsers))
	pRouter.Path("/v1/buildpacks").
		HandlerFunc(core.DecorateHandler(lObj.getBuildpacks))
	pRouter.Path("/v1/services").
		HandlerFunc(core.DecorateHandler(lObj.getServices))

	return &lObj
}

func (self *ObjectHandler) createCli(pReq *http.Request) core.CFClient {
	lApi, lErr := core.NewCCCliFromRequest(self.Config.CCEndPoint, pReq, self.Config.CCSkipVerify)
	if lErr != nil {
		panic(core.NewHttpError(lErr, 400, 10))
	}
	return lApi
}

func (self *ObjectHandler) getServices(pRes http.ResponseWriter, pReq *http.Request) {
	lApi := self.CCCreator(pReq)

	log.Info("reading services from CC api")
	lUsers, lErr := lApi.ListServices()
	if lErr != nil {
		lUerr := errors.New("unable to read services from CC api")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lUerr, 500, 50))
	}

	lRes := []Service{}
	for _, cEl := range lUsers {
		lElem := Service{cEl.Label, cEl.Guid}
		lRes = append(lRes, lElem)
	}
	log.WithFields(log.Fields{"services": lRes}).
		Debug("fetched services from CC api")
	core.WriteJson(pRes, lRes)
}

func (self *ObjectHandler) getBuildpacks(pRes http.ResponseWriter, pReq *http.Request) {
	lApi := self.CCCreator(pReq)

	log.Info("reading buildpacks from CC api")
	lUsers, lErr := lApi.ListBuildpacks()
	if lErr != nil {
		lUerr := errors.New("unable to read buildpacks from CC api")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lUerr, 500, 50))
	}

	lRes := []Buildpack{}
	for _, cEl := range lUsers {
		lElem := Buildpack{cEl.Name, cEl.Guid}
		lRes = append(lRes, lElem)
	}
	log.WithFields(log.Fields{"buildpacks": lRes}).
		Debug("fetched buildpacks from CC api")
	core.WriteJson(pRes, lRes)
}

func (self *ObjectHandler) getUsers(pRes http.ResponseWriter, pReq *http.Request) {
	lApi := self.CCCreator(pReq)

	log.Info("reading users from CC api")
	lQuery := url.Values{}
	lQuery.Set("results-per-page", "100")
	lUsers, lErr := lApi.ListUsersByQuery(lQuery)
	if lErr != nil {
		lUerr := errors.New("unable to read users from CC api")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lUerr, 500, 50))
	}

	lRes := []User{}
	for _, cEl := range lUsers {
		lElem := User{cEl.Username, cEl.Guid}
		lRes = append(lRes, lElem)
	}
	log.WithFields(log.Fields{"users": lRes}).
		Debug("fetched users from CC api")
	core.WriteJson(pRes, lRes)
}

func (self *ObjectHandler) getOrgs(pRes http.ResponseWriter, pReq *http.Request) {
	lApi := self.CCCreator(pReq)

	log.Info("reading orgs from CC api")
	lOrgs, lErr := lApi.ListOrgs()
	if lErr != nil {
		lUerr := errors.New("unable to read organizations from CC api")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lUerr, 500, 50))
	}

	lRes := []Org{}
	for _, cEl := range lOrgs {
		lElem := Org{cEl.Name, cEl.Guid}
		lRes = append(lRes, lElem)
	}

	log.WithFields(log.Fields{"orgs": lRes}).
		Debug("fetched organization from CC api")
	core.WriteJson(pRes, lRes)
}

func (self *ObjectHandler) getOrgSpaces(pRes http.ResponseWriter, pReq *http.Request) {
	lVars := mux.Vars(pReq)
	lOrgID := lVars["guid"]

	lApi := self.CCCreator(pReq)

	log.WithFields(log.Fields{
		"org": lOrgID,
	}).Info("reading spaces member of org from CC api")
	lQuery := url.Values{}
	lQuery.Add("q", fmt.Sprintf("organization_guid:%s", lOrgID))
	lSpaces, lErr := lApi.ListSpacesByQuery(lQuery)
	if lErr != nil {
		lUerr := errors.New("unable to read spaces from CC api")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lUerr, 500, 50))
	}

	lRes := []Space{}
	for _, cEl := range lSpaces {
		lElem := Space{cEl.Name, cEl.Guid, cEl.OrganizationGuid}
		lRes = append(lRes, lElem)
	}
	log.WithFields(log.Fields{"spaces": lRes}).
		Debug("fetched spaces from CC api")
	core.WriteJson(pRes, lRes)
}



func (self *ObjectHandler) getSpaces(pRes http.ResponseWriter, pReq *http.Request) {
	lApi := self.CCCreator(pReq)

	log.Info("reading spaces from CC api")
	lSpaces, lErr := lApi.ListSpaces()
	if lErr != nil {
		lUerr := errors.New("unable to read services from CC api")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lUerr, 500, 50))
	}

	lRes := []Space{}
	for _, cEl := range lSpaces {
		lElem := Space{cEl.Name, cEl.Guid, cEl.OrganizationGuid}
		lRes = append(lRes, lElem)
	}
	log.WithFields(log.Fields{"spaces": lRes}).
		Debug("fetched spaces from CC api")
	core.WriteJson(pRes, lRes)
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
