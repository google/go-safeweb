// echo implements a simple echo server. Visit /echo?message=YourMessage.
package main

import (
	"log"
	"time"

	"github.com/google/safehtml"
	"github.com/google/safehtml/template"

	"github.com/google/go-safeweb/examples/echo/security/web"
	"github.com/google/go-safeweb/safehttp"
)

var start time.Time

func main() {
	mc := web.NewMuxConfig()

	mc.Handle("/echo", safehttp.MethodGet, safehttp.HandlerFunc(echo))
	mc.Handle("/uptime", safehttp.MethodGet, safehttp.HandlerFunc(uptime))

	start = time.Now()
	log.Println("Visit http://localhost:8080")
	log.Println("Listening on localhost:8080...")
	log.Fatal(web.ListenAndServeDev(8080, mc))
}

func echo(w safehttp.ResponseWriter, req *safehttp.IncomingRequest) safehttp.Result {
	q, err := req.URL.Query()
	if err != nil {
		return w.WriteError(safehttp.StatusBadRequest)
	}
	x := q.String("message", "")
	if len(x) == 0 {
		return w.WriteError(safehttp.StatusBadRequest)
	}
	return w.Write(safehtml.HTMLEscaped(x))
}

var uptimeTmpl *template.Template = template.Must(template.New("uptime").Parse(
	`<h1>Uptime: {{ .Uptime }}</h1>
{{- if .EasterEgg }}<h1>You've found an easter egg using "{{ .EasterEgg }}". Congrats!</h1>{{ end -}}`))

func uptime(w safehttp.ResponseWriter, req *safehttp.IncomingRequest) safehttp.Result {
	var x struct {
		Uptime    time.Duration
		EasterEgg string
	}
	x.Uptime = time.Now().Sub(start)

	// Easter egg handling.
	q, err := req.URL.Query()
	if err != nil {
		return w.WriteError(safehttp.StatusBadRequest)
	}
	if egg := q.String("easter_egg", ""); len(egg) != 0 {
		x.EasterEgg = egg
	}

	return safehttp.ExecuteTemplate(w, uptimeTmpl, x)
}
