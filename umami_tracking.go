package traefik_umami_plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type SendPayload struct {
	Website  string                 `json:"website"`
	Hostname string                 `json:"hostname"`
	Language string                 `json:"language"`
	Url      string                 `json:"url"`
	Referer  string                 `json:"referer"`
	Name     string                 `json:"name"`
	Data     map[string]interface{} `json:"data"`
}

type SendBody struct {
	Payload SendPayload `json:"payload"`
	Type    string      `json:"type"`
}

func buildSendPayload(req *http.Request, websiteId string) SendPayload {
	return SendPayload{
		Website:  websiteId,
		Hostname: parseDomainFromHost(req.Host),
		Language: parseAcceptLanguage(req.Header.Get("Accept-Language")),
		Url:      req.URL.String(),
		Referer:  req.Referer(),
		Name:     "traefik",
		Data:     map[string]interface{}{},
	}
}

func parseAcceptLanguage(acceptLanguage string) string {
	const parseAcceptLanguagePattern = `([a-zA-Z\-]+)(?:;q=\d\.\d)?(?:,\s)?`
	parseAcceptLanguageRegexp := regexp.MustCompile(parseAcceptLanguagePattern)
	matches := parseAcceptLanguageRegexp.FindAllStringSubmatch(acceptLanguage, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[0][1]
}

func buildTrackingRequest(clientReq *http.Request, config *Config) (*http.Request, error) {
	// build body
	sendBody := SendBody{
		Payload: buildSendPayload(clientReq, config.WebsiteId),
		Type:    "event",
	}
	bodyJson, err := json.Marshal(sendBody)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(bodyJson)

	// build url
	url := fmt.Sprintf("%s/api/send", config.UmamiHost)

	// build request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, err
	}

	// set headers
	req.Header.Set("Content-Type", "application/json")
	copyHeaders(req.Header, clientReq.Header)
	removeHeaders(req.Header, hopHeaders...)
	writeXForwardedHeaders(req.Header, clientReq)

	return req, nil
}

// send the tracking request to umami's /api/send.
func (h *PluginHandler) sendTrackingRequest(trackingReq *http.Request) error {
	// make request
	trackingRes, err := h.client.Do(trackingReq)
	if err != nil {
		h.log(fmt.Sprintf("Error: %s", err))
		return err
	}

	status := trackingRes.StatusCode
	if status < 200 || status >= 300 {
		h.log(fmt.Sprintf("Tracking request status: %d", status))
	}

	return nil
}

// opts the port from the host.
func parseDomainFromHost(host string) string {
	// check if the host has a port
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	return host
}

// check if the requested domain is in the list of domains
// if the list is empty, return true.
func hostnameInDomains(req *http.Request, domains []string) bool {
	if len(domains) == 0 {
		return true
	}
	hostname := parseDomainFromHost(req.Host)
	for _, domain := range domains {
		if domain == hostname {
			return true
		}
	}
	return false
}

// check if server side tracking should be done.
func shouldServerSideTrack(req *http.Request, config *Config) bool {
	if config.ServerSideTracking && hostnameInDomains(req, config.Domains) {
		return true
	}
	return false
}
