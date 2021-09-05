package cronctl

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var (
	nameLiteral  = "name"
	tokenLiteral = "token"

	contentType     = "Content-Type"
	applicationJson = "application/json"

	parameterNotFoundErr = errors.New("parameter not found")
	tokenNotValidErr     = errors.New("token not valid")
)

func (crontab *Crontab) HttpControl(token string) {
	http.HandleFunc("/crontab-start", crontab.httpStartup(token))
	http.HandleFunc("/crontab-stop", crontab.httpStop(token))
	http.HandleFunc("/crontab-disable", crontab.httpDisable(token))
	http.HandleFunc("/crontab-enable", crontab.httpEnable(token))
	http.HandleFunc("/crontab-details", crontab.httpDetails(token))
}

func (crontab *Crontab) httpDetails(token string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		method := "httpDetails"
		err := tokenValid(token, r)
		if err != nil {
			handleErr(method, err, w)
			return
		}

		details, err := crontab.Details()
		if err != nil {
			handleErr(method, err, w)
			return
		}

		marshal, err := json.Marshal(details)
		if err != nil {
			handleErr(method, err, w)
		}

		w.Header().Set(contentType, applicationJson)
		_, _ = w.Write(marshal)
	}
}

func (crontab *Crontab) httpDisable(token string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		method := "httpDisable"
		err := tokenValid(token, r)
		if err != nil {
			handleErr(method, err, w)
			return
		}

		name, err := parameter(nameLiteral, r)
		if err != nil {
			handleErr(method, err, w)
			return
		}

		err = crontab.Disable(name)
		handleErr(method, err, w)
	}
}

func (crontab *Crontab) httpEnable(token string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		method := "httpEnable"
		err := tokenValid(token, r)
		if err != nil {
			handleErr(method, err, w)
			return
		}

		name, err := parameter(nameLiteral, r)
		if err != nil {
			handleErr(method, err, w)
			return
		}

		err = crontab.Enable(name)
		handleErr(method, err, w)
	}
}

func (crontab *Crontab) httpStop(token string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		method := "httpStop"
		err := tokenValid(token, r)
		if err != nil {
			handleErr(method, err, w)
			return
		}

		err = crontab.Stop()
		handleErr(method, err, w)
	}
}

func (crontab *Crontab) httpStartup(token string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		method := "httpStartup"
		err := tokenValid(token, r)
		if err != nil {
			handleErr(method, err, w)
			return
		}

		err = crontab.Startup()
		handleErr(method, err, w)
	}
}

func parameter(name string, r *http.Request) (string, error) {
	values := r.URL.Query()
	val, ok := values[name]
	if !ok || len(val) < 1 {
		return "", parameterNotFoundErr
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

func handleErr(method string, err error, w http.ResponseWriter) {
	msg := "ok"
	if err != nil {
		msg = err.Error()
		log.Errorf("%v: %v", method, err)
	}
	_, _ = w.Write([]byte(msg))
}
