package api

import "fmt"
import "strings"
import "encoding/json"
import "errors"
import "net/http"
import "net/url"
import "github.com/cloudfoundry-community/go-cfclient"
import "github.com/gorilla/mux"
import "github.com/golang-commonmark/markdown"
import log "github.com/sirupsen/logrus"
import "github.com/cloudfoundry-community/gautocloud"
import "github.com/cloudfoundry-community/gautocloud/connectors/smtp/raw"
import "github.com/cloudfoundry-community/gautocloud/connectors/smtp/smtptype"
import "gopkg.in/gomail.v2"
import "github.com/orange-cloudfoundry/cf-wall/core"


type MessageReqCtx struct {
	CCCli     *cfclient.Client
	UserMails map[string]string
	ReqData   MessageRequest
	ResData   MessageResponse
}

type MessageHandler struct {
	UaaCli *core.UaaCli
	Config *core.AppConfig
}

type MessageRequest struct {
	Users      []string `json:"users"`
	Spaces     []string `json:"spaces"`
	Orgs       []string `json:"orgs"`
	Services   []string `json:"services"`
	BuildPacks []string `json:"buildpacks"`
	Subject    string   `json:"subject"`
	Message    string   `json:"message"`
}

type MessageResponse struct {
	Emails  []string `json:"emails"`
	Message string   `json:"message"`
}

func NewMessageHandler(pConf *core.AppConfig, pRouter *mux.Router) (*MessageHandler, error) {
	lCli, lErr := core.NewUaaCli(pConf)
	if lErr != nil {
		log.WithError(lErr).Error("failed to create core UaaClient", lErr)
		return nil, lErr
	}

	lObj := MessageHandler{
		UaaCli: lCli,
		Config: pConf,
	}

	pRouter.Path("/v1/message").
		HandlerFunc(core.DecorateHandler(lObj.HandleMessage)).
		HeadersRegexp("Content-Type", "application/json.*").
		Methods("POST")

	return &lObj, nil
}

func (self *MessageHandler) createCtx(pUsers map[string]string, pReq *http.Request) (*MessageReqCtx, error) {
	lCccli, lErr := core.NewCCCliFromRequest(self.Config.CCEndPoint, pReq)
	if lErr != nil {
		log.WithError(lErr).Error("unable to create CC client")
		return nil, lErr
	}

	lCtx := MessageReqCtx{
		CCCli:     lCccli,
		UserMails: pUsers,
		ReqData:   MessageRequest{},
		ResData:   MessageResponse{},
	}

	lDecoder := json.NewDecoder(pReq.Body)
	lErr = lDecoder.Decode(&lCtx.ReqData)
	if lErr != nil {
		return nil, lErr
	}

	return &lCtx, nil
}

func (self *MessageHandler) getUaaUsers() (map[string]string, error) {
	lRes := make(map[string]string, 0)
	log.Debug("reading UAA users")
	lUsers, lErr := self.UaaCli.GetUserList()
	if lErr != nil {
		log.WithError(lErr).Error("unable to featch UAA users")
		return lRes, lErr
	}
	for _, cEl := range lUsers {
		lRes[cEl.Id] = cEl.Email
	}
	return lRes, nil
}

func (self *MessageHandler) HandleMessage(pRes http.ResponseWriter, pReq *http.Request) {
	lUsers, lErr := self.getUaaUsers()
	if lErr != nil {
		panic(core.NewHttpError(lErr, 400, 10))
	}

	lCtx, lErr := self.createCtx(lUsers, pReq)
	if lErr != nil {
		panic(core.NewHttpError(lErr, 400, 10))
	}

	lCtx.process()

	self.sendMessage(
		self.Config.MailFrom,
		lCtx.ResData.Emails,
		lCtx.ReqData.Subject,
		lCtx.ResData.Message)

	core.WriteJson(pRes, lCtx.ResData)
}

func (self *MessageHandler) sendMessage(
	pFrom string,
	pTo []string,
	pSubject string,
	pContent string) {

	var lOpts smtptype.Smtp
	lErr := gautocloud.Inject(&lOpts)
	if lErr != nil {
		lUerr := errors.New("unable to get smtp settings")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lUerr, 500, 20))
	}
	log.WithFields(log.Fields{"smtp": lOpts}).
		Debug("fetched settings from gautocloud")

	lDialer := gomail.NewPlainDialer(
		lOpts.Host,
		lOpts.Port,
		lOpts.User,
		lOpts.Password)

	lMsg := gomail.NewMessage()
	lMsg.SetHeader("From", pFrom)
	lMsg.SetHeader("To", pTo...)
	lMsg.SetHeader("Subject", pSubject)
	lMsg.SetBody("text/html", pContent)

	lErr = lDialer.DialAndSend(lMsg)
	if lErr != nil {
		lUerr := errors.New("unable to send mail")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lUerr, 500, 20))
	}
}

func (self *MessageReqCtx) process() {
	self.addOrgs(self.ReqData.Orgs)
	self.addSpaces(self.ReqData.Spaces)
	self.addUsers(self.ReqData.Users)
	// TODO: addServices
	// TODO: addBuidPack

	lMk   := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	lHtml := lMk.RenderToString([]byte(self.ReqData.Message))
	self.ResData.Message = lHtml
}

func (self *MessageReqCtx) addOrgs(pOrgs []string) {
	if len(pOrgs) == 0 {
		return
	}

	lUsers := self.getOrgsUsers(pOrgs)
	for _, cEl := range lUsers {
		self.addUser(cEl.Guid)
	}
}

func (self *MessageReqCtx) addSpaces(pSpaces []string) {
	if 0 == len(pSpaces) {
		return
	}

	lUsers := self.getSpacesUsers(pSpaces)
	for _, cEl := range lUsers {
		self.addUser(cEl.Guid)
	}
}

func (self *MessageReqCtx) addUsers(pUsers []string) {
	if 0 == len(pUsers) {
		return
	}
	for _, cId := range pUsers {
		self.addUser(cId)
	}
}

func (self *MessageReqCtx) addUser(pGuid string) {
	lMail, lOk := self.UserMails[pGuid]
	if lOk {
		self.ResData.Emails = append(self.ResData.Emails, lMail)
	}
}

func (self *MessageReqCtx) addBuidPack(pGuid string) {
	// todo
}

func (self *MessageReqCtx) addService(pGuid string) {
	// todo
}

func (self *MessageReqCtx) getOrgsUsers(pList []string) cfclient.Users {
	lOrgs := strings.Join(pList, ",")
	lQuery := url.Values{}
	lQuery.Add("q", fmt.Sprintf("organization_guid IN %s", lOrgs))

	log.WithFields(log.Fields{"orgs": lOrgs}).
		Debug("reading org users")

	lUsers, lErr := self.CCCli.ListUsersByQuery(lQuery)
	if lErr != nil {
		lUerr := errors.New("unable to fetch users from CC api")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lErr, 500, 20))
	}

	return lUsers
}

func (self *MessageReqCtx) getSpacesUsers(pList []string) cfclient.Users {
	lSpaces := strings.Join(pList, ",")
	lQuery := url.Values{}
	lQuery.Add("q", fmt.Sprintf("space_guid IN %s", lSpaces))

	log.WithFields(log.Fields{"spaces": lSpaces}).
		Debug("reading space users")

	lUsers, lErr := self.CCCli.ListUsersByQuery(lQuery)
	if lErr != nil {
		lUerr := errors.New("unable to fetch users from CC api")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lErr, 500, 20))
	}

	return lUsers
}


func init() {
	gautocloud.RegisterConnector(raw.NewSmtpRawConnector())
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
