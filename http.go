package cronctl

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	startLiteral    = "start"
	continueLiteral = "continue"
	suspendLiteral  = "suspend"
	disableLiteral  = "disable"
	enableLiteral   = "enable"
	detailLiteral   = "details"

	nameLiteral  = "name"
	tokenLiteral = "token"

	contentType     = "Content-Type"
	applicationJson = "application/json"
)

var (
	unknownTypeErr   = errors.New("unknown type")
	tokenNotValidErr = errors.New("token not valid")
)

func (crontab *Crontab) HttpCronCtrlFunc(token string, logging bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		method := "httpCronCtrl"
		err := tokenValid(token, r)
		if err != nil {
			crontab.handleErr(method, err, w, logging)
			return
		}

		typ, err := parameter("type", r)
		if err != nil {
			crontab.handleErr(method, err, w, logging)
			return
		}

		switch typ {
		case startLiteral:
			crontab.httpStartup(w, logging)
		case continueLiteral:
			crontab.httpContinue(w, logging)
		case suspendLiteral:
			crontab.httpSuspend(w, logging)
		case detailLiteral:
			crontab.httpDetails(w, logging)
		case enableLiteral:
			crontab.httpEnable(w, r, logging)
		case disableLiteral:
			crontab.httpDisable(w, r, logging)
		default:
			crontab.handleErr(method, unknownTypeErr, w, logging)
		}
	}
}

func (crontab *Crontab) httpDetails(w http.ResponseWriter, logging bool) {
	method := "httpDetails"
	details, err := crontab.Details()
	if err != nil {
		crontab.handleErr(method, err, w, logging)
		return
	}

	marshal, err := json.Marshal(details)
	if err != nil {
		crontab.handleErr(method, err, w, logging)
	}

	w.Header().Set(contentType, applicationJson)
	_, _ = w.Write(marshal)
}

func (crontab *Crontab) httpDisable(w http.ResponseWriter, r *http.Request, logging bool) {
	method := "httpDisable"
	name, err := parameter(nameLiteral, r)
	if err != nil {
		crontab.handleErr(method, err, w, logging)
		return
	}

	err = crontab.Disable(name)
	crontab.handleErr(method, err, w, logging)
}

func (crontab *Crontab) httpEnable(w http.ResponseWriter, r *http.Request, logging bool) {
	method := "httpEnable"
	name, err := parameter(nameLiteral, r)
	if err != nil {
		crontab.handleErr(method, err, w, logging)
		return
	}

	err = crontab.Enable(name)
	crontab.handleErr(method, err, w, logging)
}

func (crontab *Crontab) httpSuspend(w http.ResponseWriter, logging bool) {
	method := "httpSuspend"

	err := crontab.Suspend()
	crontab.handleErr(method, err, w, logging)
}

func (crontab *Crontab) httpStartup(w http.ResponseWriter, logging bool) {
	method := "httpStartup"

	err := crontab.Startup()
	crontab.handleErr(method, err, w, logging)
}

func (crontab *Crontab) httpContinue(w http.ResponseWriter, logging bool) {
	method := "httpContinue"

	err := crontab.Continue()
	crontab.handleErr(method, err, w, logging)
}

func parameter(name string, r *http.Request) (string, error) {
	values := r.URL.Query()
	val, ok := values[name]
	if !ok || len(val) < 1 {
		return "", fmt.Errorf(`parameter "%s" not found`, name)
	}

	return val[0], nil
}

func tokenValid(token string, r *http.Request) error {
	t, err := parameter(tokenLiteral, r)
	if err != nil || t != token {
		return tokenNotValidErr
	}

	return nil
}

func (crontab *Crontab) handleErr(method string, err error, w http.ResponseWriter, logging bool) {
	msg := "ok"
	if err != nil {
		msg = err.Error()
		if logging {
			crontab.logger.Error(err, method)
		}
	}
	_, _ = w.Write([]byte(msg))
}
