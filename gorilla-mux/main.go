package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const httpAddr = "127.0.0.1:8080"

func main() {
	router := setupRouter()

	log.Printf("listening %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, router); err != nil {
		log.Fatal(err)
	}
}

func setupRouter() http.Handler {
	topRouter := mux.NewRouter()
	topRouter.Use(loggingMiddleware(log.Default()))

	sdkRouter := topRouter.NewRoute().Subrouter()
	sdkRouter.Use(dataResidencyMiddleware())

	sdkRouter.HandleFunc("/session", sessionHandler)
	sdkRouter.HandleFunc("/cn/session", sessionHandler)
	sdkRouter.HandleFunc("/eu/session", sessionHandler)
	sdkRouter.HandleFunc("/ru/session", sessionHandler)

	topRouter.Path("/impression/{token}").HandlerFunc(impressionHandler)
	topRouter.Path("/{token}").HandlerFunc(clickHandler)

	return topRouter
}

func clickHandler(w http.ResponseWriter, r *http.Request) {
	writeResp(w, r, "click")
}

func impressionHandler(w http.ResponseWriter, r *http.Request) {
	writeResp(w, r, "impression")
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	writeResp(w, r, "session")
}

func writeResp(w http.ResponseWriter, r *http.Request, prefix string) {
	io.WriteString(w, fmt.Sprintf("%s: hello %s\n", prefix, r.URL.Path))
}

func loggingMiddleware(logger *log.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Printf("%s %s %s %s", r.Method, r.RequestURI, r.Header, time.Since(start))
		})
	}
}

const (
	httpHeaderDRRegion = "X-Dr-Region"
)

func dataResidencyMiddleware() mux.MiddlewareFunc {
	targetUrl, _ := url.Parse(fmt.Sprintf("http://%s", httpAddr))
	director := func(r *http.Request) {
		r.URL.Scheme = targetUrl.Scheme
		r.URL.Host = targetUrl.Host
		r.URL.Path = fmt.Sprintf("/%s%s", r.Header.Get(httpHeaderDRRegion), r.URL.Path)

		if _, ok := r.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			r.Header.Set("User-Agent", "")
		}
	}
	proxy := &httputil.ReverseProxy{
		Director: director,
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			drRegion := r.Header.Get(httpHeaderDRRegion)
			if strings.HasPrefix(r.URL.Path, fmt.Sprintf("/%s/", drRegion)) {
				next.ServeHTTP(w, r)
				return
			}

			body, _ := ioutil.ReadAll(r.Body)
			if body != nil {
				r.Body = ioutil.NopCloser(bytes.NewReader(body))
			}
			queryRegion := r.FormValue("region")
			if queryRegion != "" {
				r.Header.Set(httpHeaderDRRegion, queryRegion)
				// restore request body if needed
				if body != nil {
					r.Body = ioutil.NopCloser(bytes.NewReader(body))
				}
				proxy.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
