package main

import      "fmt"
import      "strings"
import      "net/url"
import      "encoding/json"
import      "net/http"
import      "github.com/cloudfoundry-community/go-cfclient"
import      "github.com/gorilla/mux"
import log  "github.com/sirupsen/logrus"

type MessageReqCtx struct {
  CCCli *cfclient.Client
  UserMails map[string]string
  ReqData MessageRequest
  ResData MessageResponse
}

type MessageHandler struct {
  UaaCli *UaaCli
}

type MessageRequest struct {
  Users       []string  `json:"users"`
  Spaces      []string  `json:"spaces"`
  Orgs        []string  `json:"orgs"`
  Services    []string  `json:"services"`
  BuildPacks  []string  `json:"buildpacks"`
  Subject     string    `json:"subject"`
  Message     string    `json:"message"`
}

type MessageResponse = []string


func NewMessageHandler(p_app *App, p_router *mux.Router) (MessageHandler) {
  l_obj := MessageHandler{
    UaaCli: p_app.UaaCli,
  }

  p_router.Path("/v1/message").
    HandlerFunc(l_obj.HandleMessage).
    HeadersRegexp("Content-Type",  "application/json.*").
    Methods("POST")

  return l_obj
}

func (self *MessageHandler) createCtx(p_users map[string]string, p_req *http.Request) (*MessageReqCtx, error) {
  l_cccli, l_err := NewCCCliFromRequest(&GApp.Config, p_req)
  if (l_err != nil) {
    log.WithError(l_err).Error("unable to create CC client")
    return nil, l_err
  }

  l_ctx := MessageReqCtx{
    CCCli     : l_cccli,
    UserMails : p_users,
    ReqData   : MessageRequest{},
    ResData   : MessageResponse{},
  }

  l_decoder := json.NewDecoder(p_req.Body)
  l_err = l_decoder.Decode(&l_ctx.ReqData)
  if (l_err != nil) {
    return nil, l_err
  }

  return &l_ctx, nil
}

func (self *MessageHandler) getUaaUsers() (map[string]string, error) {
  l_res := make(map[string]string, 0)
  log.Debug("reading UAA users")
  l_users, l_err := self.UaaCli.GetUserList()
  if (l_err != nil) {
    log.WithError(l_err).Error("unable to featch UAA users")
    return l_res, l_err
  }
  for _, c_el := range l_users {
    l_res[c_el.Id] = c_el.Email
  }
  return l_res, nil
}

func (self *MessageHandler) HandleMessage(p_res http.ResponseWriter, p_req *http.Request) {
  l_users, l_err := self.getUaaUsers()
  if (l_err != nil) {
    WriteJsonError(p_res, 400, 10, l_err)
    return
  }

  l_ctx, l_err := self.createCtx(l_users, p_req)
  if (l_err != nil) {
    WriteJsonError(p_res, 400, 10, l_err)
    return
  }

  l_err = l_ctx.process()
  if (l_err != nil) {
    WriteJsonError(p_res, 500, 20, l_err)
    return
  }

  WriteJson(p_res, l_ctx.ResData)

  // create message
  // send message
}


func (self *MessageReqCtx) process() (error) {
  l_err := self.addOrgs(self.ReqData.Orgs)
  if (l_err != nil) {
    return l_err
  }
  l_err = self.addSpaces(self.ReqData.Spaces)
  if (l_err != nil) {
    return l_err
  }
  self.addUsers(self.ReqData.Users)
  // addServices
  // addBuidPack
  return nil
}

func (self *MessageReqCtx) addOrgs(p_orgs []string) (error) {
  if (len(p_orgs) != 0) {
    l_users, l_err := self.getOrgsUsers(p_orgs)
    if (l_err != nil) {
      return l_err
    }
    for _, c_el := range(l_users) {
      self.addUser(c_el.Guid)
    }
  }
  return nil
}

func (self *MessageReqCtx) addSpaces(p_spaces []string)  (error) {
  if (len(p_spaces) != 0) {
    l_users, l_err := self.getSpacesUsers(p_spaces)
    if (l_err != nil) {
      return l_err
    }
    for _, c_el := range(l_users) {
      self.addUser(c_el.Guid)
    }
  }
  return nil
}

func (self *MessageReqCtx) addUsers(p_users []string) {
  if (len(p_users) != 0) {
    for _, c_id := range(p_users) {
      self.addUser(c_id)
    }
  }
}


func (self *MessageReqCtx) addUser(p_guid string) {
  l_mail, l_ok := self.UserMails[p_guid]
  if (l_ok) {
    self.ResData = append(self.ResData, l_mail)
  }
}

func (self *MessageReqCtx) addBuidPack(p_guid string) (error) {
  // todo
  return nil
}

func (self *MessageReqCtx) addService(p_guid string) (error) {
  // todo
  return nil
}

func (self *MessageReqCtx) getOrgsUsers(p_list []string) (cfclient.Users, error) {
  l_orgs  := strings.Join(p_list, ",")
  l_query := url.Values{}
  l_query.Add("q", fmt.Sprintf("organization_guid IN %s", l_orgs))
  log.WithFields(log.Fields{"orgs" : l_orgs}).
    Debug("reading org users")
  l_users, l_err := self.CCCli.ListUsersByQuery(l_query)
  if (l_err != nil) {
    log.WithError(l_err).Error("unable to fetch users from CC api")
  }
  return l_users, l_err
}

func (self *MessageReqCtx) getSpacesUsers(p_list []string) (cfclient.Users, error) {
  l_spaces := strings.Join(p_list, ",")
  l_query := url.Values{}
  l_query.Add("q", fmt.Sprintf("space_guid IN %s", l_spaces))
  log.WithFields(log.Fields{"spaces" : l_spaces}).
    Debug("reading space users")
  l_users, l_err := self.CCCli.ListUsersByQuery(l_query)
  if (l_err != nil) {
    log.WithError(l_err).Error("unable to fetch users from CC api")
  }
  return l_users, l_err
}


// Local Variables:
// ispell-local-dictionary: "american"
// End:
