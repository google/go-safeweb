// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package server TODO(clap|kele) describe this as what the users would normally have to write.
package server

import (
	"github.com/google/go-safeweb/safehttp"

	"embed"

	"github.com/google/go-safeweb/safehttp/plugins/htmlinject"
	"github.com/google/safehtml/template"

	"github.com/google/go-safeweb/examples/sample-application/secure/auth"
	"github.com/google/go-safeweb/examples/sample-application/secure/responses"
	"github.com/google/go-safeweb/examples/sample-application/storage"
)

//go:embed static
var staticFiles embed.FS

//go:embed templates
var templatesFS embed.FS

var templates *template.Template

func init() {
	tplSrc := template.TrustedSourceFromConstant("templates/*.tpl.html")
	var err error
	// Automatically inject CSP nonces and XSRF tokens placeholders.
	templates, err = htmlinject.LoadGlobEmbed(nil, htmlinject.LoadConfig{}, tplSrc, templatesFS)
	if err != nil {
		panic(err)
	}
}

type serverDeps struct {
	db *storage.DB
}

func Load(db *storage.DB, cfg *safehttp.ServeMuxConfig) {
	deps := &serverDeps{
		db: db,
	}

	// Private endpoints, only accessible to authenticated users (default).
	cfg.Handle("/notes/", "GET", getNotesHandler(deps))
	cfg.Handle("/notes", "POST", postNotesHandler(deps))
	cfg.Handle("/logout", "POST", logoutHandler(deps))

	// Public enpoints, no auth checks performed.
	cfg.Handle("/login", "POST", postLoginHandler(deps), auth.Skip{})
	cfg.Handle("/static/", "GET", safehttp.FileServerEmbed(staticFiles), auth.Skip{})
	cfg.Handle("/", "GET", indexHandler(deps), auth.Skip{})
}

func getNotesHandler(deps *serverDeps) safehttp.Handler {
	return safehttp.HandlerFunc(func(rw safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		user := auth.User(r)
		notes := deps.db.GetNotes(user)
		return safehttp.ExecuteNamedTemplate(rw, templates, "notes.tpl.html", map[string]interface{}{
			"notes": notes,
			"user":  user,
		})
	})
}

func postNotesHandler(deps *serverDeps) safehttp.Handler {
	noFormErr := responses.NewError(
		safehttp.StatusBadRequest,
		template.MustParseAndExecuteToHTML(`Please submit a valid form with "title" and "text" parameters.`),
	)
	noFieldsErr := responses.NewError(
		safehttp.StatusBadRequest,
		template.MustParseAndExecuteToHTML("Both title and text must be specified."),
	)

	return safehttp.HandlerFunc(func(rw safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		form, err := r.PostForm()
		if err != nil {
			return rw.WriteError(noFormErr)
		}
		title := form.String("title", "")
		body := form.String("text", "")
		if title == "" || body == "" {
			return rw.WriteError(noFieldsErr)
		}
		user := auth.User(r)
		deps.db.AddOrEditNote(user, storage.Note{Title: title, Text: body})

		notes := deps.db.GetNotes(user)
		return safehttp.ExecuteNamedTemplate(rw, templates, "notes.tpl.html", map[string]interface{}{
			"notes": notes,
			"user":  user,
		})
	})
}

func indexHandler(deps *serverDeps) safehttp.Handler {
	return safehttp.HandlerFunc(func(rw safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		user := auth.User(r)
		if user != "" {
			return safehttp.Redirect(rw, r, "/notes/", safehttp.StatusTemporaryRedirect)
		}
		return safehttp.ExecuteNamedTemplate(rw, templates, "index.tpl.html", nil)
	})
}

// Logout and Login handlers would normally be centralized and provided by a separate package owned by the security team.
// Since this is a simple example application they are here together with the rest.
func logoutHandler(deps *serverDeps) safehttp.Handler {
	return safehttp.HandlerFunc(func(rw safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		auth.ClearSession(r)
		return safehttp.Redirect(rw, r, "/", safehttp.StatusSeeOther)
	})
}

var invalidAuthErr responses.Error = responses.NewError(
	safehttp.StatusBadRequest,
	template.MustParseAndExecuteToHTML("Please specify a username and a password, both must be non-empty and your password must match the one you use to register."),
)

func postLoginHandler(deps *serverDeps) safehttp.Handler {
	// Always return the same error to not leak the existence of a user.
	return safehttp.HandlerFunc(func(rw safehttp.ResponseWriter, r *safehttp.IncomingRequest) safehttp.Result {
		form, err := r.PostForm()
		if err != nil {
			return rw.WriteError(invalidAuthErr)
		}
		username := form.String("username", "")
		password := form.String("password", "")
		if username == "" || password == "" {
			return rw.WriteError(invalidAuthErr)
		}
		if err := deps.db.AddOrAuthUser(username, password); err != nil {
			return rw.WriteError(invalidAuthErr)
		}
		auth.CreateSession(r, username)
		return safehttp.Redirect(rw, r, "/notes/", safehttp.StatusSeeOther)
	})
}
