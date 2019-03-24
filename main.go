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
import "path/filepath"

var GApp App

type App struct {
	Config         core.AppConfig
	UiHandler      *ui.UiHandler
	ObjectHandler  *api.ObjectHandler
	MessageHandler *api.MessageHandler
	MailHandler    *mail.MailHandler
}

func NewApp(pRouter *mux.Router) *App {
	conf := core.NewAppConfig()
	objH := api.NewObjectHandler(&conf, pRouter)
	uiH := ui.NewUiHandler(&conf, pRouter)
	mailer, err := mail.NewMailHandler(&conf, pRouter)
	msgH, err := api.NewMessageHandler(&conf, pRouter, mailer.Queue)

	if err != nil {
		log.WithError(err).Error("failed to create api MessageHandler", err)
		os.Exit(1)
	}

	return &App{
		Config:         conf,
		ObjectHandler:  objH,
		UiHandler:      uiH,
		MessageHandler: msgH,
		MailHandler:    mailer,
	}
}

func (self *App) ListenAndServe(pRouter *mux.Router) {
	var err error

	if ("" != self.Config.HttpCert) && ("" != self.Config.HttpKey) {
		log.WithFields(log.Fields{
			"port": self.Config.HttpPort,
			"ssl":  true,
		}).Info("starting ssl web server")
		err = http.ListenAndServeTLS(
			fmt.Sprintf(":%d", self.Config.HttpPort),
			self.Config.HttpCert,
			self.Config.HttpKey,
			pRouter)
	} else {
		log.WithFields(log.Fields{
			"port": self.Config.HttpPort,
			"ssl":  false,
		}).Info("starting web server")
		err = http.ListenAndServe(fmt.Sprintf(":%d", self.Config.HttpPort), pRouter)
	}

	if err != nil {
		log.WithError(err).Error("unable to start web server")
		os.Exit(1)
	}
}

func main() {
	binDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	os.Chdir(binDir)

	router := mux.NewRouter()
	app := NewApp(router)

	app.MailHandler.Run()
	app.ListenAndServe(router)
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
