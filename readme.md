# traefik-umami-plugin

Use [Umami Analytics]() with the [Traefik Reverse Proxy]().

This plugin provides a middleware to provide umami anytics to a web servive.

Pros:
- No need to modify the web service
- Harder to block by adblockers (same domain)
- No need for JavaScript Loading (source injection mode)
- No need for JavaScript (soon: server side tracking)

# Features
- [X] Script Tag Injection - Inject the `script.js` as a script tag
- [X] Script Source Injection - Inject the `script.js` as raw JS code
- [X] Request Forwarding - Forward all requests behind `forwardingPath` to th eunami server
- [ ] Server Side Tracking - Trigger tracking event from the plugin, No JS needed.

# Configuration

To add this plugin to traefik reference this repository as a plugin in the static config.
The version references a git tag.

```yaml
experimental:
  plugins:
    traefik-umami-plugin:
      moduleName: "github.com/1cedsoda/traefik-umami-plugin"
      version: "v1.0.0" 
```
```toml
[experimental.plugins.traefik-umami-plugin]
  moduleName = "github.com/1cedsoda/traefik-umami-plugin"
  version = "v1.0.0"
```
With the plugin installed, you can configure a middleware in a dynamic configuration such as a `config.yml` or docker labels.

```yaml
http:
  middlewares:
    my-umami-middleware:
      plugin:
        traefik-umami-plugin:
          siteId: "umami-site-id"
          scriptUrl: "https://umami.example.com/umami.js"
```
```toml
[http.middlewares]
  [http.middlewares.umami.plugin.traefik-umami-plugin]
    umamiHost = "umami:3000"
    websiteId = "d4617504-241c-4797-8eab-5939b367b3ad"
```
Inside the `traefik-umami-plugin` object the plugin can be configured with the following options:

| key         | default | type     | description                                                                    |
| ----------- | ------- | -------- | ------------------------------------------------------------------------------ |
| `umamiHost` | -       | `string` | Umami server host, reachable from within traefik (container). eg. `umami:3000` |
| `websiteId` | -       | `string` | Website ID as configured in umami.                                             |

The middleware can then be used in a router. Remember to reference the correct provider.

## Request Forwarding

Request forwarding allows for the analytics related requests to be hosted on the same domain as the web service. This makes it harder to block by adblockers.
Request forwarding is always enabled.

| key           | default | type     | description                                               |
| ------------- | ------- | -------- | --------------------------------------------------------- |
| `forwardPath` | `umami` | `string` | Forwards requests with this url prefix to the `umamiHost` |

Requests with a matching url are forwarded to the `umamiHost`. The path is preserved.

- `<forwardPath>/script.js` -> `<umamiHost>/script.js`
- `<forwardPath>/api/send` -> `<umamiHost>/api/send`

## Script Injection

If `scriptInjection` is enabled (by default) and the response `Content-Type` is `text/html`, the plugin will inject the umami script tag/source at the end of the response body.

The [`data-website-id`](https://umami.is/docs/tracker-configuration#data-domains) will be set to the `websiteId`.

| key                     | default | type       | description                                                                                          |
| ----------------------- | ------- | ---------- | ---------------------------------------------------------------------------------------------------- |
| `scriptInjection`       | `true`  | `bool`     | Injects the umami script tag into the response                                                       |
| `scriptInjectionMode`   | `tag`   | `string`   | `tag` or `source`. See below                                                                         |
| `autoTrack`             | `true`  | `bool`     | See original docs [data-auto-track](https://umami.is/docs/tracker-configuration#data-host-url)       |
| `doNotTrack`            | `false` | `bool`     | See original docs [data-do-not-track](https://umami.is/docs/tracker-configuration#data-do-not-track) |
| `cache`                 | `false` | `bool`     | See original docs [data-cache](https://umami.is/docs/tracker-configuration#data-cache)               |
| `domains`               | `[]`    | `[]string` | See original docs [data-domains](https://umami.is/docs/tracker-configuration#data-domains)           |
| `evadeGoogleTagManager` | `false` | `bool`     | See original docs [Google Tag Manager](https://umami.is/docs/tracker-configuration)                  |

There are two modes for script injection:
- `tag`: Injects the script tag with `src="/<forwardPath>/script.js"` into the response
- `source`: Downloads & injects the script source into the response

## Server Side Tracking

The plugin can be configured to send tracking events to the umami sevrer as requests come in. This removes the need for JavaScript on the client side.
It also allows to track pages that are not `text/html` or are not rendered by a browser.

However, it is not possible to track `title` or `display` values, as they are not available on the server side.

SST can be combined with script injection, but it is recommended to turn of `autoTrack` to avoid double tracking.

Tracked events have the name `traefik`.

The `domains` configuration is considered for SST as well. If domains is empty, all hosts are tracked, otherwise the host must be in the list. The port of the host is ignored.

| key                  | default | type   | description                  |
| -------------------- | ------- | ------ | ---------------------------- |
| `serverSideTracking` | `false` | `bool` | Enables server side tracking |