package safehtml_test

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
	"github.com/google/safehtml/template"
)

type dispatcher struct {
	wr io.Writer
}

func (d *dispatcher) Write(resp safehttp.Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		_, err := d.wr.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
	return nil
}

func (d *dispatcher) ExecuteTemplate(t safehttp.Template, data interface{}) error {
	switch x := t.(type) {
	case *template.Template:
		return x.Execute(d.wr, data)
	default:
		panic("not a safe response type")
	}
	return nil
}

func TestHandleRequestWrite(t *testing.T) {
	b := &strings.Builder{}
	d := &dispatcher{wr: b}

	m := safehttp.NewMachinery(func(w *safehttp.ResponseWriter, _ safehttp.IncomingRequest) safehttp.Result {
		return w.Write(safehtml.HTMLEscaped("<h1>Escaped, so not really a heading</h1>"))
	}, d)

	m.HandleRequest("GET / HTTP/1.1\r\n" +
		"\r\n")

	want := "&lt;h1&gt;Escaped, so not really a heading&lt;/h1&gt;"
	got := b.String()
	if got != want {
		t.Errorf("response got = %q, want %q", got, want)
	}
}

func TestHandleRequestWriteTemplate(t *testing.T) {
	b := &strings.Builder{}
	d := &dispatcher{wr: b}

	m := safehttp.NewMachinery(func(w *safehttp.ResponseWriter, _ safehttp.IncomingRequest) safehttp.Result {
		return w.WriteTemplate(template.Must(template.New("name").Parse("<h1>{{ . }}</h1>")), "This is an actual heading, though.")
	}, d)

	m.HandleRequest("GET / HTTP/1.1\r\n" +
		"\r\n")

	want := "<h1>This is an actual heading, though.</h1>"
	got := b.String()
	if got != want {
		t.Errorf("response got = %q, want %q", got, want)
	}
}
