// Package plugindemo a demo plugin.
package traefik_umami_plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Config the plugin configuration.
type Config struct {
	ForwardPath            string   `json:"forwardPath"`
	UmamiHost              string   `json:"umamiHost"`
	WebsiteId              string   `json:"websiteId"`
	AutoTrack              bool     `json:"autoTrack"`
	DoNotTrack             bool     `json:"doNotTrack"`
	Cache                  bool     `json:"cache"`
	Domains                []string `json:"domains"`
	EvadeGoogleTagManager  bool     `json:"evadeGoogleTagManager"`
	ScriptInjection        bool     `json:"scriptInjection"`
	ScriptInjectionMode    string   `json:"scriptInjectionMode"`
	ServerSideTracking     bool     `json:"serverSideTracking"`
	ServerSideTrackingMode string   `json:"serverSideTrackingMode"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		ForwardPath:            "_umami",
		UmamiHost:              "",
		WebsiteId:              "",
		AutoTrack:              true,
		DoNotTrack:             false,
		Cache:                  false,
		Domains:                []string{},
		EvadeGoogleTagManager:  false,
		ScriptInjection:        true,
		ScriptInjectionMode:    SIModeTag,
		ServerSideTracking:     false,
		ServerSideTrackingMode: SSTModeAll,
	}
}

const (
	SIModeTag          string = "tag"
	SIModeSource       string = "source"
	SSTModeAll         string = "all"
	SSTModeNotinjected string = "notinjected"
)

// PluginHandler a PluginHandler plugin.
type PluginHandler struct {
	next          http.Handler
	name          string
	config        Config
	configIsValid bool
	scriptHtml    string
	LogHandler    *log.Logger
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// construct
	h := &PluginHandler{
		next:          next,
		name:          name,
		config:        *config,
		configIsValid: true,
		scriptHtml:    "",
		LogHandler:    log.New(os.Stdout, "", 0),
	}

	// check if the umami host is set
	if config.UmamiHost == "" {
		h.log("umamiHost is not set!")
		h.configIsValid = false
	}
	// check if the website id is set
	if config.WebsiteId == "" {
		h.log("websiteId is not set!")
		h.configIsValid = false
	}
	// check if scriptInjectionMode is valid
	if config.ScriptInjectionMode != SIModeTag && config.ScriptInjectionMode != SIModeSource {
		h.log("scriptInjectionMode is not valid!")
		h.config.ScriptInjection = false
		h.configIsValid = false
	}
	// check if serverSideTrackingMode is valid
	if config.ServerSideTrackingMode != SSTModeAll && config.ServerSideTrackingMode != SSTModeNotinjected {
		h.log("serverSideTrackingMode is not valid!")
		h.config.ServerSideTracking = false
		h.configIsValid = false
	}

	// build script html
	scriptHtml, err := buildUmamiScript(&h.config)
	h.scriptHtml = scriptHtml
	if err != nil {
		return nil, err
	}

	configJSON, _ := json.Marshal(config)
	h.log(fmt.Sprintf("config: %s", configJSON))
	if config.ScriptInjection {
		h.log(fmt.Sprintf("script: %s", scriptHtml))
	} else {
		h.log("script: scriptInjection is false")
	}

	return h, nil
}

func (h *PluginHandler) log(message string) {
	level := "info" // default to info
	time := time.Now().Format("2006-01-02T15:04:05Z")

	if h.LogHandler != nil {
		h.LogHandler.Println(fmt.Sprintf("time=\"%s\" level=%s msg=\"[traefik-umami-plugin] %s\"", time, level, message))
	}
}

func (h *PluginHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// check if config is valid
	if !h.configIsValid {
		h.next.ServeHTTP(rw, req)
		return
	}

	// forwarding
	shouldForwardToUmami, pathAfter := isUmamiForwardPath(req, &h.config)
	if shouldForwardToUmami {
		// h.log(fmt.Sprintf("Forward %s", req.URL.EscapedPath()))
		h.forwardToUmami(rw, req, pathAfter)
		return
	}

	// script injection
	var injected bool = false
	myReq := &Request{Request: *req}
	myRw := NewResponseWriter(rw)
	if h.config.ScriptInjection && myReq.CouldBeInjectable() {
		// intercept body
		encoding := myReq.GetSupportedEncodings(h).GetPreferred()
		myReq.SetSupportedEncoding(h)
		h.next.ServeHTTP(myRw, &myReq.Request)

		// check if response is injectable
		if myRw.IsInjectable() {
			body, err := myRw.ReadDecoded(h)
			if err != nil {
				h.log(fmt.Sprintf("Error: %s", err))
			}
			newBody := InsertAtBodyEnd(body, h.scriptHtml)
			h.log(fmt.Sprintf("newBody: %s", newBody))
			h.log(fmt.Sprintf("encoding: %+v", encoding))
			myRw.WriteEncoded(newBody, encoding)
			rw.Write(myRw.Read())

			injected = true
			h.next.ServeHTTP(rw, req)
			return
		}
	}

	// // server side tracking
	// shouldServerSideTrack := shouldServerSideTrack(req, &h.config, injected, h)
	// if shouldServerSideTrack {
	// 	// h.log(fmt.Sprintf("Track %s", req.URL.EscapedPath()))
	// 	go buildAndSendTrackingRequest(req, &h.config)
	// }

	// if !injected {
	// 	// h.log(fmt.Sprintf("Continue %s", req.URL.EscapedPath()))
	// 	h.next.ServeHTTP(rw, req)
	// }
}

type responseWriter struct {
	buffer *bytes.Buffer
	http.ResponseWriter
}
