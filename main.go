package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/hashicorp/go-hclog"
)

func main() {
	hlog := hclog.L()
	addr := lookupAddr()
	handler := handlers.CustomLoggingHandler(nil, http.HandlerFunc(rootHandler), requestLogger(hlog))

	http.Handle("/", handler)

	hlog.Info("Starting server on " + addr)

	err := http.ListenAndServe(addr, nil)

	if err != nil {
		hlog.Error(err.Error())
	}
}

func lookupAddr() string {
	addr, ok := os.LookupEnv("ADDR")

	if ok {
		return addr
	} else {
		return ":8080"
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		fmt.Fprintf(w, "pong")
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func requestLogger(hlog hclog.Logger) func(_ io.Writer, params handlers.LogFormatterParams) {
	return func(_ io.Writer, params handlers.LogFormatterParams) {
		req := params.Request

		// Extract the Client IP honoring the X-Forwarded-For header set by
		// proxies.
		clientIP, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			clientIP = req.RemoteAddr
		}
		if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			clientIP = forwardedFor
		}

		// Extract the URL scheme honoring the X-Forwarded-Proto header set by
		// proxies.
		scheme := req.URL.Scheme
		if forwardedProto := req.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
			scheme = forwardedProto
		}

		hlog.Info(
			fmt.Sprintf("HTTP request: %s %s", req.Method, req.URL.Path),
			"date", params.TimeStamp.Format(time.RFC3339Nano),
			"http.host", req.Host,
			"http.method", req.Method,
			"http.request_path", req.URL.Path,
			"http.remote_addr", clientIP,
			"http.response_size", params.Size,
			"http.scheme", scheme,
			"http.status_code", params.StatusCode,
			"http.useragent", req.UserAgent(),
			"http.version", req.Proto,
		)
	}
}
