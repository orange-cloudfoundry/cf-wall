package core

import "errors"
import "encoding/json"
import "net/http"
import log "github.com/sirupsen/logrus"


type HttpError struct {
	Error  error
	Status int
	Code   int
}

func NewHttpError(pErr error, pStatus int, pCode int) HttpError {
	return HttpError{pErr, pStatus, pCode}
}

func WriteJson(pWriter http.ResponseWriter, pObj interface{}) {
	lVal, _ := json.Marshal(pObj)
	pWriter.Header().Set("Content-Type", "application/json")
	pWriter.Write(lVal)
}

func WriteJsonError(pWriter http.ResponseWriter, pStatus int, pCode int, pErr error) {
	lErr := struct {
		Code  int    `json:"code"`
		Error string `json:"error"`
	}{ pCode, pErr.Error() }

	lVal, _ := json.Marshal(lErr)

	pWriter.Header().Set("Content-Type", "application/json")
	pWriter.WriteHeader(pStatus)
	pWriter.Write(lVal)
}

func HandlePanic(pRes http.ResponseWriter, pReq *http.Request) {
	switch lErr := recover().(type) {
	case nil:
	case HttpError:
		WriteJsonError(pRes, lErr.Status, lErr.Code, lErr.Error)
	default:
		WriteJsonError(pRes, 500, 20, errors.New("undefined panic type"))
	}
}

func DecorateHandler(pFunc func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	lReq     := WrapHandler(LogRequestHandler, pFunc)
	lRes     := WrapHandler(lReq, LogResponseHandler)
	lProtect := WrapDefer(lRes, HandlePanic)
	return lProtect
}

func LogRequestHandler(pRes http.ResponseWriter, pReq *http.Request) {
	log.WithFields(log.Fields{
		"method": pReq.Method,
		"host":   pReq.Host,
		"client": pReq.RemoteAddr,
		"path":   pReq.RequestURI,
		"headers": pReq.Header,
	}).Info("handling http request")
}

func LogResponseHandler(pRes http.ResponseWriter, pReq *http.Request) {
}

func WrapHandler(pDecorator func(http.ResponseWriter, *http.Request), pFunc func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(pRes http.ResponseWriter, pReq *http.Request) {
		pDecorator(pRes, pReq)
		pFunc(pRes, pReq)
	}
}

func WrapDefer(pFunc func(http.ResponseWriter, *http.Request), pDefer func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(pRes http.ResponseWriter, pReq *http.Request) {
		defer pDefer(pRes, pReq)
		pFunc(pRes, pReq)
	}
}
