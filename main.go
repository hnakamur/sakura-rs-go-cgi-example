package main

import (
	"errors"
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

func handleGetList(w http.ResponseWriter, r *http.Request) *webapputil.HTTPError {
	ltsvlog.Logger.Info().String("handler", "handleGetList").String("reqID", webapputil.RequestID(r)).Log()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "handleGetList -------- Standard Library ---------")
	fmt.Fprintln(w, "Method:", r.Method)
	fmt.Fprintln(w, "URL:", r.URL.String())
	fmt.Fprintln(w, "URL.Path:", r.URL.Path)
	fmt.Fprintln(w, "HTTPS:", os.Getenv("HTTPS"))
	name := "Golang GCI"
	fmt.Fprintf(w, "hello:%s!!", name)
	return nil
}

func handleGetRoot(w http.ResponseWriter, r *http.Request) *webapputil.HTTPError {
	ltsvlog.Logger.Info().String("handler", "handleGetRoot").String("reqID", webapputil.RequestID(r)).Log()
	err := r.ParseForm()
	if err != nil {
		err = ltsvlog.WrapErr(err, func(err error) error {
			return fmt.Errorf("failed to parse form, err=%s", err)
		})
		return webapputil.NewHTTPError(err, http.StatusBadRequest)
	}
	statusStr := r.FormValue("status")
	if statusStr != "" {
		status, err := strconv.Atoi(statusStr)
		if err != nil {
			err = ltsvlog.WrapErr(err, func(err error) error {
				return fmt.Errorf("failed to convert status parameter to int, err=%s", err)
			}).String("statusStr", statusStr)
			return webapputil.NewHTTPError(err, http.StatusBadRequest)
		}

		err = ltsvlog.Err(errors.New("intentional error for testing"))
		return webapputil.NewHTTPError(err, status)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "handleGetRoot -------- Standard Library ---------")
	fmt.Fprintln(w, "Method:", r.Method)
	fmt.Fprintln(w, "URL:", r.URL.String())
	fmt.Fprintln(w, "URL.Path:", r.URL.Path)
	fmt.Fprintln(w, "HTTPS:", os.Getenv("HTTPS"))
	name := "Golang GCI"
	fmt.Fprintf(w, "hello:%s!!", name)
	return nil
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

	errorHandler := func(err *webapputil.HTTPError, w http.ResponseWriter, r *http.Request) {
		lerr, ok := err.Error.(*ltsvlog.Error)
		if !ok {
			lerr = ltsvlog.Err(err.Error)
		}
		ltsvlog.Logger.Err(lerr.Int("status", err.Status).
			String("reqID", webapputil.RequestID(r)))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(err.Status)
	}
	withErrorHandler := func(next func(w http.ResponseWriter, r *http.Request) *webapputil.HTTPError) http.Handler {
		return webapputil.WithErrorHandler(next, errorHandler)
	}

	mux := http.NewServeMux()
	mux.Handle(scriptName+"/list", withErrorHandler(handleGetList))
	mux.Handle(scriptName, withErrorHandler(handleGetRoot))
	err = cgi.Serve(webapputil.RequestIDMiddleware(
		webapputil.AccessLogMiddleware(mux, writeAccessLog),
		generateRequestID))
	if err != nil {
		log.Fatal(err)
	}
}
