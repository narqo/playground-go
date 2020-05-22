package main

import (
	"context"
	"flag"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, args []string) error {
	var addr string

	flags := flag.NewFlagSet("", flag.ExitOnError)
	flags.StringVar(&addr, "addr", "localhost:8080", "address to listen")

	if err := flags.Parse(args); err != nil {
		return err
	}

	http.HandleFunc("/", handleRoot)

	log.Println("listening", addr)

	return http.ListenAndServe(addr, nil)
}

// Refer to
// - https://developers.google.com/drive/api/v3/batch
// - https://docs.microsoft.com/en-us/previous-versions/office/developer/exchange-server-2010/aa493937(v=exchg.140)
func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO(narqo): reply with multipart response (https://peter.bourgon.org/blog/2019/02/12/multipart-http-responses.html)
	// mw := multipart.NewWriter(w)
	// w.Header().Set("Content-Type", mw.FormDataContentType())

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return
			} else if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// https://godoc.org/net/http#NewRequestWithContext
			// To create a request for use with testing a Server Handler, either use
			// the NewRequest function in the net/http/httptest package, use ReadRequest,
			// or manually update the Request fields.
			// XXX(narqo): there should be more (memory) effetient way of doing that
			req := new(http.Request)
			req.Method = r.Method
			req.RequestURI = r.RequestURI
			req.Proto = r.Proto
			req.ProtoMajor = r.ProtoMajor
			req.ProtoMinor = r.ProtoMinor
			req.RemoteAddr = r.RemoteAddr
			req.Host = r.Host
			req.Header = http.Header(p.Header)
			req.Body = p

			// pass top-most IP to the underlying request
			req.Header.Set("X-Forwarded-For", r.Header.Get("X-Forwarded-For"))

			if err := handleOne(req.WithContext(r.Context())); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

func handleOne(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	ip := getDeviceIp(r)
	logIncoming(r, "one", ip)

	return nil
}

func getDeviceIp(r *http.Request) string {
	ip := r.FormValue("ip_address")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	return ip
}

func logIncoming(r *http.Request, kind, ip string) {
	r.ParseForm()
	log.Printf(
		"%s incoming method %s, ua %s, ip %s, client-ip %s, req %s %s",
		kind,
		r.Method,
		r.UserAgent(),
		r.RemoteAddr,
		ip,
		r.URL,
		r.Header,
	)
}
