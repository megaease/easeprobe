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
	"errors"
	"io"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

func assertExtractor(t *testing.T, extractor Extractor, query string, vt VarType, expected interface{}, success bool) {
	extractor.SetQuery(query)
	extractor.SetVarType(vt)
	result, err := extractor.Extract()
	if success {
		assert.Nil(t, err)
	} else {
		assert.NotNil(t, err)
	}
	assert.Equal(t, expected, result)
}
func assertExtractorSucc(t *testing.T, extractor Extractor, query string, vt VarType, expected interface{}) {
	assertExtractor(t, extractor, query, vt, expected, true)
}
func assertExtractorFail(t *testing.T, extractor Extractor, query string, vt VarType, expected interface{}) {
	assertExtractor(t, extractor, query, vt, expected, false)
}

func TestHTMLExtractor(t *testing.T) {
	var htmlDoc = `
	<html>
	<head>
		<title>Hello World Example</title>
	</head>
	<body>
		<h1>Hello World Example</h1>
		<p>
		<a href="https://github.com/megaease/easeprobe">EaseProbe Github </a> is a simple, standalone, and lightWeight tool that can do health/status checking, written in Go.
		</p>
		<div id=number>
			<div id=one class=one>1</div>
			<div id=two class=two>2</div>
		</div>
		<div id=person>
			<div id=name class=name>Bob</div>
			<div id=email class=email>bob@example.com</div>
			<div id=salary class=salary>35000.12</div>
			<div id=birth class=birth>1984-10-12</div>
			<div id=work class=work>40h</div>
			<div id=fulltime class=fulltime>true</div>
		</div>
	</body>
	</html>
	`
	extractor := NewHTMLExtractor(htmlDoc)

	assertExtractorSucc(t, extractor, "//div[@id='number']/div[@class='one']", Int, 1)
	assertExtractorSucc(t, extractor, "//title", String, "Hello World Example")
	assertExtractorSucc(t, extractor, "//div[@id='person']/div[@class='name']", String, "Bob")
	assertExtractorSucc(t, extractor, "//div[@id='person']/div[@class='email']", String, "bob@example.com")
	assertExtractorSucc(t, extractor, "//div[@id='person']/div[@class='salary']", Float, 35000.12)
	expected, _ := tryParseTime("1984-10-12")
	assertExtractorSucc(t, extractor, "//div[@id='person']/div[@class='birth']", Time, expected)
	assertExtractorSucc(t, extractor, "//div[@id='person']/div[@class='work']", Duration, 40*time.Hour)
	assertExtractorSucc(t, extractor, "//div[@id='person']/div[@class='fulltime']", Bool, true)
	// multiple results only return the first one
	assertExtractorSucc(t, extractor, "//div[@id='person']/div", String, "Bob")
	// empty result
	assertExtractorSucc(t, extractor, "//div[@id='person']/div[@class='none']", String, "")
	// invalid xpath
	assertExtractorFail(t, extractor, "///asdf", String, "")
}

func TestJSONExtractor(t *testing.T) {
	jsonDoc := `
	{
		"company": {
			"name": "MegaEase",
			"person": [{
					"name": "Bob",
					"email": "bob@example.com",
					"age": 35,
					"salary": 35000.12,
					"birth": "1984-10-12",
					"work": "40h",
					"fulltime": true
				},
				{
					"name": "Alice",
					"email": "alice@example.com",
					"age": 25,
					"salary": 25000.12,
					"birth": "1985-10-12",
					"work": "30h",
					"fulltime": false
				}
			]
		}
	}`
	extractor := NewJSONExtractor(jsonDoc)

	assertExtractorSucc(t, extractor, "//name", String, "MegaEase")
	assertExtractorSucc(t, extractor, "//company/name", String, "MegaEase")
	assertExtractorSucc(t, extractor, "//email", String, "bob@example.com")
	assertExtractorSucc(t, extractor, "//company/person/*[1]/name", String, "Bob")
	assertExtractorSucc(t, extractor, "//company/person/*[2]/email", String, "alice@example.com")
	assertExtractorSucc(t, extractor, "//company/person/*[last()]/name", String, "Alice")
	assertExtractorSucc(t, extractor, "//company/person/*[last()]/age", Int, 25)
	assertExtractorSucc(t, extractor, "//company/person/*[salary=25000.12]/salary", Float, 25000.12)
	expected, _ := tryParseTime("1984-10-12")
	assertExtractorSucc(t, extractor, "//company/person/*[name='Bob']/birth", Time, expected)
	assertExtractorSucc(t, extractor, "//company/person/*[name='Alice']/work", Duration, 30*time.Hour)
	assertExtractorSucc(t, extractor, "//*/email[contains(.,'bob')]", String, "bob@example.com")
	assertExtractorSucc(t, extractor, "//work", Duration, 40*time.Hour)
	assertExtractorSucc(t, extractor, "//person/*[2]/fulltime", Bool, false)
}

