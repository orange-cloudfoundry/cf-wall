package core

import "net/http"
import "net/url"
import "fmt"
import "encoding/json"
import "github.com/pkg/errors"
import log "github.com/sirupsen/logrus"
import "code.cloudfoundry.org/clock"
import "code.cloudfoundry.org/lager"
import uaaclient "code.cloudfoundry.org/uaa-go-client"
import uaaconfig "code.cloudfoundry.org/uaa-go-client/config"

type UaaCli struct {
	Client   uaaclient.Client
	Endpoint string
	Token    string
}

type UaaUser struct {
	Id        string `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
}

func NewUaaCli(pConf *AppConfig) (*UaaCli, error) {
	lConf := &uaaconfig.Config{
		ClientName:   pConf.UaaClientName,
		ClientSecret: pConf.UaaClientSecret,
		UaaEndpoint:  pConf.UaaEndPoint,
		SkipVerification: pConf.UaaSkipVerify,
	}

	lCli, lErr := uaaclient.NewClient(lager.NewLogger("cf-wall"), lConf, clock.NewClock())
	if lErr != nil {
		log.WithError(lErr).Error("failed to create uaaclient")
		return nil, lErr
	}

	return &UaaCli{
		Client: lCli,
		Endpoint: pConf.UaaEndPoint,
		Token:  "",
	}, nil
}

func (self *UaaCli) ensureToken() error {
	if self.Token == "" {
		log.Info("fetching client token from UAA api")
		lTok, lErr := self.Client.FetchToken(true)
		if lErr != nil {
			log.WithError(lErr).Error("unable to fetch UAA client token")
			return lErr
		}
		self.Token = lTok.AccessToken
	}
	return nil
}

func (self *UaaCli) sendRequest(pUrl *url.URL) (*http.Response, error) {
	if lErr := self.ensureToken(); lErr != nil {
		return nil, lErr
	}

	lHttpCli := http.Client{}
	lHeaders := http.Header{}

	lHeaders.Add("Authorization", fmt.Sprintf("bearer %s", self.Token))
	lHeaders.Add("Accept", "application/json")

	lReq := &http.Request{
		Method: "GET",
		URL:    pUrl,
		Header: lHeaders,
	}

	log.WithFields(log.Fields{
		"url": pUrl,
	}).Info("sending request to UAA api")

	lRes, lErr := lHttpCli.Do(lReq)
	if lErr != nil {
		log.WithError(lErr).WithFields(log.Fields{
			"url": pUrl,
		}).Error("error while sending request to UAA api")
		return nil, lErr
	}

	log.WithFields(log.Fields{
		"url":      pUrl,
		"response": lRes,
	}).Debug("got UAA api response")

	if lRes.StatusCode != 200 {
		log.WithError(lErr).WithFields(log.Fields{
			"url": pUrl,
		}).Error("error while sending request to UAA api")
		return nil, errors.New("uaa-client-request")
	}

	return lRes, nil
}

type userListUaaResponse struct {
	StartIndex   int `json:"startIndex"`
	ItemsPerPage int `json:"itemsPerPage"`
	TotalResults int `json:"totalResults"`
	Resources    []struct {
		Id     string `json:"id"`
		Active bool   `json:"active"`
		Emails []struct {
			Value string `json:"value"`
		} `json:"emails"`
		Name struct {
			LastName  string `json:"familyName"`
			FirstName string `json:"givenName"`
		} `json:"name"`
	} `json:"resources"`
}

func (self *UaaCli) getUserListIndex(pIdx int) (*userListUaaResponse, error) {
	log.WithFields(log.Fields{
		"index":  pIdx,
		"fields": "id,emails,name,active",
	}).Debug("requesting UAA users index")

	lUrlfmt := "%s/Users?startIndex=%d&count=%d&attributes=id,emails,name,active"
	lUrlstr := fmt.Sprintf(lUrlfmt, self.Endpoint, pIdx, 500)
	lUrl, _ := url.Parse(lUrlstr)

	lRes, lErr := self.sendRequest(lUrl)
	if lErr != nil {
		return nil, lErr
	}

	lDecoder := json.NewDecoder(lRes.Body)
	lData := userListUaaResponse{}
	lErr = lDecoder.Decode(&lData)
	if lErr != nil {
		log.WithError(lErr).Error("unexpected UAA api user list response format")
		return nil, lErr
	}

	log.WithFields(log.Fields{
		"reponse": lData,
	}).Debug("UAA api users response")

	return &lData, nil
}

func append_users(pUsers *[]UaaUser, pUaaUsers *userListUaaResponse) {
	for _, cEl := range pUaaUsers.Resources {
		if cEl.Active {
			lUser := UaaUser{
				Id:        cEl.Id,
				FirstName: cEl.Name.FirstName,
				LastName:  cEl.Name.LastName,
			}
			if len(cEl.Emails) != 0 {
				lUser.Email = cEl.Emails[0].Value
			}
			*pUsers = append(*pUsers, lUser)
		}
	}
}

func (self *UaaCli) GetUserList() ([]UaaUser, error) {
	log.Info("requesting users on UAA api")

	lRes := make([]UaaUser, 0)
	lFirst, lErr := self.getUserListIndex(0)
	if lErr != nil {
		return lRes, lErr
	}
	append_users(&lRes, lFirst)

	for cIdx := lFirst.ItemsPerPage; cIdx < lFirst.TotalResults; cIdx += lFirst.ItemsPerPage {
		lList, lErr := self.getUserListIndex(cIdx)
		if lErr != nil {
			return lRes, lErr
		}
		append_users(&lRes, lList)
	}

	return lRes, nil
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
