package main

import "encoding/json"
import "net/http"

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
