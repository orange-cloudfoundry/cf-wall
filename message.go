package main

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
import _ "github.com/cloudfoundry-community/gautocloud/connectors/smtp"
import "github.com/cloudfoundry-community/gautocloud/connectors/smtp/smtptype"
import "gopkg.in/gomail.v2"

type MessageReqCtx struct {
	CCCli     *cfclient.Client
	UserMails map[string]string
	ReqData   MessageRequest
	ResData   MessageResponse
}

type MessageHandler struct {
	UaaCli *UaaCli
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

func NewMessageHandler(p_app *App, p_router *mux.Router) MessageHandler {
	l_obj := MessageHandler{
		UaaCli: p_app.UaaCli,
	}

	p_router.Path("/v1/message").
		HandlerFunc(DecorateHandler(l_obj.HandleMessage)).
		HeadersRegexp("Content-Type", "application/json.*").
		Methods("POST")

	return l_obj
}

func (self *MessageHandler) createCtx(p_users map[string]string, p_req *http.Request) (*MessageReqCtx, error) {
	l_cccli, l_err := NewCCCliFromRequest(&GApp.Config, p_req)
	if l_err != nil {
		log.WithError(l_err).Error("unable to create CC client")
		return nil, l_err
	}

	l_ctx := MessageReqCtx{
		CCCli:     l_cccli,
		UserMails: p_users,
		ReqData:   MessageRequest{},
		ResData:   MessageResponse{},
	}

	l_decoder := json.NewDecoder(p_req.Body)
	l_err = l_decoder.Decode(&l_ctx.ReqData)
	if l_err != nil {
		return nil, l_err
	}

	return &l_ctx, nil
}

func (self *MessageHandler) getUaaUsers() (map[string]string, error) {
	l_res := make(map[string]string, 0)
	log.Debug("reading UAA users")
	l_users, l_err := self.UaaCli.GetUserList()
	if l_err != nil {
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
	if l_err != nil {
		panic(HttpError{l_err, 400, 10})
	}

	l_ctx, l_err := self.createCtx(l_users, p_req)
	if l_err != nil {
		panic(HttpError{l_err, 400, 10})
	}

	l_ctx.process()

	// self.sendMessage(
	// 	GApp.Config.MailFrom,
	// 	l_ctx.ResData.Emails,
	// 	l_ctx.ReqData.Subject,
	// 	l_ctx.ResData.Message)

	WriteJson(p_res, l_ctx.ResData)
}

func (self *MessageHandler) sendMessage(
	p_from string,
	p_to []string,
	p_subject string,
	p_content string) {

	var l_opts smtptype.Smtp
	l_err := gautocloud.Inject(&l_opts)
	if l_err != nil {
		l_uerr := errors.New("unable to get smtp settings")
		log.WithError(l_err).Error(l_uerr.Error())
		panic(HttpError{l_uerr, 500, 20})
	}

	l_dialer := gomail.NewPlainDialer(
		l_opts.Host,
		l_opts.Port,
		l_opts.User,
		l_opts.Password)

	l_msg := gomail.NewMessage()
	l_msg.SetHeader("From", p_from)
	l_msg.SetHeader("To", p_to...)
	l_msg.SetHeader("Subject", p_subject)
	l_msg.SetBody("text/html", p_content)

	l_err = l_dialer.DialAndSend(l_msg)
	if l_err != nil {
		l_uerr := errors.New("unable to send mail")
		log.WithError(l_err).Error(l_uerr.Error())
		panic(HttpError{l_uerr, 500, 20})
	}
}

func (self *MessageReqCtx) process() {
	self.addOrgs(self.ReqData.Orgs)
	self.addSpaces(self.ReqData.Spaces)
	self.addUsers(self.ReqData.Users)
	// TODO: addServices
	// TODO: addBuidPack

	l_mk   := markdown.New(markdown.XHTMLOutput(true), markdown.Nofollow(true))
	l_html := l_mk.RenderToString([]byte(self.ReqData.Message))
	self.ResData.Message = l_html
}

func (self *MessageReqCtx) addOrgs(p_orgs []string) {
	if len(p_orgs) == 0 {
		return
	}

	l_users := self.getOrgsUsers(p_orgs)
	for _, c_el := range l_users {
		self.addUser(c_el.Guid)
	}
}

func (self *MessageReqCtx) addSpaces(p_spaces []string) {
	if 0 == len(p_spaces) {
		return
	}

	l_users := self.getSpacesUsers(p_spaces)
	for _, c_el := range l_users {
		self.addUser(c_el.Guid)
	}
}

func (self *MessageReqCtx) addUsers(p_users []string) {
	if 0 == len(p_users) {
		return
	}
	for _, c_id := range p_users {
		self.addUser(c_id)
	}
}

func (self *MessageReqCtx) addUser(p_guid string) {
	l_mail, l_ok := self.UserMails[p_guid]
	if l_ok {
		self.ResData.Emails = append(self.ResData.Emails, l_mail)
	}
}

func (self *MessageReqCtx) addBuidPack(p_guid string) {
	// todo
}

func (self *MessageReqCtx) addService(p_guid string) {
	// todo
}

func (self *MessageReqCtx) getOrgsUsers(p_list []string) cfclient.Users {
	l_orgs := strings.Join(p_list, ",")
	l_query := url.Values{}
	l_query.Add("q", fmt.Sprintf("organization_guid IN %s", l_orgs))

	log.WithFields(log.Fields{"orgs": l_orgs}).
		Debug("reading org users")

	l_users, l_err := self.CCCli.ListUsersByQuery(l_query)
	if l_err != nil {
		l_uerr := errors.New("unable to fetch users from CC api")
		log.WithError(l_err).Error(l_uerr.Error())
		panic(HttpError{l_err, 500, 20})
	}

	return l_users
}

func (self *MessageReqCtx) getSpacesUsers(p_list []string) cfclient.Users {
	l_spaces := strings.Join(p_list, ",")
	l_query := url.Values{}
	l_query.Add("q", fmt.Sprintf("space_guid IN %s", l_spaces))

	log.WithFields(log.Fields{"spaces": l_spaces}).
		Debug("reading space users")

	l_users, l_err := self.CCCli.ListUsersByQuery(l_query)
	if l_err != nil {
		l_uerr := errors.New("unable to fetch users from CC api")
		log.WithError(l_err).Error(l_uerr.Error())
		panic(HttpError{l_err, 500, 20})
	}

	return l_users
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
