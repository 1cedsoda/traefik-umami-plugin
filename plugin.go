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
	next       http.Handler
	name       string
	config     Config
	scriptHtml string
	LogHandler *log.Logger
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// check if the umami host is set
	if config.UmamiHost == "" {
		return nil, fmt.Errorf("umami host is not set")
	}
	// check if the website id is set
	if config.WebsiteId == "" {
		return nil, fmt.Errorf("website id is not set")
	}
	// check if scriptInjectionMode is valid
	if config.ScriptInjectionMode != SIModeTag && config.ScriptInjectionMode != SIModeSource {
		return nil, fmt.Errorf("scriptInjectionMode is not valid")
	}
	// check if serverSideTrackingMode is valid
	if config.ServerSideTrackingMode != SSTModeAll && config.ServerSideTrackingMode != SSTModeNotinjected {
		return nil, fmt.Errorf("serverSideTrackingMode is not valid")
	}

	// construct
	pluginHandler := &PluginHandler{
		next:       next,
		name:       name,
		config:     *config,
		scriptHtml: "",
		LogHandler: log.New(os.Stdout, "", 0),
	}

	// build script html
	scriptHtml, err := buildUmamiScript(&pluginHandler.config)
	pluginHandler.scriptHtml = scriptHtml
	if err != nil {
		return nil, err
	}

	configJSON, _ := json.Marshal(config)
	pluginHandler.log(fmt.Sprintf("config: %s", configJSON))
	if config.ScriptInjection {
		pluginHandler.log(fmt.Sprintf("script: %s", scriptHtml))
	} else {
		pluginHandler.log("script: scriptInjection is false")
	}

	return pluginHandler, nil
}

func (h *PluginHandler) log(message string) {
	level := "info" // default to info
	time := time.Now().Format("2006-01-02T15:04:05Z")

	if h.LogHandler != nil {
		h.LogHandler.Println(fmt.Sprintf("time=\"%s\" level=%s msg=\"[traefik-umami-plugin] %s\"", time, level, message))
	}
}

func (h *PluginHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// forwarding
	shouldForwardToUmami, pathAfter := isUmamiForwardPath(req, &h.config)
	if shouldForwardToUmami {
		// h.log(fmt.Sprintf("Forward %s", req.URL.EscapedPath()))
		h.forwardToUmami(rw, req, pathAfter)
		return
	}

	// script injection
	var injected bool = false
	if h.config.ScriptInjection {
		// intercept body
		myrw := &responseWriter{
			buffer:         &bytes.Buffer{},
			ResponseWriter: rw,
		}
		h.next.ServeHTTP(myrw, req)

		if myrw.Header().Get("Content-Type") == "text/html" {
			// h.log(fmt.Sprintf("Inject %s", req.URL.EscapedPath()))
			bytes := myrw.buffer.Bytes()
			newBytes := regexReplaceSingle(bytes, insertBeforeRegex, h.scriptHtml)
			rw.Write(newBytes)
			injected = true
		}
	}

	// server side tracking
	shouldServerSideTrack := shouldServerSideTrack(req, &h.config, injected, h)
	if shouldServerSideTrack {
		// h.log(fmt.Sprintf("Track %s", req.URL.EscapedPath()))
		go buildAndSendTrackingRequest(req, &h.config)
	}

	if !injected {
		// h.log(fmt.Sprintf("Continue %s", req.URL.EscapedPath()))
		h.next.ServeHTTP(rw, req)
	}
}

type responseWriter struct {
	buffer *bytes.Buffer
	http.ResponseWriter
}

func (w *responseWriter) Write(p []byte) (int, error) {
	w.buffer.Reset()
	return w.buffer.Write(p)
}
