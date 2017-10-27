package main

import "errors"
import "encoding/json"
import "net/http"

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
  } { p_code, p_err.Error() }
  WriteJson(p_writer, l_err)
}

func panicHandler(p_res http.ResponseWriter, p_req *http.Request) {
  switch l_err := recover().(type) {
  case nil:
  case HttpError:
    WriteJsonError(p_res, l_err.Status, l_err.Code, l_err.Error)
  default:
    WriteJsonError(p_res, 500, 20, errors.New("undefined panic type"))
  }
}

func DecorateHandler(p_func HttpHandler, p_decorator HttpHandler) (HttpHandler) {
  return func(p_res http.ResponseWriter, p_req *http.Request) {
    p_decorator(p_res, p_req)
    p_func(p_res, p_req)
  }
}
