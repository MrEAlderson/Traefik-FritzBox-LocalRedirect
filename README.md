# FritzBox Local Redirect Traefik plugin
<img align="right" src="./icon.png" height="130" alt="beacon-pip-frame-proxy">
<p>
  ⚡ Reduce unnecessary overhead caused by proxies
  
  This plugin aims to automatically redirect the client to the local hostname of the server, thereby avoiding any proxy (e.g., CloudFlare tunnel) and the restrictions and performance penalties that would result from it.
</p>

## How does it work?
Do note that this plugin only works if you are using a FRITZ!Box router as it provides the public API that is necessary for this plugin to work.
With every request going through the Traefik proxy, the client's IP is matched with the public one. In case of a match, the request is being replaced with a code 307 (temporary redirect).
With the redirect, the host of the client's current URL (e.g. https://google.com/pics/123) is being replaced with another (e.g. http://myserver.local/pics/123).
It does support both IPv4 and IPv6.

### Why?
While the CloudFlare Tunnel is nice for its purpose, it's an unnecessary overhead if your server and your device are on the same home network.

### What service would I need it for?
Generally, any that may cause noticeable performance penalties due to increased traffic, such as Jellyfin, Plex, or any other service that handles files of greater size.

## Configuration
Create a middleware for each service, such as:
```yaml
    jellyfin-local-redirect:
      plugin:
        FritzBox_LocalRedirect:
          LocalHost: "https://my-server.local:8080" # Required: Hostname+Schema redirected to in case access from local network
          # Optional vv, they are listed with their default value
          FritzURL: "http://192.168.178.1:49000" # Address of the router's API
          RefreshTime: "30s" # Max age of cached public IP until it's refreshed (non-blocking)
          TimeoutTime: "5s" # Max public IP refresh duration until it's cancelled
```
and assign it to your Traefik router in question, such as:
```yaml
  routers:
    portainer:
      rule: "Host(`jellyfin.my-public-server.com`)"
      service: jellyfin
      entryPoints: [ http, https ]
      middlewares: [ jellyfin-local-redirect ]
```
