package ui

import "net/http"
import "errors"
import "html/template"
import "os"
import "path/filepath"
import "github.com/gorilla/mux"
import log "github.com/sirupsen/logrus"
import "github.com/orange-cloudfoundry/cf-wall/core"


type UiHandler struct {
	Config *core.AppConfig
	Tpl    *template.Template
}

func NewUiHandler(pConf *core.AppConfig, pRouter *mux.Router) *UiHandler {
	lTpl, lErr := createTemplates()
	if lErr != nil {
		os.Exit(1)
	}

	lObj := UiHandler{
		Config: pConf,
		Tpl:    lTpl,
	}

	pRouter.PathPrefix("/ui/static/").
		Handler(http.StripPrefix("/ui/static/", http.FileServer(http.Dir("ui/static"))))
	pRouter.HandleFunc("/ui", core.DecorateHandler(lObj.HandlerRequest))
	pRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui", http.StatusMovedPermanently)
	})
	return &lObj
}

func createTemplates() (*template.Template, error) {
	lFuncMap := map[string]interface{}{
		"mkSlice": mkSlice,
		"mkDict":  mkDict,
	}

	lBinDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	lTpl, lErr := template.New("index.tpl").
		Funcs(template.FuncMap(lFuncMap)).
		ParseFiles(
		filepath.Join(lBinDir, "ui/templates/index.tpl"),
		filepath.Join(lBinDir, "ui/templates/header.tpl"),
		filepath.Join(lBinDir, "ui/templates/table.tpl"),
		filepath.Join(lBinDir, "ui/templates/accordion.tpl"))

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
	if self.Config.ReloadTemplates {
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

func mkSlice(pArgs ...interface{}) []interface{} {
	return pArgs
}

func mkDict(pValues ...interface{}) (map[string]interface{}, error) {
	if len(pValues) % 2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	lDict := make(map[string]interface{}, len(pValues) / 2)
	for cIdx := 0; cIdx < len(pValues); cIdx += 2 {
		lKey, lOk := pValues[cIdx].(string)
		if !lOk {
			return nil, errors.New("dict keys must be strings")
		}
		lDict[lKey] = pValues[cIdx + 1]
	}
	return lDict, nil
}