func TestXMLExtractor(t *testing.T) {
	xmlDoc := `
	<company>
		<name>MegaEase</name>
		<person id="emp1001">
			<name>Bob</name>
			<email>bob@example.com</email>
			<age>35</age>
			<salary>35000.12</salary>
			<birth>1984-10-12</birth>
			<work>40h</work>
			<fulltime>true</fulltime>
		</person>
		<person id="emp1002">
			<name>Alice</name>
			<email>alice@example.com</email>
			<age>25</age>
			<salary>25000.12</salary>
			<birth>1985-10-12</birth>
			<work>30h</work>
			<fulltime>false</fulltime>
		</person>
	</company>`
	extractor := NewXMLExtractor(xmlDoc)

	assertExtractorSucc(t, extractor, "//name", String, "MegaEase")
	assertExtractorSucc(t, extractor, "//company/name", String, "MegaEase")
	assertExtractorSucc(t, extractor, "//company/person[1]/name", String, "Bob")
	assertExtractorSucc(t, extractor, "//company/person[last()]/name", String, "Alice")
	assertExtractorSucc(t, extractor, "//person[@id='emp1002']/age", Int, 25)
	assertExtractorSucc(t, extractor, "//company/person[salary=35000.12]/salary", Float, 35000.12)
	assertExtractorSucc(t, extractor, "//person[salary<30000]/salary", Float, 25000.12)
	expected, _ := tryParseTime("1984-10-12")
	assertExtractorSucc(t, extractor, "//company/person[name='Bob']/birth", Time, expected)
	assertExtractorSucc(t, extractor, "//company/person[name='Alice']/work", Duration, 30*time.Hour)
	assertExtractorSucc(t, extractor, "//company/person[name='Bob']/work", Duration, 40*time.Hour)
}

func TestRegexExtractor(t *testing.T) {
	regexDoc := `name: Bob, email: bob@example.com, age: 35, salary: 35000.12, birth: 1984-10-12, work: 40h, fulltime: true`

	extractor := NewRegexExtractor(regexDoc)

	assertExtractorSucc(t, extractor, "name: (?P<name>[a-zA-Z0-9 ]*)", String, "Bob")
	assertExtractorSucc(t, extractor, "email: (?P<email>[a-zA-Z0-9@.]*)", String, "bob@example.com")
	assertExtractorSucc(t, extractor, "age: (?P<age>[0-9]*)", Int, 35)
	assertExtractorSucc(t, extractor, "age: (?P<age>\\d+)", Int, 35)
	assertExtractorSucc(t, extractor, "salary: (?P<salary>[0-9.]*)", Float, 35000.12)
	assertExtractorSucc(t, extractor, "salary: (?P<salary>\\d+\\.\\d+)", Float, 35000.12)
	expected, _ := tryParseTime("1984-10-12")
	assertExtractorSucc(t, extractor, "birth: (?P<birth>[0-9-]*)", Time, expected)
	assertExtractorSucc(t, extractor, "birth: (?P<birth>\\d{4}-\\d{2}-\\d{2})", Time, expected)
	assertExtractorSucc(t, extractor, "work: (?P<work>\\d+[hms])", Duration, 40*time.Hour)
	assertExtractorSucc(t, extractor, "fulltime: (?P<fulltime>true|false)", Bool, true)
	// no Submatch
	assertExtractorSucc(t, extractor, "name: ", String, "name: ")
	// no match
	assertExtractorFail(t, extractor, "mismatch", String, "")
}

func TestFailed(t *testing.T) {
	doc := "<div>hello world</div>"
	extractor := NewHTMLExtractor(doc)
	invalid := "///div"
	assertExtractorFail(t, extractor, invalid, Int, 0)
	assertExtractorFail(t, extractor, invalid, Float, 0.0)
	assertExtractorFail(t, extractor, invalid, Bool, false)
	assertExtractorFail(t, extractor, invalid, Time, time.Time{})
	assertExtractorFail(t, extractor, invalid, Duration, time.Duration(0))
	assertExtractorFail(t, extractor, invalid, Unknown, nil)

	monkey.Patch(html.Parse, func(io.Reader) (*html.Node, error) {
		return nil, errors.New("parse error")
	})

	extractor.VarType = String
	result, err := extractor.Extract()
	assert.NotNil(t, err)
	assert.Equal(t, "parse error", err.Error())
	assert.Equal(t, "", result)

	monkey.Unpatch(html.Parse)
}
