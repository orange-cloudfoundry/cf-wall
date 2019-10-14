package api

import "fmt"
import "strings"
import "encoding/json"
import "errors"
import "net/http"
import "net/url"
import "net/mail"
import "github.com/cloudfoundry-community/go-cfclient"
import "github.com/gorilla/mux"
import "github.com/golang-commonmark/markdown"
import log "github.com/sirupsen/logrus"
import "gopkg.in/gomail.v2"
import "github.com/orange-cloudfoundry/cf-wall/core"
import "sync"

//MessageReqCtx --
type MessageReqCtx struct {
	CCCli          core.CFClient
	UserMails      map[string]string
	ReqData        MessageRequest
	ResData        MessageResponse
	NbMaxGetParams int

	apps   []cfclient.App
	spaces []string
}

//MessageHandler --
type MessageHandler struct {
	UaaCli *core.UaaCli
	Config *core.AppConfig
	queue  chan *gomail.Message
}

// RecipientsRequest --
type RecipientsRequest struct {
	Users      []string `json:"users"`
	Spaces     []string `json:"spaces"`
	Orgs       []string `json:"orgs"`
	Services   []string `json:"services"`
	BuildPacks []string `json:"buildpacks"`
	Recipients []string `json:"recipients"`
}

//MessageRequest --
type MessageRequest struct {
	RecipientsRequest
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// RecipientsResponse --
type RecipientsResponse struct {
	Recipients []string `json:"recipients"`
}

//MessageResponse --
type MessageResponse struct {
	RecipientsResponse
	Subject string `json:"subject"`
	Message string `json:"message"`
	From    string `json:"from"`
}

// NewMessageHandler --
func NewMessageHandler(
	pConf *core.AppConfig,
	pRouter *mux.Router,
	pQueue chan *gomail.Message) (*MessageHandler, error) {

	cli, err := core.NewUaaCli(pConf)
	if err != nil {
		log.WithError(err).Error("failed to create core UaaClient", err)
		return nil, err
	}

	obj := MessageHandler{
		UaaCli: cli,
		Config: pConf,
		queue:  pQueue,
	}

	pRouter.Path("/v1/message").
		HandlerFunc(core.DecorateHandler(obj.handleMessage)).
		HeadersRegexp("Content-Type", "application/json.*").
		Methods("POST")

	pRouter.Path("/v1/recipients").
		HandlerFunc(core.DecorateHandler(obj.handleRecipients)).
		HeadersRegexp("Content-Type", "application/json.*").
		Methods("POST")

	pRouter.Path("/v1/message_all").
		HandlerFunc(core.DecorateHandler(obj.handleMessageAll)).
		HeadersRegexp("Content-Type", "application/json.*").
		Methods("POST")

	return &obj, nil
}

func (m *MessageHandler) createCtx(pUsers map[string]string, pReq *http.Request) (*MessageReqCtx, error) {
	cccli, err := core.NewCCCliFromRequest(m.Config.CCEndPoint, pReq, m.Config.CCSkipVerify)
	if err != nil {
		log.WithError(err).Error("unable to create CC client")
		return nil, err
	}

	ctx := MessageReqCtx{
		CCCli:           cccli,
		UserMails:       pUsers,
		NbMaxGetParams : m.Config.NbMaxGetParams,
		ReqData:         MessageRequest{},
		ResData:         MessageResponse{},
	}

	decoder := json.NewDecoder(pReq.Body)
	err = decoder.Decode(&ctx.ReqData)
	if err != nil {
		return nil, err
	}

	ctx.setFrom(m.Config.MailFrom)
	ctx.setSubject(ctx.ReqData.Subject, m.Config.MailTag)
	ctx.addRecipents(m.Config.MailCc)
	ctx.addRecipents(ctx.ReqData.Recipients)
	ctx.setBody(ctx.ReqData.Message)
	return &ctx, nil
}

func (m *MessageHandler) getUaaUsers() (map[string]string, error) {
	res := make(map[string]string, 0)
	log.Debug("reading UAA users")
	users, err := m.UaaCli.GetUserList()

	if err != nil {
		log.WithError(err).Error("unable to featch UAA users")
		return res, err
	}
	for _, cEl := range users {
		_, err := mail.ParseAddress(cEl.Email)
		if err == nil {
			res[cEl.Id] = cEl.Email
		}
	}
	return res, nil
}

func (m *MessageHandler) getRecipients(pReq *http.Request) (*MessageResponse, error) {
	users, err := m.getUaaUsers()
	if err != nil {
		return nil, err
	}

	ctx, err := m.createCtx(users, pReq)
	if err != nil {
		return nil, err
	}

	ctx.addOrgs(ctx.ReqData.Orgs)
	ctx.addSpaces(ctx.ReqData.Spaces)
	ctx.addBuidPacks(ctx.ReqData.BuildPacks)
	ctx.addServices(ctx.ReqData.Services)
	ctx.addUsers(ctx.ReqData.Users)
	ctx.readSpaces()

	return &ctx.ResData, nil
}

func (m *MessageHandler) handleMessage(pRes http.ResponseWriter, pReq *http.Request) {
	data, err := m.getRecipients(pReq)
	if err != nil {
		panic(core.NewHttpError(err, 500, 51))
	}

	m.sendMessages(
		data.From,
		data.Recipients,
		data.Subject,
		data.Message)

	pRes.WriteHeader(204)
}

func (m *MessageHandler) handleRecipients(pRes http.ResponseWriter, pReq *http.Request) {
	data, err := m.getRecipients(pReq)
	if err != nil {
		panic(core.NewHttpError(err, 500, 51))
	}
	core.WriteJson(pRes, data.RecipientsResponse)
}

func (m *MessageHandler) handleMessageAll(pRes http.ResponseWriter, pReq *http.Request) {
	users, err := m.getUaaUsers()
	if err != nil {
		panic(core.NewHttpError(err, 500, 51))
	}

	ctx, err := m.createCtx(users, pReq)
	if err != nil {
		panic(core.NewHttpError(err, 500, 50))
	}
	ctx.addAlusers()

	m.sendMessages(
		ctx.ResData.From,
		ctx.ResData.Recipients,
		ctx.ResData.Subject,
		ctx.ResData.Message)

	pRes.WriteHeader(204)
	//core.WriteJson(pRes, ctx.ResData)
}

func (m *MessageHandler) sendMessages(
	pFrom string,
	pTo []string,
	pSub string,
	pBody string) {

	for _, cDest := range pTo {
		msg := gomail.NewMessage()
		msg.SetHeader("From", pFrom)
		msg.SetHeader("To", cDest)
		msg.SetHeader("Subject", pSub)
		msg.SetBody("text/html", pBody)
		m.queue <- msg
	}
}

func (m *MessageReqCtx) setFrom(pFrom string) {
	m.ResData.From = pFrom
}

func (m *MessageReqCtx) setSubject(pSub string, pTag string) {
	m.ResData.Subject = pSub
	if len(pTag) != 0 {
		m.ResData.Subject = fmt.Sprintf("%s %s", pTag, pSub)
	}
}

func (m *MessageReqCtx) addRecipents(pList []string) {
	for _, cItem := range pList {
		_, err := mail.ParseAddress(cItem)
		if err != nil {
			uerr := fmt.Errorf("invalid email address '%s'", cItem)
			log.WithError(err).Error(uerr.Error())
			panic(core.NewHttpError(uerr, 500, 51))
		}
		m.ResData.Recipients = append(m.ResData.Recipients, cItem)
	}
}

func (m *MessageReqCtx) setBody(pMarkdown string) {
	mk := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	html := mk.RenderToString([]byte(pMarkdown))
	m.ResData.Message = html
}

func (m *MessageReqCtx) addSpaces(pSpaces []string) {
	for _, cID := range pSpaces {
		m.addSpace(cID)
	}
}

func (m *MessageReqCtx) addBuidPacks(pBps []string) {
	if len(pBps) == 0 {
		return
	}

	log.WithFields(log.Fields{"buildpacks": pBps})
	// 1. build map for more efficient search
	needles := map[string]bool{}
	for _, cBp := range pBps {
		needles[cBp] = true
	}

	// 2. build targeted spaces from application buildpacks
	m.mapApps(func(pApp *cfclient.App) {
		if "" != pApp.DetectedBuildpackGuid {
			_, ok := needles[pApp.DetectedBuildpackGuid]
			if ok {
				m.addSpace(pApp.SpaceGuid)
			}
		}
	})
}

func (m *MessageReqCtx) addOrgs(pOrgs []string) {
	if len(pOrgs) == 0 {
		return
	}

	users := m.getOrgsUsers(pOrgs)
	for _, cEl := range users {
		m.addUser(cEl.Guid)
	}
}

func (m *MessageReqCtx) readSpaces() {
	if 0 == len(m.spaces) {
		return
	}
	users := m.getSpacesUsers(m.spaces)
	for _, cEl := range users {
		m.addUser(cEl.Guid)
	}
}

func (m *MessageReqCtx) addUsers(pUsers []string) {
	if 0 == len(pUsers) {
		return
	}
	for _, cID := range pUsers {
		m.addUser(cID)
	}
}

func (m *MessageReqCtx) addUser(pGUID string) {
	mail, ok := m.UserMails[pGUID]
	if ok {
		m.ResData.Recipients = append(m.ResData.Recipients, mail)
	}
}

func (m *MessageReqCtx) addAlusers() {
	for _, cMail := range m.UserMails {
		m.ResData.Recipients = append(m.ResData.Recipients, cMail)
	}
}

func (m *MessageReqCtx) addServices(pGuids []string) {
	if 0 == len(pGuids) {
		return
	}

	// 1. service list to guid map
	services := map[string]bool{}
	for _, cID := range pGuids {
		services[cID] = true
	}

	// 2. search instances matching services
	usedInst := make([]string, 0)
	instances := m.getServiceInstances()
	for _, cInst := range instances {
		_, ok := services[cInst.ServiceGuid]
		if ok {

			m.addSpace(cInst.SpaceGuid)
			usedInst = append(usedInst, cInst.Guid)
		}
	}

	// 3. get bindings from instances
	bindings := m.getServiceBindings(usedInst)

	// 4. browse bindings to get application spaces
	for _, cBind := range bindings {
		m.mapApps(func(pApp *cfclient.App) {
			if pApp.Guid == cBind.AppGuid {
				m.addSpace(pApp.SpaceGuid)
			}
		})
	}
}

func (m *MessageReqCtx) addSpace(pID string) {
	m.spaces = append(m.spaces, pID)
}

func (m *MessageReqCtx) mapApps(pMap func(*cfclient.App)) {
	if 0 == len(m.apps) {
		m.apps = m.getApps()
	}
	for _, cApp := range m.apps {
		pMap(&cApp)
	}
}

func (m *MessageReqCtx) getServiceBindings(pList []string) []cfclient.ServiceBinding {
	log.Debug("reading service bindings")

	query := url.Values{}
	query.Set("results-per-page", "100")
	query.Add("q", fmt.Sprintf("service_instance_guid IN %s", strings.Join(pList, ",")))
	res, err := m.CCCli.ListServiceBindingsByQuery(query)
	if err != nil {
		uerr := errors.New("unable to fetch service bindings from CC api")
		log.WithError(err).Error(uerr.Error())
		panic(core.NewHttpError(err, 500, 50))
	}
	log.WithFields(log.Fields{"count": len(res)}).
		Debug("fetched service bindings")

	return res
}

func (m *MessageReqCtx) getServiceInstances() []cfclient.ServiceInstance {
	log.Debug("reading service instances")

	query := url.Values{}
	query.Set("results-per-page", "100")
	res, err := m.CCCli.ListServiceInstancesByQuery(query)
	if err != nil {
		uerr := errors.New("unable to fetch service instances from CC api")
		log.WithError(err).Error(uerr.Error())
		panic(core.NewHttpError(err, 500, 50))
	}
	log.WithFields(log.Fields{"count": len(res)}).
		Debug("fetched service instances")

	return res
}

func (m *MessageReqCtx) getApps() []cfclient.App {
	log.Debug("reading applications")

	query := url.Values{}
	query.Set("results-per-page", "100")
	apps, err := m.CCCli.ListAppsByQuery(query)
	if err != nil {
		uerr := errors.New("unable to fetch applications from CC api")
		log.WithError(err).Error(uerr.Error())
		panic(core.NewHttpError(err, 500, 50))
	}
	log.WithFields(log.Fields{"count": len(apps)}).
		Debug("fetched applications")

	return apps
}

func (m *MessageReqCtx) getOrgsUsers(pList []string) cfclient.Users {
	lNbElements := len(pList)
	var orgs = ""
	var lUsers []cfclient.User
	var wg sync.WaitGroup

	queue := make(chan []cfclient.User, (lNbElements/m.NbMaxGetParams)+1)

	for j := 0; j < lNbElements; j += m.NbMaxGetParams {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			
			if lNbElements < i+50 {
				orgs = strings.Join(pList[i:lNbElements], ",")
			} else {
				orgs = strings.Join(pList[i:i+50], ",")
			}

			query := url.Values{}
			query.Add("q", fmt.Sprintf("organization_guid IN %s", orgs))

			log.WithFields(log.Fields{"orgs": orgs}).
				Debug("reading org users")

			users, err := m.CCCli.ListUsersByQuery(query)
			if err != nil {
				uerr := errors.New("unable to fetch users from CC api")
				log.WithError(err).Error(uerr.Error())
				panic(core.NewHttpError(err, 500, 20))
			}
			queue <- users
		}(j)
	}
	wg.Wait()
	close(queue)

	for users := range queue {
		lUsers = append(lUsers, users...)
	}
	return lUsers
}

func (m *MessageReqCtx) getSpacesUsers(pList []string) cfclient.Users {
	lNbElements := len(pList)
	var spaces = ""
	var lUsers []cfclient.User
	var wg sync.WaitGroup

	queue := make(chan []cfclient.User, (lNbElements/m.NbMaxGetParams)+1)

	for j := 0; j < lNbElements; j += m.NbMaxGetParams {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			
			if lNbElements < i+50 {
				spaces = strings.Join(pList[i:lNbElements], ",")
			} else {
				spaces = strings.Join(pList[i:i+50], ",")
			}

			query := url.Values{}
			query.Add("q", fmt.Sprintf("space_guid IN %s", spaces))

			log.WithFields(log.Fields{"spaces": spaces}).
				Debug("reading space users")

			users, err := m.CCCli.ListUsersByQuery(query)
			if err != nil {
				uerr := errors.New("unable to fetch users from CC api")
				log.WithError(err).Error(uerr.Error())
				panic(core.NewHttpError(err, 500, 20))
			}
			queue <- users
		}(j)
	}
	wg.Wait()
	close(queue)

	for users := range queue {
		lUsers = append(lUsers, users...)
	}
	return lUsers
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
