package traefik_umami_plugin

import (
	"bytes"
	"net/http"
	"regexp"
)

type MyResponseWriter struct {
	buffer *bytes.Buffer
	http.ResponseWriter
}

func (w *MyResponseWriter) Write(p []byte) (int, error) {
	return w.buffer.Write(p)
}

func (w *MyResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *MyResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// regex body replacer.
func (w *MyResponseWriter) RegexReplaceBody(regex, replace string) {
	body := w.buffer.String()
	body = regexp.MustCompile(regex).ReplaceAllString(body, replace)
	w.buffer.Reset()
	w.buffer.Write([]byte(body))
}
