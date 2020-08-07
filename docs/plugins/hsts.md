_Authors: grenfeldt@google.com_

# HSTS Plugin

HSTS<sup>1</sup> (HTTP Strict Transport Security) informs browsers that a website should only be accessed using HTTPS and not HTTP. This plugin enforces HSTS by redirecting all HTTP traffic to HTTPS and by setting the `Strict-Transport-Security` header on all outgoing HTTPS traffic.

1) HSTS: [RFC](https://tools.ietf.org/html/rfc6797), [MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security), [Wikipedia](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security)

## Usage

To construct the plugin with safe default settings, use: `hsts.NewPlugin()`.

## Options

**Option (Default value)**: Description.

- **MaxAge (`2 years`)**: The amount of time that the browser should remember to only access this site using HTTPS. The default value is 2 years since this is part of the deployment recommendations of https://hstspreload.org/.
- **IncludeSubDomains (`enabled`)**: When this is enabled, the browser knows that all subdomains should also only be accessed via HTTPS.
- **Preload (`disabled`)**: Must be enabled for this domain to be eligible to be added to the browser HSTS preload list. For more info, see: https://hstspreload.org/. This is disabled by default so that your domain can't be added to the preload list unintentionally.
- **BehindProxy (`disabled`)**: If this server is placed behind a proxy that terminates HTTPS then only HTTP traffic will be received. This option disables redirection of HTTP to HTTPS and always sends the `Strict-Transport-Security` header, even over HTTP.
