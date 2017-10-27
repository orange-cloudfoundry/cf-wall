package main

import            "net/http"
import            "net/url"
import            "fmt"
import            "errors"
import            "encoding/json"
import            "code.cloudfoundry.org/clock"
import            "code.cloudfoundry.org/lager"
import uaaclient  "code.cloudfoundry.org/uaa-go-client"
import uaaconfig  "code.cloudfoundry.org/uaa-go-client/config"
import log        "github.com/sirupsen/logrus"

type UaaCli struct
{
  Client uaaclient.Client
  Config *AppConfig
  Token string
}

type UaaUser struct {
  Id        string `json:"id"`
  FirstName string `json:"firstname"`
  LastName  string `json:"lastname"`
  Email     string `json:"email"`
}

func NewUaaCli(p_conf *AppConfig) (*UaaCli, error) {
  l_conf := &uaaconfig.Config{
    ClientName:       p_conf.UaaClientName,
    ClientSecret:     p_conf.UaaClientSecret,
    UaaEndpoint:      p_conf.UaaEndPoint,
  }


  l_cli, l_err := uaaclient.NewClient(lager.NewLogger("cfy-wall"), l_conf, clock.NewClock())
  if (l_err != nil) {
    log.WithError(l_err).Error("unable to create uaa client", l_err)
    return nil, l_err
  }

  return &UaaCli {
    Client: l_cli,
    Config: p_conf,
    Token:  "",
  }, nil
}

func (self *UaaCli) ensureToken() (error) {
  if (self.Token == "") {
    log.Info("fetching client token from UAA api")
    l_tok, l_err := self.Client.FetchToken(true)
    if (l_err != nil) {
      log.WithError(l_err).Error("unable to fetch UAA client token")
      return l_err
    }
    self.Token = l_tok.AccessToken
  }
  return nil
}

func (self *UaaCli) sendRequest(p_url *url.URL) (*http.Response, error) {
  if l_err := self.ensureToken(); l_err != nil {
    return nil, l_err
  }

  l_httpCli := http.Client{}

  l_headers := http.Header{}
  l_headers.Add("Authorization", fmt.Sprintf("bearer %s", self.Token))
  l_headers.Add("Accept",        "application/json")

  l_req     := &http.Request{
    Method : "GET",
    URL    : p_url,
    Header : l_headers,
  }

  log.WithFields(log.Fields{
    "url" : p_url,
  }).Info("sending request to UAA api")
  l_res, l_err := l_httpCli.Do(l_req)
  if (l_err != nil) {
    log.WithError(l_err).WithFields(log.Fields{
      "url" : p_url,
    }).Error("error while sending request to UAA api")
    return nil, l_err
  }

  log.WithFields(log.Fields{
    "url" : p_url,
    "response" : l_res,
  }).Debug("got UAA api response")

  if (l_res.StatusCode != 200) {
    log.WithError(l_err).WithFields(log.Fields{
      "url" : p_url,
    }).Error("error while sending request to UAA api")
    return nil, errors.New("uaa-client-request")
  }

  return l_res, nil
}

type userListUaaResponse struct {
  StartIndex   int  `json:"startIndex"`
  ItemsPerPage int  `json:"itemsPerPage"`
  TotalResults int  `json:"totalResults"`
  Resources    []struct {
    Id string   `json:"id"`
    Active bool `json:"active"`
    Emails []struct {
      Value string `json:"value"`
    } `json:"emails"`
    Name struct {
      LastName string  `json:"familyName"`
      FirstName string `json:"givenName"`
    } `json:"name"`
  } `json:"resources"`
}


func (self *UaaCli) getUserListIndex(p_idx int) (*userListUaaResponse, error) {
  log.WithFields(log.Fields{
    "index" : p_idx,
    "fields" : "id,emails,name,active",
  }).Debug("requesting UAA users index")

  l_urlfmt := "%s/Users?startIndex=%d&count=%d&attributes=id,emails,name,active"
  l_urlstr := fmt.Sprintf(l_urlfmt, self.Config.UaaEndPoint, p_idx, 500)
  l_url, _ := url.Parse(l_urlstr)

  l_res, l_err := self.sendRequest(l_url)
  if (l_err != nil) {
    return nil, l_err
  }

  l_decoder := json.NewDecoder(l_res.Body)
  l_data    := userListUaaResponse{}
  l_err     = l_decoder.Decode(&l_data)
  if (l_err != nil) {
    log.WithError(l_err).Error("unexpected UAA api user list response format")
    return nil, l_err
  }

  log.WithFields(log.Fields{
    "reponse" : l_data,
  }).Debug("UAA api users response")

  return &l_data, nil
}

func append_users(p_users *[]UaaUser, p_uaaUsers *userListUaaResponse) {
  for _, c_el := range p_uaaUsers.Resources {
    if c_el.Active {
      l_user := UaaUser{
        Id        : c_el.Id,
        FirstName : c_el.Name.FirstName,
        LastName  : c_el.Name.LastName,
      }
      if (len(c_el.Emails) != 0) {
        l_user.Email = c_el.Emails[0].Value
      }
      *p_users = append(*p_users, l_user)
    }
  }
}

func (self *UaaCli) GetUserList() ([]UaaUser, error) {
  log.Info("requesting users on UAA api")

  l_res := make([]UaaUser, 0)
  l_first, l_err := self.getUserListIndex(0)
  if (l_err != nil) {
    return l_res, l_err
  }
  append_users(&l_res, l_first)

  for c_idx := l_first.ItemsPerPage; c_idx < l_first.TotalResults; c_idx += l_first.ItemsPerPage {
    l_list, l_err := self.getUserListIndex(c_idx)
    if (l_err != nil) {
      return l_res, l_err
    }
    append_users(&l_res, l_list)
  }

  return l_res, nil
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
