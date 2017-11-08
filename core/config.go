package core

import     "flag"
import     "os"
import     "fmt"
import     "strconv"
import     "reflect"
import     "errors"
import     "encoding/json"
import log "github.com/sirupsen/logrus"
import     "github.com/cloudfoundry-community/gautocloud"
import     "github.com/cloudfoundry-community/gautocloud/connectors/generic"

type AppConfig struct {
	ConfigFile      string
	UaaClientName   string  `json:"uaa-client"        cloud:"uaa-client"`
	UaaClientSecret string  `json:"uaa-secret"        cloud:"uaa-secret"`
	UaaEndPoint     string  `json:"uaa-url"           cloud:"uaa-url"`
	UaaSkipVerify   bool    `json:"uaa-skip-verify"   cloud:"uaa-skip-verify"`
	CCEndPoint      string  `json:"cc-url"            cloud:"cc-url"`
	CCSkipVerify    bool    `json:"cc-skip-verify"    cloud:"cc-skip-verify"`
	HttpCert        string  `json:"http-cert"         cloud:"http-cert"`
	HttpKey         string  `json:"http-key"          cloud:"http-key"`
	HttpPort        int     `json:"http-port"         cloud:"http-port"`
	LogLevel        string  `json:"log-level"         cloud:"log-level"`
	MailFrom        string  `json:"mail-from"         cloud:"mail-from"`
	MailDry         bool    `json:"mail-dry"          cloud:"mail-dry"`
	ReloadTemplates bool    `json:"reload-templates"  cloud:"reload-templates"`
}

func InitLogger(pLevel string) {
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

func NewAppConfig() AppConfig {
	lConf := AppConfig{
		ConfigFile:      "",
		UaaClientName:   "cf-wall",
		UaaClientSecret: "password",
		UaaSkipVerify:   false,
		UaaEndPoint:     "https://uaa.example.com",
		CCEndPoint:      "https://api.example.com",
		CCSkipVerify:    false,
		HttpCert:        "",
		HttpKey:         "",
		HttpPort:        80,
		LogLevel:        "error",
		MailFrom:        "cf-wall@localhost",
		MailDry:         false,
		ReloadTemplates: false,
	}

	InitLogger("debug")
	lConf.parseArgs()
	InitLogger(lConf.LogLevel)
	return lConf
}


func (self *AppConfig) parseConfig() {
	lFile, lErr := os.Open(self.ConfigFile)
	if lErr != nil {
		fmt.Printf("unable to read configuration file '%s'", self.ConfigFile)
		os.Exit(1)
	}

	lDecoder := json.NewDecoder(lFile)
	lErr = lDecoder.Decode(self)
	if lErr != nil {
		fmt.Printf("unable to parse file '%s' : %s", self.ConfigFile, lErr.Error())
		os.Exit(1)
	}
}


func (self *AppConfig) parseCmdLine() {
	flag.StringVar (&self.ConfigFile,      "config",            self.ConfigFile,      "configuration file")
	flag.StringVar (&self.UaaClientName,   "uaa-client",        self.UaaClientName,   "UAA client ID")
	flag.StringVar (&self.UaaClientName,   "uaa-secret",        self.UaaClientName,   "UAA client secret")
	flag.StringVar (&self.UaaEndPoint,     "uaa-url",           self.UaaEndPoint,     "UAA API endpoint url")
	flag.BoolVar   (&self.UaaSkipVerify,   "uaa-skip-verify",   self.UaaSkipVerify,   "Do not verify UAA SSL certificates")
	flag.StringVar (&self.CCEndPoint,      "cc-url",            self.CCEndPoint,      "Cloud Controller API endpoint url")
	flag.BoolVar   (&self.CCSkipVerify,    "cc-skip-verify",    self.CCSkipVerify,    "Do not verify Cloud Controller SSL certificates")
	flag.StringVar (&self.HttpCert,        "http-cert",         self.HttpCert,        "Web server SSL certificate path (leave empty for http)")
	flag.StringVar (&self.HttpKey,         "http-key",          self.HttpKey,         "Web server SSL server key (leave empty for http)")
	flag.IntVar    (&self.HttpPort,        "http-port",         self.HttpPort,        "Web server port")
	flag.StringVar (&self.LogLevel,        "log-level",         self.LogLevel,        "Logger verbosity level")
	flag.StringVar (&self.MailFrom,        "mail-from",         self.MailFrom,        "Mail From: address")
	flag.BoolVar   (&self.MailDry,         "mail-dry",          self.MailDry,         "Disable actual mail sending (dev)")
	flag.BoolVar   (&self.ReloadTemplates, "reload-templates",  self.ReloadTemplates, "Reload ui template on each request (dev)")
	flag.Parse()
}

func (self *AppConfig) parseArgs() {
	// 1.
	self.parseCmdLine()
	if "" != self.ConfigFile {
		self.parseConfig()
	}

	// 2.
	var lTmp AppConfig
	lErr := gautocloud.Inject(&lTmp)
	if lErr != nil {
		log.WithError(lErr).Warn("unable to load gautocloud config")
	}
	mergeObject(self, &lTmp)

	// 3.
	flag.Parse()

	lPort := os.Getenv("PORT")
	if 0 != len(lPort) {
		lVal, lErr := strconv.Atoi(lPort)
		if nil != lErr {
			log.WithError(lErr).Error("invalid PORT env variable")
			os.Exit(1)
		}
		self.HttpPort = lVal
	}
}

func mergeObject(pRef interface{}, pData interface{}) (error) {
	lType    := reflect.TypeOf(pRef)
	if lType != reflect.TypeOf(pData) {
		return errors.New("type mismatch")
	}

	lRefVal  := reflect.ValueOf(pRef)
	lDataVal := reflect.ValueOf(pData)

	if lType.Kind() == reflect.Ptr {
		lRefVal  = lRefVal.Elem()
		lDataVal = lDataVal.Elem()
	}

	for cIdx := 0; cIdx < lRefVal.NumField(); cIdx++ {
		lRefField  := lRefVal.Field(cIdx)
		lDataField := lDataVal.Field(cIdx)
		lFieldType := lDataField.Type()

		if lFieldType.Comparable() {
			lZeroValue := reflect.Zero(lFieldType).Interface()
			lDataValue := lDataField.Interface()
			if lDataValue != lZeroValue {
				lRefField.Set(lDataField)
			}
		} else {
			lRefField.Set(lDataField)
		}
	}
	return nil
}

func init() {
	gautocloud.RegisterConnector(generic.NewConfigGenericConnector(AppConfig{}))

}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
