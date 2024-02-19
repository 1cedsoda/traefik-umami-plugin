package traefik_umami_plugin

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	traefik_plugin_forward_request "github.com/kzmake/traefik-plugin-forward-request"
)

// check if the requested URL should be forwaeded to umami
// based on the ForwardPath (eg. /umami)
func isUmamiForwardPath(req *http.Request, config *Config) (bool, string) {
	currentPath := req.URL.EscapedPath()
	pathRegex := fmt.Sprintf(`^\/%s(\/)?(.+)?`, config.ForwardPath)
	match := regexp.MustCompile(pathRegex).FindStringSubmatch(currentPath)
	if match != nil {
		pathAfter := match[2]
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
		h.log(fmt.Sprintf("forward url error: %s", err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.log(fmt.Sprintf("forward url: %s", forwardUrl))

	// build request
	fReq, err := traefik_plugin_forward_request.NewForwardRequest(req, forwardUrl)
	if err != nil {
		h.log(fmt.Sprintf("build request error: %+v", err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.log(fmt.Sprintf("build request: %+v", fReq))

	// make request
	fRes, err := h.client.Do(fReq)
	if err != nil {
		h.log(fmt.Sprintf("do request error: %s", err))
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.log(fmt.Sprintf("response: %+v", fRes))

	// not 2XX -> return forward response
	if fRes.StatusCode < http.StatusOK || fRes.StatusCode >= http.StatusMultipleChoices {
		writeForwardResponse(rw, fRes)
		return
	}

	// 2XX -> next
	traefik_plugin_forward_request.OverrideHeaders(req.Header, fRes.Header)
	h.next.ServeHTTP(rw, req)
	return
}

// response to client after forwarding to umami
func writeForwardResponse(rw http.ResponseWriter, fRes *http.Response) {
	body, err := ioutil.ReadAll(fRes.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer fRes.Body.Close()

	traefik_plugin_forward_request.CopyHeaders(rw.Header(), fRes.Header)
	traefik_plugin_forward_request.RemoveHeaders(rw.Header(), traefik_plugin_forward_request.HopHeaders...)

	// Grab the location header, if any.
	redirectURL, err := fRes.Location()

	if err != nil {
		if err != http.ErrNoLocation {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if redirectURL.String() != "" {
		// Set the location in our response if one was sent back.
		rw.Header().Set("Location", redirectURL.String())
	}

	rw.WriteHeader(fRes.StatusCode)
	_, _ = rw.Write(body)
}
