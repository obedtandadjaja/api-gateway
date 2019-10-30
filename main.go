package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/obedtandadjaja/api-gateway/errors"
	"github.com/obedtandadjaja/api-gateway/helper"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
)

var Environment string
var AppHost string
var AppPort string
var AppUrl string
var ServiceToDnsResolver = map[string]int{
	"auth":        8080,
	"backend-api": 4000,
}

func init() {
	Environment = os.Getenv("ENV")
	AppHost = os.Getenv("APP_HOST")
	AppPort = os.Getenv("APP_PORT")
	AppUrl = AppHost + ":" + AppPort
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	router := httprouter.New()

	// figure out a better way to do this
	router.HandlerFunc("GET", "/api/*path", handleRequestAndRedirect)
	router.HandlerFunc("POST", "/api/*path", handleRequestAndRedirect)
	router.HandlerFunc("PUT", "/api/*path", handleRequestAndRedirect)
	router.HandlerFunc("DELETE", "/api/*path", handleRequestAndRedirect)
	router.HandlerFunc("OPTIONS", "/api/*path", handleRequestAndRedirect)

	logrus.Info("App running on port " + AppUrl)
	logrus.Fatal(http.ListenAndServe(AppUrl, router))
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	start := time.Now()

	trackId, _ := helper.GenerateRandomString(12)
	logger := logrus.WithFields(logrus.Fields{
		"originUrl":  req.URL,
		"trackId":    trackId,
		"remoteAddr": req.RemoteAddr,
		"UserAgent":  req.UserAgent(),
		"Method":     req.Method,
	})

	resolvedUrl, error := getProxyUrl(req, logger)
	if error != nil {
		http.Error(res, error.Error(), error.(errors.ProxyError).Code)
	}

	serveReverseProxy(resolvedUrl, res, req, logger, trackId)
	logRequestPerformance(req, logger, start)
}

func getProxyUrl(req *http.Request, logger *logrus.Entry) (*url.URL, error) {
	originalPath := req.URL.Path
	slashIndex := strings.Index(originalPath[5:], "/")
	if slashIndex == -1 {
		return req.URL, errors.ProxyError{errors.BadRequest, "Bad request"}
	}

	serviceName := originalPath[5 : slashIndex+5]
	if targetPort, ok := ServiceToDnsResolver[serviceName]; ok {
		colonIndex := strings.Index(req.Host, ":")
		req.URL.Host = fmt.Sprintf("%v:%v", req.Host[:colonIndex], targetPort)
		req.URL.Path = originalPath[len(serviceName)+6:]

		return req.URL, nil
	}

	return req.URL, errors.ProxyError{errors.ServiceNotFound, "Service not found"}
}

func logRequestPerformance(req *http.Request, logger *logrus.Entry, start time.Time) {
	duration := time.Now().Sub(start)

	logger.Info(fmt.Sprintf("Proxy duration %v", duration))
}

func serveReverseProxy(resolvedUrl *url.URL, res http.ResponseWriter, req *http.Request, logger *logrus.Entry, trackId string) {
	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(resolvedUrl)

	// make the new request scheme to http
	req.URL = resolvedUrl
	req.URL.Scheme = "http"

	req.Header.Set("X-Forwarded-Host", req.Host)
	req.Header.Set("X-Origin-Host", AppUrl)
	req.Header.Set("X-Track-ID", trackId)

	logger.Info(req)

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
