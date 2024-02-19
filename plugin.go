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
	ForwardPath           string   `json:"forwardPath"`
	UmamiHost             string   `json:"umamiHost"`
	WebsiteId             string   `json:"websiteId"`
	AutoTrack             bool     `json:"autoTrack"`
	DoNotTrack            bool     `json:"doNotTrack"`
	Cache                 bool     `json:"cache"`
	Domains               []string `json:"domains"`
	EvadeGoogleTagManager bool     `json:"evadeGoogleTagManager"`
	ScriptInjection       bool     `json:"scriptInjection"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		ForwardPath:           "_umami",
		UmamiHost:             "",
		WebsiteId:             "",
		AutoTrack:             true,
		DoNotTrack:            false,
		Cache:                 false,
		Domains:               []string{},
		EvadeGoogleTagManager: false,
		ScriptInjection:       true,
	}
}

// PluginHandler a PluginHandler plugin.
type PluginHandler struct {
	next       http.Handler
	name       string
	config     Config
	scriptHtml string
	LogHandler *log.Logger
	client     *http.Client
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

	// build script html
	scriptHtml, err := buildUmamiScript(config)
	if err != nil {
		return nil, err
	}

	//set http client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	pluginHandler := &PluginHandler{
		next:       next,
		name:       name,
		config:     *config,
		scriptHtml: scriptHtml,
		LogHandler: log.New(os.Stdout, "", 0),
		client:     client,
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
	h.log(fmt.Sprintf("serve http %s", req.URL.EscapedPath()))

	shouldForwardToUmami, pathAfter := isUmamiForwardPath(req, &h.config)

	// forwarding
	if shouldForwardToUmami {
		h.log(fmt.Sprintf("shouldForwardToUmami: %t", shouldForwardToUmami))
		h.log(fmt.Sprintf("pathAfter: %s", pathAfter))
		h.forwardToUmami(rw, req, pathAfter)
		return
	}

	// script injection
	if h.config.ScriptInjection {
		// intercept body
		rxrw := &responseWriter{
			buffer:         &bytes.Buffer{},
			ResponseWriter: rw,
		}
		h.next.ServeHTTP(rxrw, req)

		if rw.Header().Get("Content-Type") == "text/html" {
			bytes := rxrw.buffer.Bytes()
			newBytes := injectIntoHeader(bytes, h.scriptHtml)
			rw.Write(newBytes)
		}
		return
	}

	h.next.ServeHTTP(rw, req)
}

type responseWriter struct {
	buffer *bytes.Buffer
	http.ResponseWriter
}

func (w *responseWriter) Write(p []byte) (int, error) {
	w.buffer.Reset()
	return w.buffer.Write(p)
}
