package main

import      "net/http"
import      "strings"
import      "errors"
import      "github.com/cloudfoundry-community/go-cfclient"
import log  "github.com/sirupsen/logrus"

func NewCCCliFromRequest(p_conf *AppConfig, p_req *http.Request) (*cfclient.Client, error) {

  l_token := p_conf.CCToken
  if (l_token == "") {
    l_auth, l_ok := p_req.Header["Authorization"]
    if (l_ok == false) || (len(l_auth) == 0) {
      l_err := errors.New("Authorization header is mandatory")
      log.WithError(l_err).Error("unable to create CC client")
      return nil, l_err
    }

    l_parts  := strings.Fields(l_auth[0])
    if (len(l_parts) < 2) {
      l_err := errors.New("malformated Authorization header")
      log.WithError(l_err).Error("unable to create CC client")
      return nil, l_err
    }
    l_token = l_parts[1]
  }

  l_cli, l_err := NewCCCli(p_conf, l_token)
  if (l_err != nil) {
    log.WithError(l_err).Error("unable to create CC client")
    return nil, l_err
  }

  return l_cli, nil
}

func NewCCCli(p_conf *AppConfig, p_token string) (*cfclient.Client, error) {
  log.WithFields(log.Fields{
    "endpoint" : p_conf.CCEndPoint,
  }).Debug("creating CC client")
  l_conf := cfclient.Config{
    ApiAddress: p_conf.CCEndPoint,
    Token:      p_token,
  }
  return cfclient.NewClient(&l_conf)
}
