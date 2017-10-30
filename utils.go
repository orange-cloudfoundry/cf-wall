package main

import "errors"
import "encoding/json"
import "net/http"
import log "github.com/sirupsen/logrus"

type HttpHandler = func(http.ResponseWriter, *http.Request)
type HttpError struct {
	Error  error
	Status int
	Code   int
}

func WriteJson(p_writer http.ResponseWriter, p_obj interface{}) {
	l_val, _ := json.Marshal(p_obj)
	p_writer.Header().Set("Content-type", "application/json")
	p_writer.Write(l_val)
}

func WriteJsonError(p_writer http.ResponseWriter, p_status int, p_code int, p_err error) {
	p_writer.WriteHeader(p_status)
	l_err := struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}{p_code, p_err.Error()}
	WriteJson(p_writer, l_err)
}

func HandlePanic(p_res http.ResponseWriter, p_req *http.Request) {
	switch l_err := recover().(type) {
	case nil:
	case HttpError:
		WriteJsonError(p_res, l_err.Status, l_err.Code, l_err.Error)
	default:
		WriteJsonError(p_res, 500, 20, errors.New("undefined panic type"))
	}
}

func DecorateHandler(p_func HttpHandler) HttpHandler {
	l_req     := WrapHandler(LogRequestHandler, p_func)
	l_res     := WrapHandler(l_req, LogResponseHandler)
	l_protect := WrapDefer(l_res, HandlePanic)
	return l_protect
}

func LogRequestHandler(p_res http.ResponseWriter, p_req *http.Request) {
	log.WithFields(log.Fields{
		"method": p_req.Method,
		"host":   p_req.Host,
		"client": p_req.RemoteAddr,
		"path":   p_req.RequestURI,
	}).Info("handling http request")
}

func LogResponseHandler(p_res http.ResponseWriter, p_req *http.Request) {
}

func WrapHandler(p_decorator HttpHandler, p_func HttpHandler) HttpHandler {
	return func(p_res http.ResponseWriter, p_req *http.Request) {
		p_decorator(p_res, p_req)
		p_func(p_res, p_req)
	}
}

func WrapDefer(p_func HttpHandler, p_defer HttpHandler) HttpHandler {
	return func(p_res http.ResponseWriter, p_req *http.Request) {
		defer p_defer(p_res, p_req)
		p_func(p_res, p_req)
	}
}
