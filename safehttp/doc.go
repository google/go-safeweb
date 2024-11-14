// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package safehttp provides a framework for building secure-by-default web
// applications. See https://github.com/google/go-safeweb#readme to learn about
// the goals and features.
//
// # Safe Responses
//
// HTTP responses need to be crafted carefully in order to prevent common web
// vulnerabilities like Cross-site Scripting (XSS). To help with this, we use
// Safe Responses.
//
// Safe Responses are HTTP responses which have been determined to be safe when
// received by a modern, popular web browser. For example, we consider HTML
// responses generated using a template system enforcing contextual autoescaping
// to be safe, e.g. modern Angular or github.com/google/safehtml/template. Read
// more about contextual autoescaping here:
// https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/42934.pdf.
//
// The Go Safe Web ResponseWriter implementation only accepts Safe Responses
// instead of byte slices.
//
// Since different projects will consider different ways of crafting a response
// safe, we offer a way of configuring this in the framework. Whether a response
// is considered safe or not is determined by the Dispatcher.
//
// # Dispatcher
//
// An implementation of a Dispatcher should be provided by security experts in
// your project. The Dispatcher is called for every write method of the
// ResponseWriter and is used to determine whether the passed response should be
// considered safe. The Dispatcher is responsible for writing the response to
// the underlying http.ResponseWriter in a safe way.
//
// Go Safe Web provides a DefaultDispatcher implementation which supports
// github.com/google/safehtml responses.
//
// Warning: the security of the web application depends on a sound Dispatcher
// implementation. Make sure it is security-reviewed and keep it simple.
//
// # Interceptors
//
// Not all security features can be implemented using the Dispatcher alone. For
// instance, some requests should be rejected before they reach the handler in
// order to prevent from Cross-site Request Forgery (CSRF, XSRF). To support
// this, the framework uses Interceptors.
//
// An Interceptor implements methods that run before the request is passed to
// the handler, and after the handler has committed a response. These are,
// respectively, Before and Commit.
//
// # Life of a Request
//
// To tie it all together, we will explain how a single request goes through the
// framework.
//
// When the request reaches the server, the following happens:
//
// 1. The ServeMux routes the request (i.e. picks the appropriate Handler that
// will be eventually called). The HTTP method of the request is checked whether
// it matches any registered Handler.
//
// 2. [Before phase] The request is passed to all installed Interceptors, via
// their Before methods, in the order of installation on the ServeMux. Each of
// the Before methods can either let the execution continue by returning
// safehttp.NotWritten() (and not using any ResponseWriter write methods), or by
// actually writing a response. This would prevent further Before method calls
// of subsequent Interceptors.
//
// 3. The request is passed to the Handler. The Handler calls a write method of
// the ResponseWriter (e.g. Write, WriteError, Redirect...).
//
// 5. [Commit phase] Commit methods of the installed Interceptors are called, in
// LIFO order (i.e. first Interceptor to be called in Before phase is called
// last in the Commit phase). Commit methods can no longer influence the flow of
// the request, but can still set headers, cookies, response HTTP status code or
// even modify the response (if its type allows it).
//
// 6. [Dispatcher] After all Commit methods have been called, the framework
// passes the request to the Dispatcher. The Dispatcher determines the
// Content-Type of the response and whether the response should be considered a
// Safe Response. The Dispatcher is responsible for writing the response to the
// underlying http.ResponseWriter in a safe way.
//
// 7. The Handler returns the value returned by the ResponseWriter write method
// used. After the first write, any further writes are considered fatal errors.
// It is safe to use defer statements for cleanup tasks (e.g. closing a file
// that was used in a safehtml/template.Template response).
//
// Stack trace of the flow:
//
//	Mux.ServeHTTP()
//	--+ Mux routes the request and checks the method.
//	--+ InterceptorFoo.Before()
//	--+ InterceptorBar.Before()
//	--+ InterceptorBaz.Before()
//	--+ Handler()
//	----+ ResponseWriter.Write
//	------+ InterceptorBaz.Commit()  // notice the inverted order
//	------+ InterceptorBar.Commit()
//	------+ InterceptorFoo.Commit()
//	------+ Dispatcher.Write()
//	----+ The result of the Response.Write() call is returned.
//
// # Error Responses
//
// Error responses are written using ResponseWriter.WriteError. They go through
// the usual Commit and Dispatcher phases.
//
// # Configuring the Mux
//
// # TODO
//
// # Incremental Adoption
//
// In order to migrate your service using http.Handlers to the safehttp package,
// we recommend you start doing that one endpoint at a time. Use
// RegisteredHandler to do this.
//
//	safeMuxConfig := /* configured ServeMuxConfig, including interceptors */
//	safeMuxConfig.Handle("/bar", safehttp.MethodGET, barGETSafeHandler)
//	safeMuxConfig.Handle("/bar", safehttp.MethodPOST, barPOSTSafeHandler)
//	safeMuxConfig.Handle("/xyz", safehttp.MethodPOST, xyzSafeHandler)
//	safeMux := safeMuxConfig.Mux()
//
//	// old, not yet migrated
//	http.Handle("/foo", fooHandler)
//
//	// new, migrated
//	http.Handle("/bar", safehttp.RegisteredHandler(safeMux, "/bar"))
//	http.Handle("/xyz", safehttp.RegisteredHandler(safeMux, "/xyz"))
//
// # Restricting Risky APIs
//
// Some APIs are easy-to-misuse in a security sensitive context. We choose to
// restrict these and require a security review for their usage in order to
// prevent vulnerabilities.
//
// cmd/bancheck allows you to restrict APIs and check for their usage as part
// of the CI/CD pipeline and prevent potentially vulnerable code from being
// deployed. For detailed usage instructions, please see:
// https://pkg.go.dev/github.com/google/go-safeweb/cmd/bancheck
package safehttp
