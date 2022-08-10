/*
 * Copyright (c) 2022, MegaEase
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package eval

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	hq "github.com/antchfx/htmlquery"
	jq "github.com/antchfx/jsonquery"
	xq "github.com/antchfx/xmlquery"
	"golang.org/x/net/html"
)

// Extractor is the interface for all extractors
type Extractor interface {
	SetQuery(string)
	SetVarType(VarType)
	SetTimeFormat(string)
	Extract() (interface{}, error)
}

// BaseExtractor is the base extractor
type BaseExtractor struct {
	Name         string  `yaml:"name"`        // variable name
	VarType      VarType `yaml:"type"`        // variable type
	TimeFormat   string  `yaml:"time_format"` // time format
	Document     string  `yaml:"-"`
	ExtractStrFn func() (string, error)
}

// SetVarType sets the variable type
func (x *BaseExtractor) SetVarType(t VarType) {
	x.VarType = t
}

// SetTimeFormat sets the time format
func (x *BaseExtractor) SetTimeFormat(format string) {
	x.TimeFormat = format
}

// Extract extracts the value from the document by xpath expression
func (x *BaseExtractor) Extract() (interface{}, error) {
	switch x.VarType {
	case String:
		return x.ExtractStrFn()
	case Int:
		return x.ExtractInt()
	case Float:
		return x.ExtractFloat()
	case Bool:
		return x.ExtractBool()
	case Time:
		return x.ExtractTime()
	case Duration:
		return x.ExtractDuration()
	}
	return nil, fmt.Errorf("unknown type: %s", x.VarType)
}

// ExtractInt extracts the value from the document by xpath expression
func (x *BaseExtractor) ExtractInt() (int, error) {
	s, err := x.ExtractStrFn()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

// ExtractFloat extracts the value from the document by xpath expression
func (x *BaseExtractor) ExtractFloat() (float64, error) {
	s, err := x.ExtractStrFn()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}

// ExtractBool extracts the value from the document by xpath expression
func (x *BaseExtractor) ExtractBool() (bool, error) {
	s, err := x.ExtractStrFn()
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(s)
}

// ExtractTime extracts the value from the document by xpath expression
func (x *BaseExtractor) ExtractTime() (time.Time, error) {
	s, err := x.ExtractStrFn()
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(x.TimeFormat, s)
}

// ExtractDuration extracts the value from the document by xpath expression
func (x *BaseExtractor) ExtractDuration() (time.Duration, error) {
	s, err := x.ExtractStrFn()
	if err != nil {
		return 0, err
	}
	return time.ParseDuration(s)
}

// -----------------------------------------------------------------------------

// XPathNode is the generic type for xpath node
type XPathNode interface {
	jq.Node | xq.Node | html.Node
}

// XPathExtractor is a struct for extracting values from a html/xml/json string
type XPathExtractor[T XPathNode] struct {
	BaseExtractor
	XPath  string `yaml:"xpath"` // xpath expression
	Parser func(string) (*T, error)
	Query  func(*T, string) (*T, error)
	Inner  func(*T) string
}

// SetQuery sets the xpath expression
func (x *XPathExtractor[T]) SetQuery(q string) {
	x.XPath = q
}

// Query query the string from the document by xpath expression
func Query[T XPathNode](document, xpath string,
	parser func(string) (*T, error),
	query func(*T, string) (*T, error),
	inner func(*T) string) (string, error) {
	doc, err := parser(document)
	if err != nil {
		return "", err
	}
	n, err := query(doc, xpath)
	if err != nil {
		return "", err
	}
	if n == nil {
		return "", nil
	}
	return inner(n), nil
}

// ExtractString extracts the value from the document by xpath expression
func (x *XPathExtractor[T]) ExtractStr() (string, error) {
	return Query(x.Document, x.XPath, x.Parser, x.Query, x.Inner)
}

// NewJSONExtractor creates a new JSONExtractor
func NewJSONExtractor(document string) *XPathExtractor[jq.Node] {
	x := &XPathExtractor[jq.Node]{
		BaseExtractor: BaseExtractor{
			VarType:  String,
			Document: document,
		},
		Parser: func(document string) (*jq.Node, error) {
			return jq.Parse(strings.NewReader(document))
		},
		Query: func(doc *jq.Node, xpath string) (*jq.Node, error) {
			return jq.Query(doc, xpath)
		},
		Inner: func(n *jq.Node) string {
			return n.InnerText()
		},
	}
	x.ExtractStrFn = x.ExtractStr
	return x
}

// NewXMLExtractor creates a new XMLExtractor
func NewXMLExtractor(document string) *XPathExtractor[xq.Node] {
	x := &XPathExtractor[xq.Node]{
		BaseExtractor: BaseExtractor{
			VarType:  String,
			Document: document,
		},
		Parser: func(document string) (*xq.Node, error) {
			return xq.Parse(strings.NewReader(document))
		},
		Query: func(doc *xq.Node, xpath string) (*xq.Node, error) {
			return xq.Query(doc, xpath)
		},
		Inner: func(n *xq.Node) string {
			return n.InnerText()
		},
	}
	x.ExtractStrFn = x.ExtractStr
	return x
}

// NewHTMLExtractor creates a new HTMLExtractor
func NewHTMLExtractor(document string) *XPathExtractor[html.Node] {
	x := &XPathExtractor[html.Node]{
		BaseExtractor: BaseExtractor{
			VarType:  String,
			Document: document,
		},
		Parser: func(document string) (*html.Node, error) {
			return html.Parse(strings.NewReader(document))
		},
		Query: func(doc *html.Node, xpath string) (*html.Node, error) {
			return hq.Query(doc, xpath)
		},
		Inner: func(n *html.Node) string {
			return hq.InnerText(n)
		},
	}
	x.Document = document
	x.ExtractStrFn = x.ExtractStr
	return x
}

//------------------------------------------------------------------------------

// RegexExtractor is a struct for extracting values from a plain string
type RegexExtractor struct {
	BaseExtractor
	Regex string `yaml:"regex"` // regex expression
}

// SetQuery sets the regex expression
func (r *RegexExtractor) SetQuery(q string) {
	r.Regex = q
}

// MatchStr matches the string with the regex expression
func (r *RegexExtractor) MatchStr() (string, error) {
	re := regexp.MustCompile(r.Regex)
	match := re.FindStringSubmatch(r.Document)
	if match == nil {
		return "", fmt.Errorf("no match found for - %s", r.Regex)
	}
	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(match) {
			if len(name) > 0 {
				r.Name = name
			}
			return match[i], nil
		}
	}
	return match[0], nil
}

// NewRegexExtractor creates a new RegexExtractor
func NewRegexExtractor(document string) *RegexExtractor {
	x := &RegexExtractor{
		BaseExtractor: BaseExtractor{
			VarType:  String,
			Document: document,
		},
	}
	x.ExtractStrFn = x.MatchStr
	return x
}
