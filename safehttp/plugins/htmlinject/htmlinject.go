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

// Package htmlinject provides utilities to pre-process HTML templates and inject additional parts into them before parsing.
package htmlinject

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/google/safehtml/template"
	"github.com/google/safehtml/template/uncheckedconversions"
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
	// AddAttributes is a list of strings to add to the HTML as attributes.
	// All the given strings will be appended verbatim after the matched tag so they
	// should be prefixed with a space.
	AddAttributes []string
	// AddNodes is a list of nodes to append immediately after the opening tag that matched.
	// This means that for elements that have a matching closing tag the added node will be
	// a child node, for self-closing tags it will be a sibling.
	AddNodes []string
}

func (r Rule) String() string { return r.Name }

// TransformConfig is a slice of Rules that are somehow related to each other.
type TransformConfig []Rule

// CSPNoncesDefaultFuncName is the default func name for the func that generates CSP nonces.
const CSPNoncesDefaultFuncName = "CSPNonce"

// CSPNoncesDefault is the default config for CSP Nonces. The rewritten template
// expects the CSPNonce Func to be available in the template to provide nonces.
var CSPNoncesDefault = CSPNonces(`nonce="{{` + CSPNoncesDefaultFuncName + `}}"`)

// CSPNonces constructs a Config to add CSP nonces to a template. The given nonce
// attribute will be automatically prefixed with the required empty space.
func CSPNonces(nonceAttr string) TransformConfig {
	nonceAttr = " " + nonceAttr
	return TransformConfig{
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
		Rule{
			Name:          "Nonces for styles",
			OnTag:         "style",
			AddAttributes: []string{nonceAttr},
		},
	}
}

// XSRFTokensDefaultFuncName is the default func name for the func that generates XSRF tokens.
const XSRFTokensDefaultFuncName = `XSRFToken`

// XSRFTokensDefault is the default config to add hidden inputs to forms to provide
// an anti-XSRF token. The rewritten template expects the XSRFToken Func to be available
// in the template to provide tokens and sets the name for the sent value to be "xsrf-token".
var XSRFTokensDefault = XSRFTokens(`<input type="hidden" name="xsrf-token" value="{{` + XSRFTokensDefaultFuncName + `}}">`)

// XSRFTokens constructs a Config to add the given string as a child node to forms.
func XSRFTokens(inputTag string) TransformConfig {
	return TransformConfig{Rule{
		Name:     "XSRFTokens on forms",
		OnTag:    "form",
		AddNodes: []string{inputTag}}}
}

// Transform rewrites the given template according to the given configs.
// If the passed io.Rewriter has a `Size() int64` method it will be used to pre-allocate buffers.
func Transform(src io.Reader, cfg ...TransformConfig) (string, error) {
	rw := rewriter{
		rules:     map[string][]Rule{},
		tokenizer: html.NewTokenizer(src),
		out:       &strings.Builder{},
	}
	if sizer, ok := src.(interface{ Size() int64 }); ok {
		rw.out.Grow(int(sizer.Size()))
	}
	for _, c := range cfg {
		for _, r := range c {
			rw.rules[r.OnTag] = append(rw.rules[r.OnTag], r)
		}
	}
	if err := rw.rewrite(); err != nil {
		return "", err
	}
	return rw.out.String(), nil
}

// LoadConfig is a configuration to use with loaders when processing a template.
type LoadConfig struct {
	// DisableCSP disables CSP autononcing
	DisableCSP bool
	// DisableXSRF disables XSRF token injection
	DisableXSRF bool
}

// LoadTrustedTemplate processes the given TrustedTemplate with the specified default configurations and
// adds it to the given template.
// If the given template is nil a new one is created.
func LoadTrustedTemplate(tpl *template.Template, lcfg LoadConfig, src template.TrustedTemplate) (*template.Template, error) {
	var cfg []TransformConfig
	noop := func() string {
		panic("this function should never be called, templates should be cloned and injected with the noncing functions, not executed directly")
	}
	funcMap := map[string]interface{}{}
	if !lcfg.DisableCSP {
		cfg = append(cfg, CSPNoncesDefault)
		funcMap[CSPNoncesDefaultFuncName] = noop
	}
	if !lcfg.DisableXSRF {
		cfg = append(cfg, XSRFTokensDefault)
		funcMap[XSRFTokensDefaultFuncName] = noop
	}
	got, err := Transform(strings.NewReader(src.String()), cfg...)
	if err != nil {
		return nil, err
	}
	// We took a TrustedTemplate and transformed it with rules that we trust, so we know the output is still trusted.
	tt := uncheckedconversions.TrustedTemplateFromStringKnownToSatisfyTypeContract(got)
	if tpl == nil {
		tpl = template.New("htmlinjected")
	}
	return tpl.Funcs(funcMap).ParseFromTrustedTemplate(tt)
}

