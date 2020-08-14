# Plugins

The framework supports plugins. They provide a flexible way to address security 
issues by setting non-overridable security headers or interrupting and
responding to incoming requests before they reach the handler.

## Provided plugins

- [HSTS](plugins/hsts.md): Automatically redirects HTTP traffic to HTTPS and
sets the `Strict-Transport-Security` header.
- [staticheaders](plugins/staticheaders.md): Sets the `X-Content-Type-Options` header and the `X-XSS-Protection` header.
