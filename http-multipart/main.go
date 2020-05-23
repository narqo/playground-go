package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"os"
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

	http.HandleFunc("/batch", handleBatch)

	log.Println("listening", addr)

	return http.ListenAndServe(addr, nil)
}

// Refer to following examples of multipart/* requests:
// - https://developers.google.com/drive/api/v3/batch
// - https://docs.microsoft.com/en-us/previous-versions/office/developer/exchange-server-2010/aa493937(v=exchg.140)
func handleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// dummyReq is a prototype for the sub-requests from multipart chunks.
	// We don't use Request.Clone to prevent cloning incoming request's headers and form;
	// we don't use shallow clone to prevent cloning fields, that net/http might use internally.
	// Refering to https://godoc.org/net/http
	// "To create a request for use with (testing) a Server Handler, either use
	// the NewRequest function in the net/http/httptest package, use ReadRequest,
	// or manually update the Request fields."
	dummyReq := &http.Request{
		Method:     r.Method,
		RequestURI: r.RequestURI,
		Proto:      r.Proto,
		ProtoMajor: r.ProtoMajor,
		ProtoMinor: r.ProtoMinor,
		Host:       r.Host,
		RemoteAddr: r.RemoteAddr,
	}

	ctx := r.Context()

	mw := multipart.NewWriter(w)
	w.Header().Set("Content-Type", "multipart/mixed; boundary="+mw.Boundary())
	defer func() {
		if err := mw.Close(); err != nil {
			log.Printf("close multipart writer: %s\n", err)
		}
	}()

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var req *http.Request
		if p.Header.Get("Content-Type") == "application/http" {
			// special case: if part is application/http, treat it as a self-contained request
			// TODO(narqo): reply with application/http
			b := bufio.NewReader(p)
			req, err = http.ReadRequest(b)
		} else {
			// WithContext shallow clones dummy request
			req = dummyReq.WithContext(ctx)
			req.Header = http.Header(p.Header)
			req.Body = p
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// pass top-most forwarded IP to current request
		req.Header.Set("X-Forwarded-For", r.Header.Get("X-Forwarded-For"))

		if err := handle(req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h := textproto.MIMEHeader{
			"Content-Type": {"application/json"},
		}
		pw, err := mw.CreatePart(h)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.WriteString(pw, `{"status": "ok"}`)
	}
}

func handle(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	logIncoming(r)

	return nil
}

func logIncoming(r *http.Request) {
	r.ParseForm()

	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}

	remoteAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteAddr = r.RemoteAddr
	}

	clientIP := r.FormValue("ip_address")
	if clientIP == "" {
		clientIP = r.Header.Get("X-Forwarded-For")
	}

	log.Printf(
		"incoming method=%s uri=%s host=%s ip=%s client-ip=%s ua=%s header=%s form=%s\n",
		r.Method,
		r.RequestURI,
		host,
		remoteAddr,
		clientIP,
		r.UserAgent(),
		r.Header,
		r.PostForm,
	)
}
