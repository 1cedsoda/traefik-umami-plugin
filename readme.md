# traefik-umami-plugin

Use [Umami Analytics]() with the [Traefik Reverse Proxy]().

This plugin provides a middleware to provide umami anytics to a web servive.

Pros:
- No need to modify the web service
- Harder to block by adblockers (same domain)
- No need for JavaScript Loading (soon: script source injection)
- No need for JavaScript (soon: server side tracking)

# Roadmap
- [X] script tag injection
- [ ] script source injection
- [ ] request forwarding
- [ ] server side tracking

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
| key           | default | type   | description                                                                    |
| ------------- | ------- | ------ | ------------------------------------------------------------------------------ |
| `umamiHost`   | -       | string | Umami server host, reachable from within traefik (container). eg. `umami:3000` |
| `websiteId`   | -       | string | Website ID as configured in umami.                                             |
| `forwardPath` | umami   | string | Forwards requests with this url prefix to the `umamiHost`                      |

The middleware can then be used in a router. Remember to reference the correct provider.

## Script Tag Injection

If `scriptInjection` is enabled (by default) and the response `Content-Type` is `text/html`, the plugin will inject the umami script tag into the response head.

The script `src` will be set to `/<forwardPath>/script.js` and `data-website-id` will be set to the `websiteId`.

| key                     | default | type     | description                                                                                          |
| ----------------------- | ------- | -------- | ---------------------------------------------------------------------------------------------------- |
| `scriptInjection`       | true    | bool     | Injects the umami script tag into the response head                                                  |
| `autoTrack`             | true    | bool     | See original docs [data-auto-track](https://umami.is/docs/tracker-configuration#data-host-url)       |
| `doNotTrack`            | false   | bool     | See original docs [data-do-not-track](https://umami.is/docs/tracker-configuration#data-do-not-track) |
| `cache`                 | false   | bool     | See original docs [data-cache](https://umami.is/docs/tracker-configuration#data-cache)               |
| `domains`               | []      | []string | See original docs [data-domains](https://umami.is/docs/tracker-configuration#data-domains)           |
| `evadeGoogleTagManager` | false   | bool     | See original docs [Google Tag Manager](https://umami.is/docs/tracker-configuration)                  |