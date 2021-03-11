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
	"sync"
	"time"
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
	mux := http.NewServeMux()
	mux.HandleFunc("/session", sessionHandler)
	mux.HandleFunc("/cn/session", sessionHandler)
	mux.HandleFunc("/eu/session", sessionHandler)
	mux.HandleFunc("/ru/session", sessionHandler)

	mux.HandleFunc("/impression/", impressionHandler)
	mux.HandleFunc("/", clickHandler)

	router := loggingMiddleware(log.Default(), mux)
	router = dataResidencyMiddleware(router)

	return router
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
	if err := handleDataResidency(w, r, prefix); err != nil {
		return
	}
	io.WriteString(w, fmt.Sprintf("%s: hello %s\n", prefix, r.URL.Path))
}

func handleDataResidency(w http.ResponseWriter, r *http.Request, prefix string) error {
	queryRegion := r.FormValue("region")
	if queryRegion == "" {
		return nil
	}
	if strings.HasPrefix(r.URL.Path, fmt.Sprintf("/%s/", queryRegion)) {
		return nil
	}

	drUrlTarget := fmt.Sprintf("http://%s/%s%s", httpAddr, queryRegion, r.URL.Path)
	w.Header().Set("X-Accel-Redirect", drUrlTarget)
	w.WriteHeader(http.StatusUnavailableForLegalReasons)

	return fmt.Errorf("%s: found data residency for %v", prefix, r)
}

func loggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Printf("%s %s %s %s", r.Method, r.RequestURI, r.Header, time.Since(start))
	})
}

func dataResidencyMiddleware(next http.Handler) http.Handler {
	director := func(r *http.Request) {
		if _, ok := r.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			r.Header.Set("User-Agent", "")
		}
	}
	proxy := &httputil.ReverseProxy{
		Director: director,
	}
	return &DataResidencyMiddleware{
		proxy:   proxy,
		handler: next,
	}
}

type DataResidencyMiddleware struct {
	proxy   *httputil.ReverseProxy
	handler http.Handler
}

func (h *DataResidencyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw := newResponseWriter(w, nil)

	var bodySnap io.ReadCloser
	if r.Body != nil {
		buf := bufPool.Get().(*bytes.Buffer)
		defer bufPool.Put(buf)
		buf.Reset()
		_, err := io.Copy(buf, r.Body)
		if err != nil {
			return
		}
		if err := r.Body.Close(); err != nil {
			return
		}
		bodySnap = ioutil.NopCloser(buf)
		r.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	}

	h.handler.ServeHTTP(rw, r)

	uri := w.Header().Get("X-Accel-Redirect")
	if uri == "" {
		rw.WriteResponse()
		return
	}
	u, _ := url.ParseRequestURI(uri)
	r = r.Clone(r.Context())
	r.URL.Scheme = u.Scheme
	r.URL.Host = u.Host
	r.URL.Path = u.Path
	r.Body = bodySnap
	h.proxy.ServeHTTP(w, r)
}

type drResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	stream      bool
	buf         *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter, buf *bytes.Buffer) *drResponseWriter {
	if buf == nil {
		buf = new(bytes.Buffer)
	}
	return &drResponseWriter{
		ResponseWriter: w,
		buf:            buf,
	}
}

func (rw *drResponseWriter) WriteHeader(statusCode int) {
	if rw.wroteHeader {
		return
	}
	rw.statusCode = statusCode
	rw.wroteHeader = true

	rw.stream = !rw.shouldBuffer(statusCode, rw.Header())

	// if not buffering the response, stream it to the underlying ResponseWriter
	if rw.stream {
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *drResponseWriter) Write(data []byte) (int, error) {
	rw.WriteHeader(http.StatusOK)
	if rw.stream {
		return rw.ResponseWriter.Write(data)
	} else {
		return rw.buf.Write(data)
	}
}

func (rw *drResponseWriter) WriteResponse() error {
	if rw.stream || !rw.wroteHeader {
		return nil
	}
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	rw.ResponseWriter.WriteHeader(rw.statusCode)
	_, err := io.Copy(rw.ResponseWriter, rw.buf)
	return err
}

func (rw *drResponseWriter) shouldBuffer(statusCode int, headers http.Header) bool {
	if statusCode < 400 || statusCode > 500 {
		return false
	}
	if headers.Get("X-Accel-Redirect") == "" {
		return false
	}
	return true
}

// bufPool is used for buffering requests and responses.
var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
