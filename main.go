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

func initLogger(pLevel string) {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	log.SetOutput(os.Stdout)
	lLevel, lErr := log.ParseLevel(pLevel)
	if lErr != nil {
		lLevel = log.ErrorLevel
	}
	log.SetLevel(lLevel)
}

func NewApp() App {
	lConf := NewAppConfig()
	lApp := App{
		Config: lConf,
		UaaCli: nil,
	}
	initLogger(lConf.LogLevel)
	lCli, lErr := NewUaaCli(&lConf)
	if lErr != nil {
		log.WithError(lErr).
			Error("unable to initialize uaa-client")
		os.Exit(1)
	}
	lApp.UaaCli = lCli
	return lApp
}


func main() {
	initLogger("warning")

	r := mux.NewRouter()

	GApp            = NewApp()
	GUiHandler      = NewUiHandler(r)
	GObjectHandler  = NewObjectHandler(r)
	GMessageHandler = NewMessageHandler(&GApp, r)

	r.Path("/").
		HandlerFunc(DecorateHandler(func(pRes http.ResponseWriter, pReq *http.Request) {
		http.Redirect(pRes, pReq, "/ui", http.StatusMovedPermanently)
	}))

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

// Local Variables:
// ispell-local-dictionary: "american"
// End:
