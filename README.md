# go-safeweb

**DISCLAIMER**: This is not an officially supported Google product.

`go-safeweb` is a collection of libraries for writing secure-by-default HTTP
servers in Go.

## Contributing

This project is in an early stage. We are currently **not accepting** any
contributions.

## Overview

The flexibility of Go’s [`net/http`](https://pkg.go.dev/net/http/) package
allows users to quickly implement HTTP servers.

Responses are then written simply as slices of bytes, headers can be arbitrarily
manipulated and so on. This approach offers much needed flexibility for these
who really need it.

Unfortunately, this approach leaves great space for introducing security
vulnerabilities and even experienced developers tend to do so.

This document aims to design an HTTP API that eliminates whole classes of bugs,
like Cross-Site Scripting (XSS) or Cross-Site Request Forgery (XSRF). This can
be achieved by an approach known at Google as _safe coding_. Learn more at
[Securing the Tangled Web (Chistoph Kern, 2014)](http://static.googleusercontent.com/media/research.google.com/en//pubs/archive/42934.pdf)
or
[Preventing Security Bugs through Software Design (Christoph Kern, 2016)](https://www.youtube.com/watch?v=ccfEu-Jj0as).

## Goals and Non-Goals

### Goals

#### G1: Secure-by-default

Security mechanisms are applied by default (opt-out, not opt-in).

#### G2: Unsafe Usage is Easy to Review, Track and Restrict

All opt-outs from security mechanisms are explicit. Wherever possible, they’re
contained inside a package or an option that’s easy to restrict.

#### G3: Designed for Evolving Security Requirements

Enforcing new security measures is feasible through AST manipulation. Existing
users can be migrated using static analysis and/or runtime monitoring. Read more
[here](#evolving-security-requirements).

#### G4: High Compatibility with Go’s Standard Library and Existing Open-Source Frameworks

Whenever possible, keep existing layouts, function signatures and other API
parts the same as the Go’s standard library. High compatibility enables wide
adoption.

### Non Goals

#### NG1: Safe API [Completeness](<https://en.wikipedia.org/wiki/Completeness_(logic)>)

Creating safe APIs for all the corner cases might result in a bloated codebase.
Our experience shows that this isn’t necessary.

#### NG2: Full Compatibility with Go’s Standard Library and Existing Open-Source Frameworks

Existing open-source frameworks or the Go standard library need to support each
developer scenario. This would have left us with limited options of creating
safe-by-default HTTP servers.

## Security Vulnerabilities and Mitigations

On a high level, we plan to address, or provide the needed infrastructure to
address, following issues (not an exhaustive list):

- **XSS (cross-site scripting) and XSSI (cross-site script inclusion)** - e.g.
  by controlling how responses are generated
- **XSRF (cross-site request forgery)** - e.g. by using Fetch Metadata policies,
  supporting token-based XSRF protection
- **CORS (cross-origin resource sharing)** - e.g. by taking control of CORS
  response headers and handling CORS preflight requests
- [**CSP (content security policy)**](https://csp.withgoogle.com/docs/index.html) -
  e.g. by automatically adding script nonces to HTML responses, adding relevant
  security headers
- **Transport Security** - e.g. by [enforcing HSTS support](plugins/hsts.md)
- **IFraming** - e.g. by setting relevant HTTP headers to restrict framing or
  providing server-side support for origin selection
- **Auth (access control)** - e.g. by providing infrastructure for plugging in
  access control logic in an uniform, auditable way
- **HTTP Request Parsing Bugs** - e.g. by implementing strict and well
  documented parsing behavior
- **Error responses** - e.g. by providing infrastructure for uniform error
  handling (e.g. to prevent accidental leaks or XSS from error responses)
- **Enforcement of other security specific HTTP headers**

## Appendix

### Evolving Security Requirements (example)

Imagine an API for configuring access control. It features three types of rules:

- `ALLOW(user)` - allows a given `user`
- `DENY(user)` - denies a given `user` (has priority over `ALLOW`)
- `REPORT(user)` - reports that it has seen a request from a given `user`

Imagine now that at some point, security standards need to be increased and
`user = "frombulator"` has been determined to not meet the desired bar.

How do we, for all the services running in our company, address this?

1.  For existing services, we add a `LegacyFrombulatorAccess` option like so:
    `security.AccessControl(rules, unsafe.LegacyFrombulatorAccess())`.
1.  We change the `security.AccessControl()` call to add by default a
    `DENY("frombulator")` rule. This rule **is not added** if
    `unsafe.LegacyFrombulatorAccess` is applied.
1.  Instead, `unsafe.LegacyFrombulatorAccess` adds a `REPORT("frombulator")`
    rule.

This way, we have:

- Ensured that all new callers of `security.AccessControl` use the safe setting
  by default.
- Can monitor existing services dependence on calls from the `frombulator`.
  After a period of observation (let’s say, 30 days):
  - If the service doesn’t receive requests from the `frombulator`: **prune the
    `unsafe.LegacyFrombulatorAccess`** option.
  - If the service does receive requests from the `frombulator`: **inform the
    service owners and plan a fix.**

Crucially, only the last case (dependence on unsafe configuration) requires
engineering work per service. The rest can be automated.

**This approach is possible due to careful API design.** A missing `DENY` or
`REPORT` rule, or a single sink in the form of `security.AccessControl` would
make this infeasible.

### Source Code Headers

Every file containing source code must include copyright and license
information. This includes any JS/CSS files that you might be serving out to
browsers. (This is to help well-intentioned people avoid accidental copying that
doesn't comply with the license.)

Apache header:

    Copyright 2020 Google LLC

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        https://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
