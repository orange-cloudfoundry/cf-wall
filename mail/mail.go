package mail
import "github.com/orange-cloudfoundry/cf-wall/core"
import "github.com/cloudfoundry-community/gautocloud"
import "github.com/cloudfoundry-community/gautocloud/connectors/smtp/raw"
import "github.com/cloudfoundry-community/gautocloud/connectors/smtp/smtptype"
import "github.com/gorilla/mux"
import "net/http"
import "gopkg.in/gomail.v2"
import "time"
import "errors"
import log "github.com/sirupsen/logrus"
import "fmt"

type MailHandler struct {
	config *core.AppConfig
	Queue  chan *gomail.Message
	opts   smtptype.Smtp
}

type StatusResponse struct {
	Outgoing int `json:"outgoing"`
}

func NewMailHandler(pConf *core.AppConfig, pRouter *mux.Router) (*MailHandler, error) {
	lObj := MailHandler{
		config: pConf,
		Queue: make(chan *gomail.Message, 5000),
	}

	pRouter.Path("/v1/mail/status").
		HandlerFunc(core.DecorateHandler(lObj.HandleMessage)).
		Methods("GET")

	lErr := gautocloud.Inject(&lObj.opts)
	if lErr != nil {
		lUerr := errors.New("unable to get smtp settings")
		log.WithError(lErr).Error(lUerr.Error())
		panic(core.NewHttpError(lUerr, 500, 52))
	}
	log.WithFields(log.Fields{"smtp": lObj.opts}).
		Debug("fetched settings from gautocloud")


	return &lObj, nil
}

func (self *MailHandler) send(pMsg *gomail.Message) {
	lDialer := gomail.NewPlainDialer(
		self.opts.Host,
		self.opts.Port,
		self.opts.User,
		self.opts.Password)
	lSender, lErr := lDialer.Dial()
	defer lSender.Close()

	if lErr != nil {
		lUerr := errors.New("could not connect mail server")
		log.WithError(lErr).Error(lUerr.Error())
	}

	log.WithFields(log.Fields{}).Debug("sending mail")
	if false == self.config.MailDry {
		lErr = gomail.Send(lSender, pMsg)
		if lErr != nil {
			lUerr := errors.New("could not send mail")
			log.WithError(lErr).Error(lUerr.Error())
		}
	}
}

func (self *MailHandler) checkInterval(pStart *time.Time, pCount *int) {
	lElapsed := time.Now().Sub(*pStart)
	lSize, _ := time.ParseDuration(fmt.Sprintf("%ds", self.config.MailRateDuration))

	// no event on last interval
	if lElapsed > lSize {
		log.WithFields(log.Fields{
			"elapsed(s)" : lElapsed.Seconds(),
			"size(s)"    : lSize.Seconds(),
			"count"      : *pCount,
		}).Debug("last interval too old, reset")
		*pStart = time.Now()
		*pCount = 0
	} else if *pCount < self.config.MailRateCount {
		log.WithFields(log.Fields{
			"elapsed(s)" : lElapsed.Seconds(),
			"size(s)"    : lSize.Seconds(),
			"count"      : *pCount,
		}).Debug("request within rate limit")
	} else {
		lDelay := lSize - lElapsed
		log.WithFields(log.Fields{
			"elapsed(s)" : lElapsed.Seconds(),
			"size(s)"    : lSize.Seconds(),
			"count"      : *pCount,
			"delay"      : lDelay.Seconds(),
		}).Debug("reached rate limit, sleeping")
		time.Sleep(lDelay)
		*pStart = time.Now()
		*pCount = 0
	}
}

func (self *MailHandler) run() {
	lStart := time.Now()
	lCount := 0

	for {
		lMsg := <-self.Queue
		self.checkInterval(&lStart, &lCount)
		self.send(lMsg)
		lCount += 1
	}
}

func (self *MailHandler) Run() {
	go self.run()
}

func (self *MailHandler) HandleMessage(pRes http.ResponseWriter, pReq *http.Request) {
	core.WriteJson(pRes, StatusResponse{ len(self.Queue) })
}

func init() {
	gautocloud.RegisterConnector(raw.NewSmtpRawConnector())
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
