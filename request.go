package traefik_umami_plugin

import (
	"net/http"
	"strings"
)

type Request struct {
	http.Request
}

func (req *Request) SetSupportedEncoding() {
	acceptEncoding := ParseEncodings(req.Header.Get("Accept-Encoding"))
	supported := acceptEncoding.FilterSupported().String()
	req.Header.Set("Accept-Encoding", supported)
}

func (req *Request) GetPreferredSupportedEncoding() *Encoding {
	acceptEncoding := req.Header.Get("Accept-Encoding")
	return ParseEncodings(acceptEncoding).FilterSupported().GetPreferred()
}

func (req *Request) CouldBeInjectable() bool {
	// return false on non-GET requests
	if req.Method != http.MethodGet {
		return false
	}

	// ignore websockets
	if strings.Contains(req.Header.Get("Upgrade"), "websocket") {
		return false
	}

	return true
}

func (req *Request) IsHtml() bool {
	return strings.Contains(req.Header.Get("Accept"), "text/html")
}
