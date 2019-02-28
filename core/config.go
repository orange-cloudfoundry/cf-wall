package core

import "flag"
import "os"
import "fmt"
import "strconv"
import "encoding/json"
import log "github.com/sirupsen/logrus"
import "github.com/cloudfoundry-community/gautocloud"
import "github.com/cloudfoundry-community/gautocloud/connectors/generic"
import "strings"

type MailCC []string

type AppConfig struct {
	ConfigFile       string
	UaaClientName    string `json:"uaa-client"         cloud:"uaa-client"`
	UaaClientSecret  string `json:"uaa-secret"         cloud:"uaa-secret"`
	UaaEndPoint      string `json:"uaa-url"            cloud:"uaa-url"`
	UaaSkipVerify    bool   `json:"uaa-skip-verify"    cloud:"uaa-skip-verify"`
	CCEndPoint       string `json:"cc-url"             cloud:"cc-url"`
	CCSkipVerify     bool   `json:"cc-skip-verify"     cloud:"cc-skip-verify"`
	HttpCert         string `json:"http-cert"          cloud:"http-cert"`
	HttpKey          string `json:"http-key"           cloud:"http-key"`
	HttpPort         int    `json:"http-port"          cloud:"http-port"`
	LogLevel         string `json:"log-level"          cloud:"log-level"`
	MailFrom         string `json:"mail-from"          cloud:"mail-from"`
	MailDry          bool   `json:"mail-dry"           cloud:"mail-dry"`
	MailCc           MailCC `json:"mail-cc"            cloud:"mail-cc"`
	MailTag          string `json:"mail-tag"           cloud:"mail-tag"`
	MailRateCount    int    `json:"mail-rate-count"    cloud:"mail-rate-count"`
	MailRateDuration int    `json:"mail-rate-duration" cloud:"mail-rate-duration"`
	ReloadTemplates  bool   `json:"reload-templates"   cloud:"reload-templates"`
}

func (self *MailCC) String() string {
	return strings.Join(*self, ",")
}

func (self *MailCC) Set(pVal string) error {
	*self = append(*self, pVal)
	return nil
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
	lConf := AppConfig{}

	InitLogger("error")
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
	flag.StringVar(&self.ConfigFile, "config", self.ConfigFile, "configuration file")
	flag.StringVar(&self.UaaClientName, "uaa-client", self.UaaClientName, "UAA client ID")
	flag.StringVar(&self.UaaClientName, "uaa-secret", self.UaaClientName, "UAA client secret")
	flag.StringVar(&self.UaaEndPoint, "uaa-url", self.UaaEndPoint, "UAA API endpoint url")
	flag.BoolVar(&self.UaaSkipVerify, "uaa-skip-verify", self.UaaSkipVerify, "Do not verify UAA SSL certificates")
	flag.StringVar(&self.CCEndPoint, "cc-url", self.CCEndPoint, "Cloud Controller API endpoint url")
	flag.BoolVar(&self.CCSkipVerify, "cc-skip-verify", self.CCSkipVerify, "Do not verify Cloud Controller SSL certificates")
	flag.StringVar(&self.HttpCert, "http-cert", self.HttpCert, "Web server SSL certificate path (leave empty for http)")
	flag.StringVar(&self.HttpKey, "http-key", self.HttpKey, "Web server SSL server key (leave empty for http)")
	flag.IntVar(&self.HttpPort, "http-port", self.HttpPort, "Web server port")
	flag.StringVar(&self.LogLevel, "log-level", self.LogLevel, "Logger verbosity level")
	flag.StringVar(&self.MailFrom, "mail-from", self.MailFrom, "Mail From: address")
	flag.BoolVar(&self.MailDry, "mail-dry", self.MailDry, "Disable actual mail sending (dev)")
	flag.StringVar(&self.MailTag, "mail-tag", self.MailTag, "Additional tag prefix for sent mails")
	flag.IntVar(&self.MailRateCount, "mail-rate-count", self.MailRateCount, "Limit number of mail sent per timed window")
	flag.IntVar(&self.MailRateDuration, "mail-rate-duration", self.MailRateDuration, "Duration (in seconds) of timed window")
	flag.BoolVar(&self.ReloadTemplates, "reload-templates", self.ReloadTemplates, "Reload ui template on each request (dev)")

	flag.Var(&self.MailCc, "mail-cc", "List of additional recipients to all mails (can give multiple times)")
	flag.Parse()
}

func (self *AppConfig) parseArgs() {
	// 1.
	self.parseCmdLine()
	if "" != self.ConfigFile {
		self.parseConfig()
	}

	// 2.
	lErr := gautocloud.Inject(self)
	if lErr != nil {
		log.WithError(lErr).Warn("unable to load gautocloud config")
	}
	log.WithField("conf", self).Debug("final conf")

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

func init() {
	gautocloud.RegisterConnector(generic.NewConfigGenericConnector(AppConfig{}))
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
