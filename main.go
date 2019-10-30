package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

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
	router.HandlerFunc("GET", "/api", handleRequestAndRedirect)
	router.HandlerFunc("POST", "/api", handleRequestAndRedirect)
	router.HandlerFunc("PUT", "/api", handleRequestAndRedirect)
	router.HandlerFunc("DELETE", "/api", handleRequestAndRedirect)
	router.HandlerFunc("OPTIONS", "/api", handleRequestAndRedirect)

	logrus.Fatal(http.ListenAndServe(AppUrl, router))
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	trackId, _ := helper.GenerateRandomString(10)
	logger := logrus.WithFields(logrus.Fields{
		"originUrl": req.URL,
		"trackId":   trackId,
	})

	resolvedUrl, error := getProxyUrl(req, logger)
	if error != nil {
		http.Error(res, error.Error(), error.(errors.ProxyError).Code)
	}

	serveReverseProxy(resolvedUrl, res, req, logger, trackId)
	logRequestPerformance(req, logger)
}

func getProxyUrl(req *http.Request, logger *logrus.Entry) (*url.URL, error) {
	originalPath := req.URL.Path
	slashIndex := strings.Index(originalPath, "/")

	serviceName := originalPath[:slashIndex]
	if targetPort, ok := ServiceToDnsResolver[serviceName]; ok {
		req.URL.Host = fmt.Sprintf("%s:%s", req.URL.Hostname(), targetPort)
		req.URL.Path = originalPath[slashIndex+1:]

		return req.URL, nil
	}

	return req.URL, errors.ProxyError{errors.ServiceNotFound, "Service not found"}
}

// TODO: consider moving this to be a middleware
func logRequestPerformance(req *http.Request, logger *logrus.Entry) {
}

func serveReverseProxy(resolvedUrl *url.URL, res http.ResponseWriter, req *http.Request, logger *logrus.Entry, trackId string) {
	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(resolvedUrl)

	// Update the headers to allow for SSL redirection
	req.URL.Host = resolvedUrl.Host
	req.URL.Scheme = resolvedUrl.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Header.Set("X-Origin-Host", AppUrl)
	req.Host = resolvedUrl.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
