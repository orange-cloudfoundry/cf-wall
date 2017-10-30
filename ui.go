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

func NewUiHandler(p_router *mux.Router) UiHandler {
	l_tpl, l_err := createTemplates()
	if l_err != nil {
		os.Exit(1)
	}

	l_obj := UiHandler{Tpl: l_tpl}
	p_router.PathPrefix("/ui/static/").
		Handler(http.StripPrefix("/ui/static/", http.FileServer(http.Dir("ui/static"))))
	p_router.HandleFunc("/ui", DecorateHandler(l_obj.HandlerRequest))

	return l_obj
}

func createTemplates() (*template.Template, error) {
	l_funcMap := map[string]interface{}{
		"mkSlice": mkSlice,
		"mkDict":  mkDict,
	}

	l_tpl, l_err := template.New("index.tpl").
		Funcs(template.FuncMap(l_funcMap)).
		ParseFiles("ui/index.tpl", "ui/header.tpl", "ui/table.tpl", "ui/accordion.tpl")

	if l_err != nil {
		log.WithError(l_err).Error("unable to parse ui template")
		return nil, l_err
	}
	return l_tpl, nil

}

func (self *UiHandler) reloadTempaltes() error {
	l_tpl, l_err := createTemplates()
	if l_err != nil {
		return l_err
	}
	self.Tpl = l_tpl
	return nil
}

func (self *UiHandler) HandlerRequest(p_res http.ResponseWriter, p_req *http.Request) {
	if GApp.Config.ReloadTemplates {
		l_err := self.reloadTempaltes()
		if l_err != nil {
			p_res.Write([]byte(l_err.Error()))
		}
	}

	l_err := self.Tpl.Execute(p_res, nil)
	if l_err != nil {
		log.WithError(l_err).Error("unable to render ui template")
		p_res.Write([]byte(l_err.Error()))
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
