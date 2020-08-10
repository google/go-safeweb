package htmlinject

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Rule is a directive to instruct Transform on how to rewrite the given template.
type Rule struct {
	// Name is used for debug purposes in case rewriting fails.
	Name string
	// OnTag is the tag to be used to trigger the rule.
	OnTag string
	// WithAttributes is a filter applied on tags to decide whether to run the Rule:
	// only tags with the given attributes key:value will be matched.
	WithAttributes map[string]string
	// AddAttributes is a list of strings to add to the HTML in place of attributes.
	// All the given strings will be appended verbatim after the matched tag so they
	// should be prefixed with a space.
	AddAttributes []string
	// AddNodes is a list of nodes to append immediately after the openin tag that matched.
	// This means that for elements that have a matching closing tag the added node will be
	// a child node, for self-closing tags it will be a sibling.
	AddNodes []string
}

func (r Rule) String() string { return r.Name }

// Config is a slice of Rules that are somehow related to each other.
type Config []Rule

// CSPNoncesDefault is the default config for CSP Nonces. The rewritten template
// expects the CSPNonce Func to be available in the template to provide nonces.
var CSPNoncesDefault = CSPNonces(`nonce="{{CSPNonce}}`)

// CSPNonces constructs a Config to add CSP nonces to a template. The given nonce
// attribute will be automatically prefixed with the required empty space.
func CSPNonces(nonceAttr string) Config {
	nonceAttr = " " + nonceAttr
	return Config{
		Rule{
			Name:          "Nonces for scripts",
			OnTag:         "script",
			AddAttributes: []string{nonceAttr},
		},
		Rule{
			Name:           "Nonces for link as=script rel=preload",
			OnTag:          "link",
			WithAttributes: map[string]string{"rel": "preload", "as": "script"},
			AddAttributes:  []string{nonceAttr},
		},
	}
}

// XSRFTokensDefault is the default config to add hidden inputs to forms to provide
// an anti-XSRF token. The rewritten template expects the XSRFToken Func to be available
// in the template to provide tokens and sets the name for the sent value to be "xsrftoken".
var XSRFTokensDefault = XSRFTokens(`<input type="hidden" name="xsrftoken" value="{{XSRFToken}}">`)

// XSRFTokens constructs a Config to add the given string as a child node to forms.
func XSRFTokens(inputTag string) Config {
	return Config{Rule{OnTag: "form", AddNodes: []string{inputTag}}}
}

// Transform rewrites the given template according to the given configs.
func Transform(src io.Reader, cfg ...Config) (tpl string, _ error) {
	rw := rewriter{
		rules:     map[string][]Rule{},
		tokenizer: html.NewTokenizer(src),
		out:       &strings.Builder{},
	}
	for _, c := range cfg {
		for _, r := range c {
			rw.rules[r.OnTag] = append(rw.rules[r.OnTag], r)
		}
	}
	if err := rw.rewrite(); err != nil {
		return "", fmt.Errorf("transforming template: %v", err)
	}
	return rw.out.String(), nil
}

type rewriter struct {
	// tag -> rules for that tag
	rules     map[string][]Rule
	tokenizer *html.Tokenizer
	out       *strings.Builder
}

// emitRaw copies the current raw token to the output.
func (r rewriter) emitRaw() error {
	_, err := r.out.Write(r.tokenizer.Raw())
	return err
}

// emit writes the given slice to the output.
func (r rewriter) emit(p []byte) error {
	_, err := r.out.Write(p)
	return err
}

// rewrite runs the rewriter.
func (r rewriter) rewrite() error {
	for {
		switch tkn := r.tokenizer.Next(); tkn {
		case html.ErrorToken:
			if err := r.tokenizer.Err(); !errors.Is(err, io.EOF) {
				return err
			}
			// We got EOF, let's just emit the last token and exit.
			return r.emitRaw()
		case html.StartTagToken, html.SelfClosingTagToken:
			if err := r.processTag(); err != nil {
				return err
			}
		default:
			if err := r.emitRaw(); err != nil {
				return err
			}
		}
	}
}

func (r rewriter) processTag() error {
	// TODO(empijei): implement
	return r.emitRaw()

}
