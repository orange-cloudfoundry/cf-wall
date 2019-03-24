package api

import "fmt"
import "net/url"
import "net/http"
import "errors"
import "github.com/gorilla/mux"
import log "github.com/sirupsen/logrus"
import "github.com/orange-cloudfoundry/cf-wall/core"
import "regexp"

// Org --
type Org struct {
	Name string `json:"name"`
	ID   string `json:"guid"`
}

// Space --
type Space struct {
	Name  string `json:"name"`
	ID    string `json:"guid"`
	OrgID string `json:"org_guid"`
}

// User --
type User struct {
	Name string `json:"name"`
	ID   string `json:"guid"`
}

// Buildpack --
type Buildpack struct {
	Name string `json:"name"`
	ID   string `json:"guid"`
}

// Service --
type Service struct {
	Name string `json:"name"`
	ID   string `json:"guid"`
}

// ObjectHandler --
type ObjectHandler struct {
	Config    *core.AppConfig
	CCCreator func(*http.Request) core.CFClient
}

// NewObjectHandler --
func NewObjectHandler(pConf *core.AppConfig, pRouter *mux.Router) *ObjectHandler {
	obj := ObjectHandler{
		Config: pConf,
	}
	obj.CCCreator = obj.createCli

	pRouter.Path("/v1/orgs").
		HandlerFunc(core.DecorateHandler(obj.getOrgs))
	pRouter.Path("/v1/orgs/{guid}/spaces").
		HandlerFunc(core.DecorateHandler(obj.getOrgSpaces))
	pRouter.Path("/v1/spaces").
		HandlerFunc(core.DecorateHandler(obj.getSpaces))
	pRouter.Path("/v1/users").
		HandlerFunc(core.DecorateHandler(obj.getUsers))
	pRouter.Path("/v1/buildpacks").
		HandlerFunc(core.DecorateHandler(obj.getBuildpacks))
	pRouter.Path("/v1/buildpacks/{regexp}").
		HandlerFunc(core.DecorateHandler(obj.getBuildpackRegexp))
	pRouter.Path("/v1/services").
		HandlerFunc(core.DecorateHandler(obj.getServices))

	return &obj
}

func (s *ObjectHandler) createCli(pReq *http.Request) core.CFClient {
	api, err := core.NewCCCliFromRequest(s.Config.CCEndPoint, pReq, s.Config.CCSkipVerify)
	if err != nil {
		panic(core.NewHttpError(err, 400, 10))
	}
	return api
}

func (s *ObjectHandler) getServices(pRes http.ResponseWriter, pReq *http.Request) {
	api := s.CCCreator(pReq)

	log.Info("reading services from CC api")
	users, err := api.ListServices()
	if err != nil {
		uerr := errors.New("unable to read services from CC api")
		log.WithError(err).Error(uerr.Error())
		panic(core.NewHttpError(uerr, 500, 50))
	}

	res := []Service{}
	for _, cEl := range users {
		elem := Service{cEl.Label, cEl.Guid}
		res = append(res, elem)
	}
	log.WithFields(log.Fields{"services": res}).
		Debug("fetched services from CC api")
	core.WriteJson(pRes, res)
}

func (s *ObjectHandler) getBuildpackRegexp(pRes http.ResponseWriter, pReq *http.Request) {
	vars := mux.Vars(pReq)
	re, err := regexp.Compile(vars["regexp"])
	if err != nil {
		uerr := fmt.Errorf("invalid input regexp '%s'", vars["regexp"])
		log.WithError(err).Error(uerr.Error())
		panic(core.NewHttpError(uerr, 400, 40))
	}

	api := s.CCCreator(pReq)
	log.Info("reading buildpacks from CC api")
	users, err := api.ListBuildpacks()
	if err != nil {
		uerr := errors.New("unable to read buildpacks from CC api")
		log.WithError(err).Error(uerr.Error())
		panic(core.NewHttpError(uerr, 500, 50))
	}

	res := []Buildpack{}
	for _, cEl := range users {
		if re.Match([]byte(cEl.Name)) {
			elem := Buildpack{cEl.Name, cEl.Guid}
			res = append(res, elem)
		}
	}
	log.WithFields(log.Fields{"buildpacks": res}).
		Debug("fetched buildpacks from CC api")
	core.WriteJson(pRes, res)
}

