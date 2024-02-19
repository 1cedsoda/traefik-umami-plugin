// Package plugindemo a demo plugin.
package traefik_umami_plugin

import (
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
	DataCache             bool     `json:"dataCache"`
	DataDomains           []string `json:"dataDomains"`
	EvadeGoogleTagManager bool     `json:"evadeGoogleTagManager"`
	InjectScript          bool     `json:"injectScript"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		ForwardPath:           "_umami",
		UmamiHost:             "",
		WebsiteId:             "",
		AutoTrack:             true,
		DoNotTrack:            false,
		DataCache:             false,
		DataDomains:           []string{},
		EvadeGoogleTagManager: false,
		InjectScript:          true,
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
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
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
	if config.InjectScript {
		pluginHandler.log(fmt.Sprintf("script: %s", scriptHtml))
	} else {
		pluginHandler.log("script: injectScript is false")
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
	// h.log(fmt.Sprintf("serve http %s", req.URL.EscapedPath()))
	h.next.ServeHTTP(rw, req)
	// shouldForwardToUmami, pathAfter := isUmamiForwardPath(req, &h.config)
	// h.log(fmt.Sprintf("shouldForwardToUmami: %t", shouldForwardToUmami))
	// h.log(fmt.Sprintf("pathAfter: %s", pathAfter))
	// if shouldForwardToUmami {
	// 	log.Println("forwarding to umami")
	// 	h.forwardToUmami(rw, req, pathAfter)
	// 	return
	// }
	// if h.config.InjectScript {
	// 	h.log("injecting script?")
	// 	writer := &MyResponseWriter{
	// 		buffer:         &bytes.Buffer{},
	// 		ResponseWriter: rw,
	// 	}
	// 	h.next.ServeHTTP(writer, req)
	// 	// if content type is text/html
	// 	if req.Header.Get("Content-Type") == "text/html" {
	// 		injectIntoHeader(writer, &h.scriptHtml)
	// 		h.log("injecting script!")
	// 	}
	// }
}
