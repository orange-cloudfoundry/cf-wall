package main

import "flag"
import "os"
import "fmt"
import "encoding/json"
import "code.cloudfoundry.org/lager"

type AppConfig struct {
  ConfigFile      string
  UaaClientName   string `json:"uaa-client"`
  UaaClientSecret string `json:"uaa-secret"`
  UaaEndPoint     string `json:"uaa-url"`
  CCEndPoint      string `json:"cc-url"`
  CCToken         string `json:"cc-override-token"`
  HttpCert        string `json:"http-cert"`
  HttpKey         string `json:"http-key"`
  HttpPort        int    `json:"http-port"`
  LogLevel        string `json:"log-level"`
  LogName         string `json:"log-name"`
}


func NewLogger(p_level string) (lager.Logger) {
  l_res := lager.NewLogger("cfy-wall")
  l_level := lager.ERROR
  switch p_level {
  case "debug": l_level = lager.DEBUG
  case "info":  l_level = lager.INFO
  case "error": l_level = lager.ERROR
  case "fatal": l_level = lager.FATAL
  }

  l_sink := lager.NewWriterSink(os.Stdout, l_level)
  l_res.RegisterSink(l_sink)
  return l_res
}


func NewAppConfig() (AppConfig) {
  l_conf := AppConfig{
    ConfigFile      : "",
    UaaClientName   : "cfy-wall",
    UaaClientSecret : "password",
    UaaEndPoint     : "https://uaa.example.com",
    CCEndPoint      : "https://api.example.com",
    CCToken         : "",
    HttpCert        : "",
    HttpKey         : "",
    HttpPort        : 443,
    LogLevel        : "error",
    LogName         : "cfy-wall",
  }
  l_conf.parseArgs()
  return l_conf
}

func (self *AppConfig) parseConfig() {
  l_file, l_err := os.Open(self.ConfigFile)
  if (l_err != nil) {
    fmt.Printf("unable to read configuration file '%s'", self.ConfigFile)
    os.Exit(1)
  }

  l_decoder := json.NewDecoder(l_file)
  l_err = l_decoder.Decode(&self)
  if l_err != nil {
    fmt.Printf("unable to parse file '%s' : %s", self.ConfigFile, l_err.Error())
    os.Exit(1)
  }
}

func (self *AppConfig) parseCmdLine() {
  flag.StringVar(&self.ConfigFile,    "config",             self.ConfigFile,    "configuration file")
  flag.StringVar(&self.UaaClientName, "uaa-client",         self.UaaClientName, "UAA client ID")
  flag.StringVar(&self.UaaClientName, "uaa-secret",         self.UaaClientName, "UAA client secret")
  flag.StringVar(&self.UaaEndPoint,   "uaa-url",            self.UaaEndPoint,   "UAA API endpoint url")
  flag.StringVar(&self.CCEndPoint,    "cc-url",             self.CCEndPoint,    "Cloud Controller API endpoint url")
  flag.StringVar(&self.CCToken,       "cc-override-token",  self.CCToken,       "Override token used to communicates with Cloud Controller")
  flag.StringVar(&self.HttpCert,      "http-cert",          self.HttpCert,      "Web server SSL certificate path (leave empty for http)")
  flag.StringVar(&self.HttpKey,       "http-key",           self.HttpKey,       "Web server SSL server key (leave empty for http)")
  flag.IntVar(&self.HttpPort,         "http-port",          self.HttpPort,      "Web server port")
  flag.StringVar(&self.LogLevel,      "log-level",          self.LogLevel,      "Logger verbosity level")
  flag.StringVar(&self.LogName,      "log-Name",            self.LogName,       "Logger component name")
  flag.Parse()
}

func (self *AppConfig) parseArgs() {
  self.parseCmdLine()
  if ("" != self.ConfigFile) {
    self.parseConfig()
    flag.Parse()
  }
}


// Local Variables:
// ispell-local-dictionary: "american"
// End:
