## Design

GoDoc: https://pkg.go.dev/github.com/google/go-safeweb/safehttp.


## Design Principles

Go Safe Web is designed to achieve the goals listed at https://github.com/google/go-safeweb#goals-and-non-goals.


## Users

We distinguish two categories of users:

1. Security reviewers.
2. General software engineers.

Go Safe Web provides a small subset of the public APIs meant to be reviewed
by **security reviewers**. The Go Safe Web packages are designed taking into
account the capabilities of modern linterns, allowing them to detect the
usage of these APIs. For instance, we expect that the `ServeMuxConfig` type
will never be used by **general software engineers** without insight from
**security reviewers.**

**Security reviewers** are expected to own and maintain the `Dispatcher`
implementation, and the `ServeMuxConfig` usage (crucially, they need to
install `Interceptor`s for protection).

**General software engineers** are expected to use the `ServeMux` provided by
the **security reviewers**, possibly adding `InterceptorConfig`s when
registering handlers to configure their behavior (e.g. enable CORS). They own
the implementation of the `Handler`s.


## Core Concepts

### Safe Responses and the `Dispatcher`

`Handler` implementations return HTTP responses using the `ResponseWriter`.
In contrast to the `net/http.ResponseWriter`, it accepts only `Response`
objects, instead of byte slices.

These `Response` objects are then passed to the `Dispatcher`. The
`Dispatcher` is responsible for detecting whether an implementation of the
`Response` interface is considered a **safe response**, or not. **Safe
responses are objects that are considered safe to send back as a HTTP
response.** Examples include:

*   HTML constructed using safe libraries (e.g. `github.com/google/safehtml`)
*   JSON responses protected from XSSI using the parser breaking prefix
    `")]}',\n"`

The Go Safe Web framework provides a `DefaultDispatcher` that can be used to
generate safe HTML, JSON and error responses.

**Security reviewers are responsible for the implementation of the
`Dispatcher` used in their project.**


### Handlers

Go Safe Web `Handler`s mimic the ones defined in `net/http`, with the main
differences being:

*   Only **safe response** types can be written to the `ResponseWriter` passed to
    the `Handler`. These bundle the HTTP response status code with the response
    itself.
*   HTTP response headers are set using the `ResponseWriter`s safe header
    manipulation APIs. Some of the headers might be owned by the `Interceptor`s,
    therefore not accessible to the `Handler`.
*   A `Handler` writes a `Response` exactly once and returns the `Result`,
    `IncomingRequest` replaces `http.Request` and provides safe methods to access
    headers and forms.


### Interceptors

`Interceptors` run:

*   **Before** the execution of the `Handler`, and
*   after the `Handler` **commits** to writing a `Response`.

These are implemented as, respectively, `Interceptor.Before()` and
`Interceptor.Commit()` methods.
Examples of interceptors include:

*   checks on `POST` requests whether an anti-XSRF token has been submitted;
    rejects requests if this is not the case;
*   setting an appropriate Content Security Policy in responses;
*   logging the time needed to process the request.

The `Before` phase of an `Interceptor` is capable of writing a `Response` to
the `ResponseWriter`, therefore controlling the execution of further
interceptors and/or the `Handler`. This capability is needed for e.g.
centralized access control or anti-XSRF protection.

The `Commit` phase of an `Interceptor` can change the `Response` (e.g. inject
anti-XSRF tokens in HTML forms) and HTTP response headers (e.g. Content
Security Policy). `Commit` calls are run in a reversed order compared to
`Before` calls (i.e. the interceptor whose `Before` was called last gets
their `Commit` called first).

`Interceptor`s can be configured during `Handler` registration on the
`ServeMuxConfig` using `InterceptorConfig` objects (e.g. to allow iframing of
the given handler, overriding the default `DENY` configuration).


### Plugins

Plugins are packages that help build a secure web server. They provide
`Interceptor` implementations or ways to communicate with them (e.g. because
the `Interceptor` **claims a header** in the **before interceptor phase**,
but still wants to allow for safe alterations by the `Handler`).


### Response Headers Claiming

HTTP response headers can be security-sensitive. Examples include the Content
Security Policy or Content Type headers.

To prevent vulnerabilities introduced by `Handler` implementations, the
`ResponseWriter` provides a way to **claim a response header.** A call to
`ResponseWriter.Header().Claim("foo")` **claims** the `foo` HTTP response
header and returns a function that allows setting it. `Interceptor`s can use
this feature to **claim** the response headers they need to provide security
guarantees. Any attempts to alter a claimed response header other than by the
setter returned by `Claim()` result in a `panic`. If an `Interceptor` wants
to allow some ways of safely altering the header by `Handler`
implementations, it should **claim** the relevant header in the **before
**phase and provide methods to alter the claimed header inside the `Handler`.
The setter returned by `Claim()` can e.g. be stored and retrieved inside the
`IncomingRequest.Context()`.
