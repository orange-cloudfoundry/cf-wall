package main

import "fmt"
import "os"
import "net/http"
import "github.com/gorilla/mux"
import log "github.com/sirupsen/logrus"

var GApp App
var GUiHandler UiHandler
var GObjectHandler ObjectHandler
var GMessageHandler MessageHandler

type App struct {
	Config AppConfig
	UaaCli *UaaCli
}

func (self *App) initLogger() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(os.Stdout)
	l_level, l_err := log.ParseLevel(self.Config.LogLevel)
	if l_err != nil {
		l_level = log.ErrorLevel
	}
	log.SetLevel(l_level)
}

func NewApp() App {
	l_conf := NewAppConfig()
	l_app := App{
		Config: l_conf,
		UaaCli: nil,
	}
	l_app.initLogger()
	l_cli, l_err := NewUaaCli(&l_conf)
	if l_err != nil {
		log.WithError(l_err).
			Error("unable to initialize uaa-client")
		os.Exit(1)
	}
	l_app.UaaCli = l_cli
	return l_app
}


func main() {
	r := mux.NewRouter()

	GApp            = NewApp()
	GUiHandler      = NewUiHandler(r)
	GObjectHandler  = NewObjectHandler(r)
	GMessageHandler = NewMessageHandler(&GApp, r)

	if ("" != GApp.Config.HttpCert) && ("" != GApp.Config.HttpKey) {
		log.WithFields(log.Fields{
			"port": GApp.Config.HttpPort,
			"ssl":  true,
		}).Info("starting ssl web server")
		http.ListenAndServeTLS(
			fmt.Sprintf(":%d", GApp.Config.HttpPort),
			GApp.Config.HttpCert,
			GApp.Config.HttpKey,
			r)
	} else {
		log.WithFields(log.Fields{
			"port": GApp.Config.HttpPort,
			"ssl":  false,
		}).Info("starting web server")
		http.ListenAndServe(fmt.Sprintf(":%d", GApp.Config.HttpPort), r)
	}
}
