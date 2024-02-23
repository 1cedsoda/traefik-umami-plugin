package traefik_umami_plugin

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// builds the umami script.
func buildUmamiScript(config *Config) (string, error) {
	// check if the script should be injected
	if config.ScriptInjection == false {
		return "", nil
	}

	// download the script
	var scriptJs string
	if config.ScriptInjectionMode == SIModeSource {
		_scriptJs, err := downloadScript(config, context.Background())
		if err != nil {
			return "", err
		}
		scriptJs = _scriptJs
	}

	// src url
	var src string
	if config.ScriptInjectionMode == SIModeTag {
		src = fmt.Sprintf(`/%s/script.js`, config.ForwardPath)
	}

	if config.EvadeGoogleTagManager {
		return buildUmamiScriptWithEvade(config, scriptJs, src), nil
	} else {
		return buildUmamiScriptWithoutEvade(config, scriptJs, src), nil
	}
}

func buildUmamiScriptWithEvade(config *Config, scriptJs, src string) string {
	html := "<script>"
	html += "(function () {"
	html += "var el = document.createElement('script');"
	html += fmt.Sprintf("el.setAttribute('data-host-url', '%s');", config.ForwardPath)
	if config.ScriptInjectionMode == SIModeTag {
		html += fmt.Sprintf("el.setAttribute('src', '%s');", src)
	} else if config.ScriptInjectionMode == SIModeSource {
		scriptBase64 := base64.StdEncoding.EncodeToString([]byte(scriptJs))
		html += "el.setAttribute('type', 'text/javascript');"
		html += fmt.Sprintf("el.innerHTML = atob('%s');", scriptBase64)
	}
	html += fmt.Sprintf("el.setAttribute('data-website-id', '%s');", config.WebsiteId)
	if config.AutoTrack {
		html += "el.setAttribute('data-auto-track', 'true');"
	} else {
		html += "el.setAttribute('data-auto-track', 'false');"
	}
	if config.DoNotTrack {
		html += "el.setAttribute('data-do-not-track', 'true');"
	}
	if config.Cache {
		html += "el.setAttribute('data-cache', 'true');"
	}
	if len(config.Domains) > 0 {
		html += fmt.Sprintf("el.setAttribute('data-domains', '%s');", strings.Join(config.Domains, ","))
	}
	html += "document.body.appendChild(el);"
	html += "})();"
	html += "</script>"
	return html
}

func buildUmamiScriptWithoutEvade(config *Config, scriptJs, src string) string {
	html := "<script"
	html += " async"
	html += " defer"
	html += fmt.Sprintf(" data-host-url='/%s'", config.ForwardPath)
	if config.ScriptInjectionMode == SIModeTag {
		html += fmt.Sprintf(" src='%s'", src)
	}
	html += fmt.Sprintf(" data-website-id='%s'", config.WebsiteId)
	if config.AutoTrack {
		html += " data-auto-track='true'"
	} else {
		html += " data-auto-track='false'"
	}
	if config.DoNotTrack {
		html += " data-do-not-track='true'"
	}
	if config.Cache {
		html += " data-cache='true'"
	}
	if len(config.Domains) > 0 {
		html += fmt.Sprintf(" data-domains='%s'", strings.Join(config.Domains, ","))
	}
	html += ">"
	if config.ScriptInjectionMode == SIModeSource {
		html += scriptJs
	}
	html += "</script>"
	return html
}

func downloadScript(config *Config, ctx context.Context) (string, error) {
	// request
	url := fmt.Sprintf("%s/script.js", config.UmamiHost)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "traefik-umami-plugin")
	req.Header.Set("Accept", "application/javascript")
	req.Header.Set("Accept-Encoding", "identity")

	// make request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// read response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	// return the script
	return string(body), nil
}
