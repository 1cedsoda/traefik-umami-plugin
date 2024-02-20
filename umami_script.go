package traefik_umami_plugin

import (
	"fmt"
	"regexp"
	"strings"
)

const headRegexPattern = `</head>`

var headRegexp = regexp.MustCompile(headRegexPattern)

// injects the umami script into the response head
func regexReplaceSingle(bytes []byte, match *regexp.Regexp, replace string) []byte {
	rx := match.FindIndex(bytes)
	if len(rx) == 0 {
		return bytes
	}
	// insert the script before the head tag
	return append(bytes[:rx[0]], append([]byte(replace), bytes[rx[0]:]...)...)
}

// builds the umami script
func buildUmamiScript(config *Config) (string, error) {
	if config.ScriptInjection == false {
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
		if config.AutoTrack {
			html += "el.setAttribute('data-auto-track', 'true');"
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
	} else {
		html += "<script"
		html += " async"
		html += " defer"
		html += fmt.Sprintf(" src='%s'", src)
		html += fmt.Sprintf(" data-website-id='%s'", config.WebsiteId)
		if config.AutoTrack {
			html += " data-auto-track='true'"
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
		html += "></script>"
	}
	return html, nil
}
