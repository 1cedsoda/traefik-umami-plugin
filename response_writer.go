package traefik_umami_plugin

import (
	"bytes"
	"net/http"
)

type ResponseWriter struct {
	buffer *bytes.Buffer
	http.ResponseWriter
}

func NewResponseWriter(rw http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		buffer:         &bytes.Buffer{},
		ResponseWriter: rw,
	}
}

func (w *ResponseWriter) IsInjectable() bool {
	return w.Header().Get("Content-Type") == "text/html"
}

// Body bytes
// Might be compressed.
func (w *ResponseWriter) Read() []byte {
	return w.buffer.Bytes()
}

// Body bytes
// Always uncompressed
// Error if encoding is not supported.
func (w *ResponseWriter) ReadDecoded() ([]byte, error) {
	encoding, err := w.GetContentEncoding()
	if err != nil {
		return nil, err
	}
	return Decode(w.buffer, encoding)
}

// Write body bytes.
func (w *ResponseWriter) Write(p []byte) (int, error) {
	w.buffer.Reset()
	return w.buffer.Write(p)
}

// Write body bytes
// Compresses the body to the target encoding.
func (w *ResponseWriter) WriteEncoded(plain []byte, encoding *Encoding) (int, error) {
	encoded, err := Encode(plain, encoding)
	if err != nil {
		return 0, err
	}
	w.Write(encoded)
	w.SetContentEncoding(encoding)
	return len(plain), nil
}

// Content-Encoding header.
func (w *ResponseWriter) GetContentEncoding() (*Encoding, error) {
	str := w.Header().Get("Content-Encoding")
	return ParseEncoding(str)
}

// Set Content-Encoding header.
func (w *ResponseWriter) SetContentEncoding(encoding *Encoding) {
	w.Header().Set("Content-Encoding", encoding.name)
}
