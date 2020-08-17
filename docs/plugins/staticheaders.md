_Authors: grenfeldt@google.com_

# Static headers Plugin

This plugin sets the `X-Content-Type-Options` header to `nosniff` and the
`X-XSS-Protection` header to `0` on all responses.

- `X-Content-Type-Options: nosniff` tells browsers to not try to sniff the
`Content-Type` of responses.
[MDN documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Content-Type-Options).
- `X-XSS-Protection: 0` tells the browser to disable any built in XSS filters.
These built in XSS filters are unnecessary when other, stronger, protections are
available and can introduce cross-site leaks vulnerabilities.
[MDN documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-XSS-Protection).
