package main

import "net/http"
import "errors"
import "html/template"
import "os"
import "github.com/gorilla/mux"
import log "github.com/sirupsen/logrus"

type UiHandler struct {
	Tpl *template.Template
}

func NewUiHandler(pRouter *mux.Router) UiHandler {
	lTpl, lErr := createTemplates()
	if lErr != nil {
		os.Exit(1)
	}

	lObj := UiHandler{Tpl: lTpl}
	pRouter.PathPrefix("/ui/static/").
		Handler(http.StripPrefix("/ui/static/", http.FileServer(http.Dir("ui/static"))))
	pRouter.HandleFunc("/ui", DecorateHandler(lObj.HandlerRequest))

	return lObj
}

func createTemplates() (*template.Template, error) {
	lFuncMap := map[string]interface{}{
		"mkSlice": mkSlice,
		"mkDict":  mkDict,
	}

	lTpl, lErr := template.New("index.tpl").
		Funcs(template.FuncMap(lFuncMap)).
		ParseFiles("ui/index.tpl", "ui/header.tpl", "ui/table.tpl", "ui/accordion.tpl")

	if lErr != nil {
		log.WithError(lErr).Error("unable to parse ui template")
		return nil, lErr
	}
	return lTpl, nil

}

func (self *UiHandler) reloadTempaltes() error {
	lTpl, lErr := createTemplates()
	if lErr != nil {
		return lErr
	}
	self.Tpl = lTpl
	return nil
}

func (self *UiHandler) HandlerRequest(pRes http.ResponseWriter, pReq *http.Request) {
	if GApp.Config.ReloadTemplates {
		lErr := self.reloadTempaltes()
		if lErr != nil {
			pRes.Write([]byte(lErr.Error()))
		}
	}

	lErr := self.Tpl.Execute(pRes, nil)
	if lErr != nil {
		log.WithError(lErr).Error("unable to render ui template")
		pRes.Write([]byte(lErr.Error()))
	}
}

func mkSlice(args ...interface{}) []interface{} {
	return args
}

func mkDict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}
