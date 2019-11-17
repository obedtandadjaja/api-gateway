package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

type PathResolver struct {
	Path         Path
	ReverseProxy *httputil.ReverseProxy
}

var Environment string
var AppHost string
var AppPort string
var AppUrl string
var PathResolvers = []PathResolver{}

func init() {
	Environment = os.Getenv("ENV")
	AppHost = os.Getenv("APP_HOST")
	AppPort = os.Getenv("APP_PORT")
	AppUrl = AppHost + ":" + AppPort

	// set up the paths and reverseProxy -> PathResolver
	for _, path := range Paths {
		baseUrlString := fmt.Sprintf("http://%s", AppUrl)
		baseUrl, err := url.Parse(baseUrlString)
		if err != nil {
			panic(fmt.Sprintf("Cannot parse url: %s", baseUrlString))
		}

		reverseProxy := httputil.NewSingleHostReverseProxy(baseUrl)
		reverseProxy.Director = func(r *http.Request) {
			r.Header.Add("X-Forwarded-Host", r.Host)
			r.Header.Add("X-Origin-Host", AppUrl)

			r.URL.Host = path.ServiceName
			r.URL.Scheme = baseUrl.Scheme
			r.URL.Path = path.ActualPath
		}
		reverseProxy.ModifyResponse = func(r *http.Response) error {
			if r.StatusCode >= 500 {
				logrus.Warn(err)

				buf := bytes.NewBufferString("Internal Server Error")
				r.Body = ioutil.NopCloser(buf)
				r.Header["Content-Length"] = []string{fmt.Sprint(buf.Len())}
				r.Header["Content-Type"] = []string{"text/plain"}
				r.StatusCode = 500
			}

			return nil
		}

		PathResolvers = append(PathResolvers, PathResolver{path, reverseProxy})
	}
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	router := httprouter.New()
	router.GET("/api/health", api.Health)

	for i := range PathResolvers {
		router.Handle(PathResolvers[i].Path.Method, PathResolvers[i].Path.ProxyPath, PathResolvers[i].resolve)
	}

	logrus.Info("App running on port " + AppUrl)
	logrus.Fatal(http.ListenAndServe(AppUrl, router))
}

func (resolver *PathResolver) resolve(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	start := time.Now()

	trackId, _ := helper.GenerateRandomString(12)
	logger := logrus.WithFields(logrus.Fields{
		"TrackId": trackId,
	})

	// authentication flow
	if !resolver.Path.AuthWhitelisted {
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

	resolver.ReverseProxy.ServeHTTP(w, r)

	logger.WithFields(logrus.Fields{
		"RemoteAddr": r.RemoteAddr,
		"UserAgent":  r.UserAgent(),
		"Method":     r.Method,
		"Duration":   time.Now().Sub(start).Milliseconds(),
	}).Info(fmt.Sprintf("Successfully redirected %s%s to %s:%v%s",
		AppUrl, resolver.Path.ProxyPath, AppHost, ServiceToDnsResolver[resolver.Path.ServiceName], resolver.Path.ActualPath))
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
		fmt.Sprintf("http://%s/verify", AUTH_SERVICE),
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
