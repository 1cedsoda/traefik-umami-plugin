package traefik_umami_plugin

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

// check if the requested URL should be forwaeded to umami
// based on the ForwardPath (eg. /umami)
// only forwards /api/send and /script.js
func isUmamiForwardPath(req *http.Request, config *Config) (bool, string) {
	currentPath := req.URL.EscapedPath()
	pathRegex := fmt.Sprintf(`\/%s\/((?:script\.js)|(?:api\/send))`, config.ForwardPath)
	match := regexp.MustCompile(pathRegex).FindStringSubmatch(currentPath)
	if match != nil {
		pathAfter := match[1]
		return true, pathAfter
	}
	return false, ""
}

// build the new URL to umami
// based on the UmamiHost and pathAfter
func (h *PluginHandler) getForwardUrl(pathAfter string) (string, error) {
	// return path.Join(config.UmamiConfig.UmamiHost, pathAfter)
	urlString := fmt.Sprintf("%s/%s", h.config.UmamiHost, pathAfter)
	// validate the URL
	_, err := url.Parse(urlString)
	// return the URL and error
	return urlString, err
}

// forward the incoming request to umami
// if not 2XX, shortcut and return forward response
// if 2XX, continue to next handler
func (h *PluginHandler) forwardToUmami(rw http.ResponseWriter, req *http.Request, pathAfter string) {
	// build URL
	forwardUrl, err := h.getForwardUrl(pathAfter)
	if err != nil {
		h.log(fmt.Sprintf("h.getForwardUrl: %+v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// build proxy request
	proxyReq, err := newForwardRequest(req, forwardUrl)
	if err != nil {
		h.log(fmt.Sprintf("traefik_plugin_forward_request.NewForwardRequest: %+v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// make proxy request
	proxyRes, err := h.client.Do(proxyReq)
	if err != nil {
		h.log(fmt.Sprintf("h.client.Do: %+v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// build response
	copyHeaders(rw.Header(), proxyRes.Header)
	removeHeaders(rw.Header(), hopHeaders...)
	rw.WriteHeader(proxyRes.StatusCode)
	body, err := io.ReadAll(proxyRes.Body)
	if err != nil {
		h.log(fmt.Sprintf("io.ReadAll: %+v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.Write(body)
}
