package traefik_umami_plugin

import (
	"fmt"
	"regexp"
	"strings"
)

const headRegexPattern = `</head>`

var headRx = regexp.MustCompile(headRegexPattern)

// injects the umami script into the response head
func injectIntoHeader(bytes []byte, script string) []byte {
	rx := headRx.FindIndex(bytes)
	if len(rx) == 0 {
		return bytes
	}
	// insert the script before the head tag
	return append(bytes[:rx[0]], append([]byte(script), bytes[rx[0]:]...)...)
}

// builds the umami script
func buildUmamiScript(config *Config) (string, error) {
	if config.InjectScript == false {
		return "", nil
	}
	src := fmt.Sprintf(`/%s/script.js`, config.ForwardPath)
	html := ""
	if config.EvadeGoogleTagManager {
		html += "<script>"
		html += "(function () {"
		html += "var el = document.createElement('script');"
		html += fmt.Sprintf("el.setAttribute('src', '%s');", src)
		html += fmt.Sprintf("el.setAttribute('data-website-id', '%s');", config.WebsiteId)
		html += fmt.Sprintf("el.setAttribute('data-auto-track', '%t');", config.AutoTrack)
		html += fmt.Sprintf("el.setAttribute('data-do-not-track', '%t');", config.DoNotTrack)
		html += fmt.Sprintf("el.setAttribute('data-cache', '%t');", config.DataCache)
		if len(config.DataDomains) > 0 {
			html += fmt.Sprintf("el.setAttribute('data-domains', '%s');", strings.Join(config.DataDomains, ","))
		}
		html += "el.setAttribute('data-do-not-track', 'true');"
		if config.AutoTrack {
			html += "el.setAttribute('data-auto-track', 'true');"
		}

		html += "document.body.appendChild(el);"
		html += "})();"
		html += "</script>"
	} else {
		html += "<script"
		html += " async"
		html += " defer"
		html += fmt.Sprintf(" src='%s'", src)
		html += fmt.Sprintf(" data-website-id='%s'", config.WebsiteId)
		html += fmt.Sprintf(" data-auto-track='%t'", config.AutoTrack)
		html += fmt.Sprintf(" data-do-not-track='%t'", config.DoNotTrack)
		html += fmt.Sprintf(" data-cache='%t'", config.DataCache)
		if len(config.DataDomains) > 0 {
			html += fmt.Sprintf(" data-domains='%s'", strings.Join(config.DataDomains, ","))
		}
		html += "></script>"
	}
	return html, nil
}
