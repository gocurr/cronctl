package cronctl

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
)

var (
	reg = regexp.MustCompile("\\s+")

	nameLiteral  = "name"
	tokenLiteral = "token"

	contentType     = "Content-Type"
	applicationJson = "application/json"

	parameterNotFoundErr = errors.New("parameter not found")
	tokenNotValidErr     = errors.New("token not valid")
)

func (crontab *Crontab) HttpControl(path, token string) {
	basePath := base(path)
	http.HandleFunc(basePath+"crontab-start", crontab.httpStartup(token))
	http.HandleFunc(basePath+"crontab-continue", crontab.httpContinue(token))
	http.HandleFunc(basePath+"crontab-suspend", crontab.httpSuspend(token))
	http.HandleFunc(basePath+"crontab-disable", crontab.httpDisable(token))
	http.HandleFunc(basePath+"crontab-enable", crontab.httpEnable(token))
	http.HandleFunc(basePath+"crontab-details", crontab.httpDetails(token))
}

func compressStr(str string) string {
	if str == "" {
		return ""
	}
	return reg.ReplaceAllString(str, "")
}

func base(path string) string {
	path = compressStr(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	return path
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

func (crontab *Crontab) httpSuspend(token string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		method := "httpSuspend"
		err := tokenValid(token, r)
		if err != nil {
			handleErr(method, err, w)
			return
		}

		err = crontab.Suspend()
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

func (crontab *Crontab) httpContinue(token string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		method := "httpContinue"
		err := tokenValid(token, r)
		if err != nil {
			handleErr(method, err, w)
			return
		}

		err = crontab.Continue()
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