// LoadFiles matches the behavior of safehtml.ParseFiles but runs a transformation on every loaded template.
func LoadFiles(tpl *template.Template, lcfg LoadConfig, filenames ...template.TrustedSource) (*template.Template, error) {
	// The naming juggling below is quite odd but is kept for consistency.
	if len(filenames) == 0 {
		return nil, fmt.Errorf("no files named in call to LoadFiles")
	}
	for _, fnts := range filenames {
		fn := fnts.String()
		b, err := ioutil.ReadFile(fn)
		if err != nil {
			return nil, err
		}
		name := filepath.Base(fn)
		var t *template.Template
		if tpl == nil {
			tpl = template.New(name)
		}
		if name == tpl.Name() {
			t = tpl
		} else {
			t = tpl.New(name)
		}
		// We are loading a file from a TrustedSource, so this conversion is safe.
		tts := uncheckedconversions.TrustedTemplateFromStringKnownToSatisfyTypeContract(string(b))
		_, err = LoadTrustedTemplate(t, lcfg, tts)
		if err != nil {
			return nil, err
		}
	}
	return tpl, nil
}

// LoadGlob matches the behavior of safehtml.ParseGlob but runs a transformation on every loaded template.
func LoadGlob(tpl *template.Template, lcfg LoadConfig, pattern template.TrustedSource) (*template.Template, error) {
	filenames, err := filepath.Glob(pattern.String())
	if err != nil {
		return nil, err
	}
	if len(filenames) == 0 {
		return nil, fmt.Errorf("pattern matches no files: %#q", pattern.String())
	}
	var tts []template.TrustedSource
	for _, fn := range filenames {
		// The pattern expanded from a trusted source, so the expansion is still trusted.
		tts = append(tts, uncheckedconversions.TrustedSourceFromStringKnownToSatisfyTypeContract(fn))
	}
	return LoadFiles(tpl, lcfg, tts...)
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
	// Copy raw tokens to better formats
	var (
		tagname    string
		attributes = map[string]string{}
		raw        = make([]byte, len(r.tokenizer.Raw()))
	)
	{
		copy(raw, r.tokenizer.Raw())
		tag, hasAttr := r.tokenizer.TagName()
		tagname = string(tag)
		for hasAttr {
			key, val, more := r.tokenizer.TagAttr()
			hasAttr = more
			attributes[string(key)] = string(val)
		}
	}

	// Filter rules by attributes
	var triggeredRules []Rule
	{
		for _, r := range r.rules[tagname] {
			match := true
			for k, v := range r.WithAttributes {
				if attributes[k] != v {
					match = false
					break
				}
			}
			if match {
				triggeredRules = append(triggeredRules, r)
			}
		}
	}

	// Emit the rewritten HTML
	{
		attrPos := len(tagname) + 1
		// Write the "<" symbol and the tag name, e.g. "<script"
		if _, err := r.out.Write(raw[:attrPos]); err != nil {
			return fmt.Errorf("copying beginning of tag: %w", err)
		}
		// Write the attributes we have to add
		for _, rule := range triggeredRules {
			for _, attr := range rule.AddAttributes {
				if _, err := r.out.WriteString(attr); err != nil {
					return fmt.Errorf("executing rule %q: %w", rule.Name, err)
				}
			}
		}
		// Write the rest of the opening tag, e.g. ` src="foo.js">`
		if _, err := r.out.Write(raw[attrPos:]); err != nil {
			return fmt.Errorf("copying end of tag: %w", err)
		}
		// Write the nodes we have to add
		for _, rule := range triggeredRules {
			for _, node := range rule.AddNodes {
				if _, err := r.out.WriteString(node); err != nil {
					return fmt.Errorf("executing rule %q: %w", rule.Name, err)
				}
			}
		}
	}
	return nil
}
