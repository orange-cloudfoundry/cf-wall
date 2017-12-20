package main

import "fmt"
import "os"
import "net/http"
import "github.com/gorilla/mux"
import log "github.com/sirupsen/logrus"
import "github.com/orange-cloudfoundry/cf-wall/core"
import "github.com/orange-cloudfoundry/cf-wall/api"
import "github.com/orange-cloudfoundry/cf-wall/ui"
import "github.com/orange-cloudfoundry/cf-wall/mail"

var GApp App

type App struct {
	Config         core.AppConfig
	UiHandler      *ui.UiHandler
	ObjectHandler  *api.ObjectHandler
	MessageHandler *api.MessageHandler
	MailHandler    *mail.MailHandler
}

func NewApp(pRouter *mux.Router) *App {
	lConf 				:= core.NewAppConfig()
	lObjH 				:= api.NewObjectHandler(&lConf, pRouter)
	lUiH          := ui.NewUiHandler(&lConf, pRouter)
	lMailer, lErr := mail.NewMailHandler(&lConf, pRouter)
	lMsgH,   lErr := api.NewMessageHandler(&lConf, pRouter, lMailer.Queue)

	if lErr != nil {
		log.WithError(lErr).Error("failed to create api MessageHandler", lErr)
		os.Exit(1)
	}

	return &App{
		Config:         lConf,
		ObjectHandler:  lObjH,
		UiHandler:      lUiH,
		MessageHandler: lMsgH,
		MailHandler:    lMailer,
	}
}

func (self *App) ListenAndServe(pRouter *mux.Router) {
	var lErr error

	if ("" != self.Config.HttpCert) && ("" != self.Config.HttpKey) {
		log.WithFields(log.Fields{
			"port": self.Config.HttpPort,
			"ssl":  true,
		}).Info("starting ssl web server")
		lErr = http.ListenAndServeTLS(
			fmt.Sprintf(":%d", self.Config.HttpPort),
			self.Config.HttpCert,
			self.Config.HttpKey,
			pRouter)
	} else {
		log.WithFields(log.Fields{
			"port": self.Config.HttpPort,
			"ssl":  false,
		}).Info("starting web server")
		lErr = http.ListenAndServe(fmt.Sprintf(":%d", self.Config.HttpPort), pRouter)
	}

	if lErr != nil {
		log.WithError(lErr).Error("unable to start web server")
		os.Exit(1)
	}
}


func main() {
	lRouter := mux.NewRouter()
	lApp    := NewApp(lRouter)

	lApp.MailHandler.Run()
	lApp.ListenAndServe(lRouter)
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
