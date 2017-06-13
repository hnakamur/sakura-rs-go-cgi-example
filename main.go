package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	cgiutil "github.com/hnakamur/sakura-rs-cgiutil"
	"github.com/hnakamur/webapputil"

	"github.com/hnakamur/ltsvlog"
)

func handleGetList(w http.ResponseWriter, r *http.Request) {
	ltsvlog.Logger.Info().String("handler", "handleGetList").String("reqID", webapputil.RequestID(r)).Log()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "handleGetList -------- Standard Library ---------")
	fmt.Fprintln(w, "Method:", r.Method)
	fmt.Fprintln(w, "URL:", r.URL.String())
	fmt.Fprintln(w, "URL.Path:", r.URL.Path)
	fmt.Fprintln(w, "HTTPS:", os.Getenv("HTTPS"))
	name := "Golang GCI"
	fmt.Fprintf(w, "hello:%s!!", name)
}

func handleGetRoot(w http.ResponseWriter, r *http.Request) {
	ltsvlog.Logger.Info().String("handler", "handleGetRoot").String("reqID", webapputil.RequestID(r)).Log()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "handleGetRoot -------- Standard Library ---------")
	fmt.Fprintln(w, "Method:", r.Method)
	fmt.Fprintln(w, "URL:", r.URL.String())
	fmt.Fprintln(w, "URL.Path:", r.URL.Path)
	fmt.Fprintln(w, "HTTPS:", os.Getenv("HTTPS"))
	name := "Golang GCI"
	fmt.Fprintf(w, "hello:%s!!", name)
}

func main() {
	logDir := filepath.Join(cgiutil.GetHomeDir(), "log")
	scriptBasename := path.Base(os.Getenv("SCRIPT_FILENAME"))
	accessLogFilename := path.Join(logDir, scriptBasename+".access.log")
	errorLogFilename := path.Join(logDir, scriptBasename+".error.log")

	errorLogFile, err := os.OpenFile(errorLogFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0606)
	if err != nil {
		log.Fatal(err)
	}
	defer errorLogFile.Close()
	ltsvlog.Logger = ltsvlog.NewLTSVLogger(errorLogFile, true)

	accessLogFile, err := os.OpenFile(accessLogFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0606)
	if err != nil {
		log.Fatal(err)
	}
	defer accessLogFile.Close()
	accessLogger := ltsvlog.NewLTSVLogger(accessLogFile, false, ltsvlog.SetLevelLabel(""))

	generateRequestID := func(req *http.Request) string {
		buf := make([]byte, 0, 64)
		buf = strconv.AppendInt(buf, time.Now().UnixNano(), 36)
		buf = append(buf, '_')
		buf = strconv.AppendInt(buf, int64(os.Getpid()), 36)
		return string(buf)
	}

	writeAccessLog := func(res webapputil.ResponseLogInfo, req *http.Request) {
		accessLogger.Info().String("method", req.Method).Stringer("url", req.URL).
			String("proto", req.Proto).String("host", req.Host).
			String("remoteAddr", req.RemoteAddr).
			String("ua", req.Header.Get("User-Agent")).
			String("reqID", webapputil.RequestID(req)).
			Int("status", res.StatusCode).Int64("sentBodySize", res.SentBodySize).
			Sprintf("elapsed", "%e", res.Elapsed.Seconds()).Log()
	}

	scriptName := os.Getenv("SCRIPT_NAME")

	mux := http.NewServeMux()
	mux.HandleFunc(scriptName+"/list", handleGetList)
	mux.HandleFunc(scriptName, handleGetRoot)
	err = cgi.Serve(webapputil.RequestIDMiddleware(
		webapputil.AccessLogMiddleware(mux, writeAccessLog),
		generateRequestID))
	if err != nil {
		log.Fatal(err)
	}
}
