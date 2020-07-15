package header

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-safeweb/safehttp"
	"github.com/google/safehtml"
)

type dispatcher struct{}

func (dispatcher) Write(rw http.ResponseWriter, resp safehttp.Response) error {
	switch x := resp.(type) {
	case safehtml.HTML:
		_, err := rw.Write([]byte(x.String()))
		return err
	default:
		panic("not a safe response type")
	}
}

func (dispatcher) ExecuteTemplate(rw http.ResponseWriter, t safehttp.Template, data interface{}) error {
	return nil
}

func TestAccessIncomingHeaders(t *testing.T) {
	m := safehttp.NewMachinery(func(rw safehttp.ResponseWriter, ir *safehttp.IncomingRequest) safehttp.Result {
		if got, want := ir.Header.Get("A"), "B"; got != want {
			t.Errorf(`ir.Header.Get("A") got: %v want: %v`, got, want)
		}
		return rw.Write(safehtml.HTMLEscaped("hello"))
	}, &dispatcher{})

	request := "GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"A: B\r\n\r\n"
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(request)))
	if err != nil {
		t.Fatalf("http.ReadRequest() got err: %v", err)
	}
	recorder := httptest.NewRecorder()

	m.HandleRequest(recorder, req)

	resp := recorder.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(resp.Body) got err: %v, want nil", err)
	}

	if got, want := string(body), "hello"; got != want {
		t.Errorf("resp.Body: got = %q, want %q", got, want)
	}
}
