package main

import "fmt"
import "net/url"
import "net/http"
import "errors"
import "github.com/gorilla/mux"
import log "github.com/sirupsen/logrus"

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

type OrgResponse = []Org
type SpaceResponse = []Space
type UserResponse = []User
type BuildpackResponse = []Buildpack
type ServiceResponse = []Service

type ObjectHandler struct {
}

func NewObjectHandler(p_router *mux.Router) ObjectHandler {
	l_obj := ObjectHandler{}

	p_router.Path("/v1/orgs").
		HandlerFunc(DecorateHandler(l_obj.getOrgs))
	p_router.Path("/v1/orgs/{guid}/spaces").
		HandlerFunc(DecorateHandler(l_obj.getOrgSpaces))
	p_router.Path("/v1/spaces").
		HandlerFunc(DecorateHandler(l_obj.getSpaces))
	p_router.Path("/v1/users").
		HandlerFunc(DecorateHandler(l_obj.getUsers))
	p_router.Path("/v1/buildpacks").
		HandlerFunc(DecorateHandler(l_obj.getBuildpacks))
	p_router.Path("/v1/services").
		HandlerFunc(DecorateHandler(l_obj.getServices))

	return l_obj
}

func (self *ObjectHandler) getServices(p_res http.ResponseWriter, p_req *http.Request) {
	l_api, l_err := NewCCCliFromRequest(&GApp.Config, p_req)
	if l_err != nil {
		panic(HttpError{l_err, 400, 10})
	}

	log.Info("reading services from CC api")
	l_users, l_err := l_api.ListServices()
	if l_err != nil {
		l_uerr := errors.New("unable to read services from CC api")
		log.WithError(l_err).Error(l_uerr.Error())
		panic(HttpError{l_uerr, 400, 10})
	}

	l_res := ServiceResponse{}
	for _, c_el := range l_users {
		l_elem := Service{c_el.Label, c_el.Guid}
		l_res = append(l_res, l_elem)
	}
	log.WithFields(log.Fields{"services": l_res}).
		Debug("fetched services from CC api")
	WriteJson(p_res, l_res)
}

func (self *ObjectHandler) getBuildpacks(p_res http.ResponseWriter, p_req *http.Request) {
	l_api, l_err := NewCCCliFromRequest(&GApp.Config, p_req)
	if l_err != nil {
		panic(HttpError{l_err, 400, 10})
	}

	log.Info("reading buildpacks from CC api")
	l_users, l_err := l_api.ListBuildpacks()
	if l_err != nil {
		l_uerr := errors.New("unable to read buildpacks from CC api")
		log.WithError(l_err).Error(l_uerr.Error())
		panic(HttpError{l_uerr, 400, 10})
	}

	l_res := BuildpackResponse{}
	for _, c_el := range l_users {
		l_elem := Buildpack{c_el.Name, c_el.Guid}
		l_res = append(l_res, l_elem)
	}
	log.WithFields(log.Fields{"buildpacks": l_res}).
		Debug("fetched buildpacks from CC api")
	WriteJson(p_res, l_res)
}

func (self *ObjectHandler) getUsers(p_res http.ResponseWriter, p_req *http.Request) {
	l_api, l_err := NewCCCliFromRequest(&GApp.Config, p_req)
	if l_err != nil {
		panic(HttpError{l_err, 400, 10})
	}

	log.Info("reading users from CC api")
	l_users, l_err := l_api.ListUsers()
	if l_err != nil {
		l_uerr := errors.New("unable to read users from CC api")
		log.WithError(l_err).Error(l_uerr.Error())
		panic(HttpError{l_uerr, 400, 10})
	}

	l_res := UserResponse{}
	for _, c_el := range l_users {
		l_elem := User{c_el.Username, c_el.Guid}
		l_res = append(l_res, l_elem)
	}
	log.WithFields(log.Fields{"users": l_res}).
		Debug("fetched users from CC api")
	WriteJson(p_res, l_res)
}

func (self *ObjectHandler) getOrgs(p_res http.ResponseWriter, p_req *http.Request) {
	l_api, l_err := NewCCCliFromRequest(&GApp.Config, p_req)
	if l_err != nil {
		panic(HttpError{l_err, 400, 10})
	}

	log.Info("reading orgs from CC api")
	l_orgs, l_err := l_api.ListOrgs()
	if l_err != nil {
		l_uerr := errors.New("unable to read organizations from CC api")
		log.WithError(l_err).Error(l_uerr.Error())
		panic(HttpError{l_uerr, 400, 10})
	}

	l_res := OrgResponse{}
	for _, c_el := range l_orgs {
		l_elem := Org{c_el.Name, c_el.Guid}
		l_res = append(l_res, l_elem)
	}


	log.WithFields(log.Fields{"orgs": l_res}).
		Debug("fetched organization from CC api")
	WriteJson(p_res, l_res)
}

func (self *ObjectHandler) getOrgSpaces(p_res http.ResponseWriter, p_req *http.Request) {
	l_vars := mux.Vars(p_req)
	l_orgID := l_vars["guid"]

	l_api, l_err := NewCCCliFromRequest(&GApp.Config, p_req)
	if l_err != nil {
		panic(HttpError{l_err, 400, 10})
	}

	log.WithFields(log.Fields{
		"org": l_orgID,
	}).Info("reading spaces member of org from CC api")
	l_query := url.Values{}
	l_query.Add("q", fmt.Sprintf("organization_guid:%s", l_orgID))
	l_spaces, l_err := l_api.ListSpacesByQuery(l_query)
	if l_err != nil {
		l_uerr := errors.New("unable to read spaces from CC api")
		log.WithError(l_err).Error(l_uerr.Error())
		panic(HttpError{l_uerr, 400, 10})
	}

	l_res := SpaceResponse{}
	for _, c_el := range l_spaces {
		l_elem := Space{c_el.Name, c_el.Guid, c_el.OrganizationGuid}
		l_res = append(l_res, l_elem)
	}
	log.WithFields(log.Fields{"spaces": l_res}).
		Debug("fetched spaces from CC api")
	WriteJson(p_res, l_res)
}

func (self *ObjectHandler) getSpaces(p_res http.ResponseWriter, p_req *http.Request) {
	l_api, l_err := NewCCCliFromRequest(&GApp.Config, p_req)
	if l_err != nil {
		panic(HttpError{l_err, 400, 10})
	}

	log.Info("reading spaces from CC api")
	l_spaces, l_err := l_api.ListSpaces()
	if l_err != nil {
		l_uerr := errors.New("unable to read services from CC api")
		log.WithError(l_err).Error(l_uerr.Error())
		panic(HttpError{l_uerr, 400, 10})
	}

	l_res := SpaceResponse{}
	for _, c_el := range l_spaces {
		l_elem := Space{c_el.Name, c_el.Guid, c_el.OrganizationGuid}
		l_res = append(l_res, l_elem)
	}
	log.WithFields(log.Fields{"spaces": l_res}).
		Debug("fetched spaces from CC api")
	WriteJson(p_res, l_res)
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
