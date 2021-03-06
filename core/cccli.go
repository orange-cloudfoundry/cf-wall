package core

import "net/http"
import "strings"
import "errors"
import "net/url"
import "github.com/cloudfoundry-community/go-cfclient"
import log "github.com/sirupsen/logrus"


type CFClient interface {
	ListSpaces() ([]cfclient.Space, error)
	ListUsersByQuery(url.Values) (cfclient.Users, error)
	ListServices() ([]cfclient.Service, error)
	ListBuildpacks() ([]cfclient.Buildpack, error)
	ListOrgs() ([]cfclient.Org, error)
	ListSpacesByQuery(url.Values) ([]cfclient.Space, error)
	ListAppsByQuery(query url.Values) ([]cfclient.App, error)
	// ListServicesByQuery(query url.Values) ([]cfclient.Service, error)
	ListServiceInstancesByQuery(query url.Values) ([]cfclient.ServiceInstance, error)
	ListServiceBindingsByQuery(query url.Values) ([]cfclient.ServiceBinding, error)
}

func NewCCCliFromRequest(pUrl string, pReq *http.Request, pSkipVerify bool) (CFClient, error) {
	lAuth, lOk := pReq.Header["Authorization"]
	if (lOk == false) || (len(lAuth) == 0) {
		lErr := errors.New("Authorization header is mandatory")
		log.WithError(lErr).Error("unable to create CC client")
		return nil, lErr
	}

	lParts := strings.Fields(lAuth[0])
	if len(lParts) < 2 {
		lErr := errors.New("malformated Authorization header")
		log.WithError(lErr).Error("unable to create CC client")
		return nil, lErr
	}

	lCli, lErr := NewCCCli(pUrl, lParts[1], pSkipVerify)
	if lErr != nil {
		log.WithError(lErr).Error("unable to create CC client")
		return nil, lErr
	}

	return lCli, nil
}

func NewCCCli(pUrl string, pToken string, pSkipVerify bool) (*cfclient.Client, error) {
	log.WithFields(log.Fields{
		"endpoint": pUrl,
	}).Debug("creating CC client")

	lConf := cfclient.Config{
		ApiAddress: pUrl,
		Token:      pToken,
		SkipSslValidation: pSkipVerify,
	}

	return cfclient.NewClient(&lConf)
}
