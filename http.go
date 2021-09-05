package cronctl

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
)

const (
	cronControlLiteral = "cron-control"
	startLiteral       = "start"
	continueLiteral    = "continue"
	suspendLiteral     = "suspend"
	disableLiteral     = "disable"
	enableLiteral      = "enable"
	detailLiteral      = "details"

	nameLiteral  = "name"
	tokenLiteral = "token"

	contentType     = "Content-Type"
	applicationJson = "application/json"
)

var (
	reg = regexp.MustCompile("\\s+")

	unknownTypeErr   = errors.New("unknow type")
	tokenNotValidErr = errors.New("token not valid")
)

func (crontab *Crontab) HttpControl(path, token string) {
	basePath := base(path)
	http.HandleFunc(basePath+cronControlLiteral, httpCronCtrl(crontab, token))
}

func httpCronCtrl(crontab *Crontab, token string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		method := "httpCronCtrl"
		err := tokenValid(token, r)
		if err != nil {
			handleErr(method, err, w)
			return
		}

		typ, err := parameter("type", r)
		if err != nil {
			handleErr(method, err, w)
			return
		}

		switch typ {
		case startLiteral:
			crontab.httpStartup(w)
		case continueLiteral:
			crontab.httpContinue(w)
		case suspendLiteral:
			crontab.httpSuspend(w)
		case detailLiteral:
			crontab.httpDetails(w)
		case enableLiteral:
			crontab.httpEnable(w, r)
		case disableLiteral:
			crontab.httpDisable(w, r)
		default:
			handleErr(method, unknownTypeErr, w)
		}
	}
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

func (crontab *Crontab) httpDetails(w http.ResponseWriter) {
	method := "httpDetails"
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

func (crontab *Crontab) httpDisable(w http.ResponseWriter, r *http.Request) {
	method := "httpDisable"
	name, err := parameter(nameLiteral, r)
	if err != nil {
		handleErr(method, err, w)
		return
	}

	err = crontab.Disable(name)
	handleErr(method, err, w)
}

func (crontab *Crontab) httpEnable(w http.ResponseWriter, r *http.Request) {
	method := "httpEnable"
	name, err := parameter(nameLiteral, r)
	if err != nil {
		handleErr(method, err, w)
		return
	}

	err = crontab.Enable(name)
	handleErr(method, err, w)
}

func (crontab *Crontab) httpSuspend(w http.ResponseWriter) {
	method := "httpSuspend"

	err := crontab.Suspend()
	handleErr(method, err, w)
}

func (crontab *Crontab) httpStartup(w http.ResponseWriter) {
	method := "httpStartup"

	err := crontab.Startup()
	handleErr(method, err, w)
}

func (crontab *Crontab) httpContinue(w http.ResponseWriter) {
	method := "httpContinue"

	err := crontab.Continue()
	handleErr(method, err, w)
}

func parameter(name string, r *http.Request) (string, error) {
	values := r.URL.Query()
	val, ok := values[name]
	if !ok || len(val) < 1 {
		return "", errors.New(fmt.Sprintf(`parameter "%s" not found`, name))
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