func (s *ObjectHandler) getBuildpacks(pRes http.ResponseWriter, pReq *http.Request) {
	api := s.CCCreator(pReq)

	log.Info("reading buildpacks from CC api")
	users, err := api.ListBuildpacks()
	if err != nil {
		uerr := errors.New("unable to read buildpacks from CC api")
		log.WithError(err).Error(uerr.Error())
		panic(core.NewHttpError(uerr, 500, 50))
	}

	res := []Buildpack{}
	for _, cEl := range users {
		elem := Buildpack{cEl.Name, cEl.Guid}
		res = append(res, elem)
	}
	log.WithFields(log.Fields{"buildpacks": res}).
		Debug("fetched buildpacks from CC api")
	core.WriteJson(pRes, res)
}

func (s *ObjectHandler) getUsers(pRes http.ResponseWriter, pReq *http.Request) {
	api := s.CCCreator(pReq)

	log.Info("reading users from CC api")
	query := url.Values{}
	query.Set("results-per-page", "100")
	users, err := api.ListUsersByQuery(query)
	if err != nil {
		uerr := errors.New("unable to read users from CC api")
		log.WithError(err).Error(uerr.Error())
		panic(core.NewHttpError(uerr, 500, 50))
	}

	res := []User{}
	for _, cEl := range users {
		elem := User{cEl.Username, cEl.Guid}
		res = append(res, elem)
	}
	log.WithFields(log.Fields{"users": res}).
		Debug("fetched users from CC api")
	core.WriteJson(pRes, res)
}

func (s *ObjectHandler) getOrgs(pRes http.ResponseWriter, pReq *http.Request) {
	api := s.CCCreator(pReq)

	log.Info("reading orgs from CC api")
	orgs, err := api.ListOrgs()
	if err != nil {
		uerr := errors.New("unable to read organizations from CC api")
		log.WithError(err).Error(uerr.Error())
		panic(core.NewHttpError(uerr, 500, 50))
	}

	res := []Org{}
	for _, cEl := range orgs {
		elem := Org{cEl.Name, cEl.Guid}
		res = append(res, elem)
	}

	log.WithFields(log.Fields{"orgs": res}).
		Debug("fetched organization from CC api")
	core.WriteJson(pRes, res)
}

func (s *ObjectHandler) getOrgSpaces(pRes http.ResponseWriter, pReq *http.Request) {
	vars := mux.Vars(pReq)
	orgID := vars["guid"]

	api := s.CCCreator(pReq)

	log.WithFields(log.Fields{
		"org": orgID,
	}).Info("reading spaces member of org from CC api")
	query := url.Values{}
	query.Add("q", fmt.Sprintf("organization_guid:%s", orgID))
	spaces, err := api.ListSpacesByQuery(query)
	if err != nil {
		uerr := errors.New("unable to read spaces from CC api")
		log.WithError(err).Error(uerr.Error())
		panic(core.NewHttpError(uerr, 500, 50))
	}

	res := []Space{}
	for _, cEl := range spaces {
		elem := Space{cEl.Name, cEl.Guid, cEl.OrganizationGuid}
		res = append(res, elem)
	}
	log.WithFields(log.Fields{"spaces": res}).
		Debug("fetched spaces from CC api")
	core.WriteJson(pRes, res)
}

func (s *ObjectHandler) getSpaces(pRes http.ResponseWriter, pReq *http.Request) {
	api := s.CCCreator(pReq)

	log.Info("reading spaces from CC api")
	spaces, err := api.ListSpaces()
	if err != nil {
		uerr := errors.New("unable to read services from CC api")
		log.WithError(err).Error(uerr.Error())
		panic(core.NewHttpError(uerr, 500, 50))
	}

	res := []Space{}
	for _, cEl := range spaces {
		elem := Space{cEl.Name, cEl.Guid, cEl.OrganizationGuid}
		res = append(res, elem)
	}
	log.WithFields(log.Fields{"spaces": res}).
		Debug("fetched spaces from CC api")
	core.WriteJson(pRes, res)
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
