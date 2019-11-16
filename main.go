package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/obedtandadjaja/api-gateway/api"
	"github.com/obedtandadjaja/api-gateway/helper"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

var Environment string
var AppHost string
var AppPort string
var AppUrl string

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
	router.GET("/api/health", api.Health)

	for _, resolver := range PathsResolver {
		baseUrlString := fmt.Sprintf("http://%s", AppHost)
		baseUrl, err := url.Parse(baseUrlString)
		if err != nil {
			panic(fmt.Sprintf("Cannot parse url: %s", baseUrlString))
		}

		reverseProxy := httputil.NewSingleHostReverseProxy(baseUrl)
		reverseProxy.Director = func(r *http.Request) {
			r.Header.Add("X-Forwarded-Host", r.Host)
			r.Header.Add("X-Origin-Host", AppUrl)

			r.URL.Host = baseUrl.Host
			r.URL.Scheme = baseUrl.Scheme
			r.URL.Path = resolver.ActualPath
		}

		router.Handle(resolver.Method, resolver.ProxyPath, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			start := time.Now()

			trackId, _ := helper.GenerateRandomString(12)
			logger := logrus.WithFields(logrus.Fields{
				"TrackId": trackId,
			})

			// authentication flow
			if !resolver.AuthWhitelisted {
				verified, err := authVerifyToken(r, logger)
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				} else if !verified {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}

			// setting tracking if not set by client already
			if r.Header.Get("X-Track-ID") == "" {
				r.Header.Set("X-Track-ID", trackId)
			}

			reverseProxy.ServeHTTP(w, r)

			logger.WithFields(logrus.Fields{
				"RemoteAddr": r.RemoteAddr,
				"UserAgent":  r.UserAgent(),
				"Method":     r.Method,
				"Duration":   time.Now().Sub(start),
			}).Info(fmt.Sprintf("Successfully redirected %s%s to %s:%v%s",
				AppUrl, resolver.ProxyPath, AppHost, ServiceToDnsResolver[resolver.ServiceName], resolver.ActualPath))
		})
	}

	logrus.Info("App running on port " + AppUrl)
	logrus.Fatal(http.ListenAndServe(AppUrl, router))
}

func authVerifyToken(r *http.Request, logger *logrus.Entry) (bool, error) {
	jwtToken := r.Header.Get("Authorization")
	if strings.HasPrefix(jwtToken, "Bearer ") {
		jwtToken = jwtToken[8:]
	}

	requestBody, err := json.Marshal(map[string]string{
		"jwt": jwtToken,
	})
	if err != nil {
		return false, err
	}

	res, err := http.Post(
		fmt.Sprintf("http://%s:%v/verify", AppHost, AUTH_SERVICE_PORT),
		"application/json",
		bytes.NewBuffer(requestBody),
	)

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		var responseBody map[string]interface{}
		json.NewDecoder(res.Body).Decode(&responseBody)

		return responseBody["verified"].(bool), nil
	} else if res.StatusCode == http.StatusUnauthorized {
		return false, nil
	}

	logger.Warn(fmt.Sprintf("Auth service returning unexpected response: %v", res.StatusCode))
	return false, errors.New("Auth service returns something other than 200; for more info see log")
}
