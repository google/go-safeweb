package htmlinject

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTransform(t *testing.T) {
	var tests = []struct {
		name     string
		config   []Config
		in, want string
	}{
		{
			name: "nothing to change",
			in: `
<html>
<header><title>This is title</title></header>
<body>
Hello world
</body>
</html>
`,
			want: `
<html>
<header><title>This is title</title></header>
<body>
Hello world
</body>
</html>
`,
		},
		{
			name:   "add CSP nonces",
			config: []Config{CSPNoncesDefault},
			in: `
<html>
<head>
<link rel=preload as="script" src="gopher.js">
</head>
<body>
<script type="application/javascript">alert("script")</script>
</body>
</html>
`,
			want: `
<html>
<head>
<link nonce="{{.CSPNonce}}" rel=preload as="script" src="gopher.js">
</head>
<body>
<script nonce="{{CSPNonce)}}" type="application/javascript">alert("script")</script>
</body>
</html>
`,
		},
		{
			name:   "add XSRF protection",
			config: []Config{XSRFTokensDefault},
			in: `
<form>
  First name:<br>
  <input type="text" name="firstname"><br>
  Last name:<br>
  <input type="text" name="lastname">
</form>
`,
			want: `
<form><input type="hidden" name="xsrftoken" value="{{XSRFToken}}">
  First name:<br>
  <input type="text" name="firstname"><br>
  Last name:<br>
  <input type="text" name="lastname">
</form>
`,
		},
		{
			name:   "all configs",
			config: []Config{CSPNoncesDefault, XSRFTokensDefault},
			in: `
<html>
<head>
<link rel=preload as="script" src="gopher.js">
</head>
<body>
<script type="application/javascript">alert("script")</script>
<form>
  First name:<br>
  <input type="text" name="firstname"><br>
  Last name:<br>
  <input type="text" name="lastname">
</form>
</body>
</html>
`,
			want: `
<html>
<head>
<link nonce="{{.CSPNonce}}" rel=preload as="script" src="gopher.js">
</head>
<body>
<script nonce="{{.CSPNonce}}" type="application/javascript">alert("script")</script>
<form><input type="hidden" name="xsrftoken" value="{{.XSRFToken}}">
  First name:<br>
  <input type="text" name="firstname"><br>
  Last name:<br>
  <input type="text" name="lastname">
</form>
</body>
</html>
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Transform(strings.NewReader(tt.in), tt.config...)
			if err != nil {
				t.Fatalf("Transform: got err %q, didn't want one", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("-want +got %s", diff)
			}
		})
	}
}
