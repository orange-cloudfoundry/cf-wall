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
	ListUsers() (cfclient.Users, error)
	ListOrgs() ([]cfclient.Org, error)
	ListSpacesByQuery(url.Values) ([]cfclient.Space, error)
}

func NewCCCliFromRequest(pUrl string, pReq *http.Request) (CFClient, error) {
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

	lCli, lErr := NewCCCli(pUrl, lParts[1])
	if lErr != nil {
		log.WithError(lErr).Error("unable to create CC client")
		return nil, lErr
	}

	return lCli, nil
}

func NewCCCli(pUrl string, pToken string) (*cfclient.Client, error) {
	log.WithFields(log.Fields{
		"endpoint": pUrl,
	}).Debug("creating CC client")

	lConf := cfclient.Config{
		ApiAddress: pUrl,
		Token:      pToken,
	}
	return cfclient.NewClient(&lConf)
}
