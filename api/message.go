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
	CCCli     core.CFClient
	UserMails map[string]string
	ReqData   MessageRequest
	ResData   MessageResponse

	apps      []cfclient.App
	spaces    []string
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

	pRouter.Path("/v1/message_all").
		HandlerFunc(core.DecorateHandler(lObj.HandleMessageAll)).
		HeadersRegexp("Content-Type", "application/json.*").
		Methods("POST")

	return &lObj, nil
}

func (self *MessageHandler) createCtx(pUsers map[string]string, pReq *http.Request) (*MessageReqCtx, error) {
	lCccli, lErr := core.NewCCCliFromRequest(self.Config.CCEndPoint, pReq, self.Config.CCSkipVerify)
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
		panic(core.NewHttpError(lErr, 500, 51))
	}

	lCtx, lErr := self.createCtx(lUsers, pReq)
	if lErr != nil {
		panic(core.NewHttpError(lErr, 500, 50))
	}

	lCtx.process()

	if false == self.Config.MailDry {
		log.Debug("sending mail")
		// self.sendMessage(
		// 	self.Config.MailFrom,
		// 	lCtx.ResData.Emails,
		// 	lCtx.ReqData.Subject,
		// 	lCtx.ResData.Message)
	}

	core.WriteJson(pRes, lCtx.ResData)
}




func (self *MessageHandler) HandleMessageAll(pRes http.ResponseWriter, pReq *http.Request) {
	lUsers, lErr := self.getUaaUsers()
	if lErr != nil {
		panic(core.NewHttpError(lErr, 500, 51))
	}

	lCtx, lErr := self.createCtx(lUsers, pReq)
	if lErr != nil {
		panic(core.NewHttpError(lErr, 500, 50))
	}
	lCtx.addAllUsers()

	if false == self.Config.MailDry {
		log.Debug("sending mail")
		// self.sendMessage(
		// 	self.Config.MailFrom,
		// 	lCtx.ResData.Emails,
		// 	lCtx.ReqData.Subject,
		// 	lCtx.ResData.Message)
	}

	core.WriteJson(pRes, lCtx.ResData)
}

func (self *MessageHandler) sendMessage(
	pFrom    string,
	pTo      []string,
	pSubject string,
	pContent string) {

	var lOpts smtptype.Smtp
	lErr := gautocloud.Inject(&lOpts)
	if lErr != nil {
		lUerr := errors.New("unable to get smtp settings")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lUerr, 500, 52))
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
		panic(core.NewHttpError(lUerr, 500, 53))
	}
}




func (self *MessageReqCtx) process() {
	self.addOrgs(self.ReqData.Orgs)

	for _, cId := range self.ReqData.Spaces {
		self.addSpace(cId)
	}

	self.addUsers(self.ReqData.Users)
	self.addBuidPacks(self.ReqData.BuildPacks)
	self.addServices(self.ReqData.Services)

	self.readSpaces()

	lMk   := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	lHtml := lMk.RenderToString([]byte(self.ReqData.Message))
	self.ResData.Message = lHtml
}





func (self *MessageReqCtx) addBuidPacks(pBps []string) {
	if len(pBps) == 0 {
		return
	}

	log.WithFields(log.Fields{ "buildpacks": pBps })
	// 1. build map for more efficient search
	lNeedles := map[string]bool{}
	for _, cBp := range pBps {
		lNeedles[cBp] = true
	}

	// 2. build targeted spaces from application buildpacks
	self.mapApps(func (pApp *cfclient.App) {
		if "" != pApp.DetectedBuildpackGuid {
			_, lOk := lNeedles[pApp.DetectedBuildpackGuid]
			if lOk {
				self.addSpace(pApp.SpaceGuid)
			}
		}
	});
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

func (self *MessageReqCtx) readSpaces() {
	if 0 == len(self.spaces) {
		return
	}
	lUsers := self.getSpacesUsers(self.spaces)
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


func (self *MessageReqCtx) addAllUsers() {
	for _, cMail := range self.UserMails {
		self.ResData.Emails = append(self.ResData.Emails, cMail)
	}
}


func (self *MessageReqCtx) addServices(pGuids []string) {
	if 0 == len(pGuids) {
		return
	}

	// 1. service list to guid map
	lServices := map[string]bool{}
	for _, cId := range pGuids {
		lServices[cId] = true
	}

	// 2. search instances matching services
	lUsedInst := make([]string, 0)
	lInstances := self.getServiceInstances()
	for _, cInst := range lInstances {
		_, lOk := lServices[cInst.ServiceGuid]
		if lOk {

			self.addSpace(cInst.SpaceGuid)
			lUsedInst = append(lUsedInst, cInst.Guid)
		}
	}

	// 3. get bindings from instances
	lBindings := self.getServiceBindings(lUsedInst)

	// 4. browse bindings to get application spaces
	for _, cBind := range lBindings {
		self.mapApps(func (pApp *cfclient.App) {
			if pApp.Guid == cBind.AppGuid {
				self.addSpace(pApp.SpaceGuid)
			}
		})
	}
}

func (self *MessageReqCtx) addSpace(pId string) {
	self.spaces = append(self.spaces, pId)
}

func (self *MessageReqCtx) mapApps(pMap func (*cfclient.App)) {
	if 0 == len(self.apps) {
		self.apps = self.getApps()
	}
	for _, cApp := range self.apps {
		pMap(&cApp)
	}
}

func (self *MessageReqCtx) getServiceBindings(pList []string) ([]cfclient.ServiceBinding) {
	log.Debug("reading service bindings")

	lQuery := url.Values{}
	lQuery.Set("results-per-page", "100")
	lQuery.Add("q", fmt.Sprintf("service_instance_guid IN %s", strings.Join(pList, ",")))
	lRes, lErr := self.CCCli.ListServiceBindingsByQuery(lQuery)
	if lErr != nil {
		lUerr := errors.New("unable to fetch service bindings from CC api")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lErr, 500, 50))
	}
	log.WithFields(log.Fields{"count": len(lRes)}).
		Debug("fetched service bindings")

	return lRes
}


func (self *MessageReqCtx) getServiceInstances() ([]cfclient.ServiceInstance) {
	log.Debug("reading service instances")

	lQuery := url.Values{}
	lQuery.Set("results-per-page", "100")
	lRes, lErr := self.CCCli.ListServiceInstancesByQuery(lQuery)
	if lErr != nil {
		lUerr := errors.New("unable to fetch service instances from CC api")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lErr, 500, 50))
	}
	log.WithFields(log.Fields{"count": len(lRes)}).
		Debug("fetched service instances")

	return lRes
}


func (self *MessageReqCtx) getApps() ([]cfclient.App) {
	log.Debug("reading applications")

	lQuery := url.Values{}
	lQuery.Set("results-per-page", "100")
	lApps, lErr := self.CCCli.ListAppsByQuery(lQuery)
	if lErr != nil {
		lUerr := errors.New("unable to fetch applications from CC api")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lErr, 500, 50))
	}
	log.WithFields(log.Fields{"count": len(lApps)}).
		Debug("fetched applications")

	return lApps
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

	log.WithFields(log.Fields{"count": len(lUsers)}).
		Debug("fetched users")
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

	log.WithFields(log.Fields{"count": len(lUsers)}).
		Debug("fetched users")
	return lUsers
}


func init() {
	gautocloud.RegisterConnector(raw.NewSmtpRawConnector())
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
