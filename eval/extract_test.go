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

	extractor.XPath = "//div[@id='number']/div[@class='one']"
	extractor.VarType = Int
	result, err := extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, 1, result)

	extractor.XPath = "//title"
	extractor.VarType = String
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "Hello World Example", result)

	extractor.XPath = "//div[@id='person']/div[@class='name']"
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "Bob", result)

	extractor.XPath = "//div[@class='email']"
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "bob@example.com", result)

	extractor.XPath = "//div[@id='person']/div[@class='salary']"
	extractor.VarType = Float
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, 35000.12, result)

	extractor.XPath = "//div[@id='person']/div[@class='birth']"
	extractor.VarType = Time
	result, err = extractor.Extract()
	assert.Nil(t, err)
	expected, _ := tryParseTime("1984-10-12")
	assert.Equal(t, expected, result)

	extractor.XPath = "//div[@id='person']/div[@class='work']"
	extractor.VarType = Duration
	result, err = extractor.Extract()
	assert.Nil(t, err)
	duration, _ := time.ParseDuration("40h")
	assert.Equal(t, duration, result)

	extractor.XPath = "//div[@id='person']/div[@class='fulltime']"
	extractor.VarType = Bool
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	// multiple results only return the first one
	extractor.XPath = "//div[@id='person']/div"
	result, err = extractor.ExtractStrFn()
	assert.Nil(t, err)
	assert.Equal(t, "Bob", result)

	// empty result
	extractor.XPath = "//div[@id='person']/div[@class='none']"
	extractor.VarType = String
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "", result)

	// invalid xpath
	extractor.XPath = "///ads']"
	result, err = extractor.Extract()
	assert.NotNil(t, err)
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

	extractor.XPath = "//name"
	extractor.VarType = String
	result, err := extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "MegaEase", result)

	extractor.XPath = "//email"
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "bob@example.com", result)

	extractor.XPath = "//company/person/*[1]/name"
	extractor.VarType = String
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "Bob", result)

	extractor.XPath = "//company/person/*[last()]/name"
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "Alice", result)

	extractor.XPath = "//company/person/*[last()]/age"
	extractor.VarType = Int
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, 25, result)

	extractor.XPath = "//company/person/*[salary=25000.12]/salary"
	extractor.VarType = Float
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, 25000.12, result)

	extractor.XPath = "//company/person/*[name='Bob']/birth"
	extractor.VarType = Time
	result, err = extractor.Extract()
	assert.Nil(t, err)
	expected, _ := tryParseTime("1984-10-12")
	assert.Equal(t, expected, result)

	extractor.XPath = "//*/email[contains(.,'bob')]"
	extractor.VarType = String
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "bob@example.com", result)

	extractor.XPath = "//work"
	extractor.VarType = Duration
	result, err = extractor.Extract()
	assert.Nil(t, err)
	duration, _ := time.ParseDuration("40h")
	assert.Equal(t, duration, result)

	extractor.XPath = "//person/*[2]/fulltime"
	extractor.VarType = Bool
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, false, result)
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

	extractor.XPath = "//name"
	extractor.VarType = String
	result, err := extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "MegaEase", result)

	extractor.XPath = "//company/person[1]/name"
	extractor.VarType = String
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "Bob", result)

	extractor.XPath = "//company/person[last()]/name"
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "Alice", result)

	extractor.XPath = "//person[@id='emp1002']/age"
	extractor.VarType = Int
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, 25, result)

	extractor.XPath = "//person[salary<30000]/salary"
	extractor.VarType = Float
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, 25000.12, result)

	extractor.XPath = "//person[name='Bob']/birth"
	extractor.VarType = Time
	result, err = extractor.Extract()
	assert.Nil(t, err)
	expected, _ := tryParseTime("1984-10-12")
	assert.Equal(t, expected, result)

	extractor.XPath = "//person[name='Bob']/work"
	extractor.VarType = Duration
	result, err = extractor.Extract()
	assert.Nil(t, err)
	duration, _ := time.ParseDuration("40h")
	assert.Equal(t, duration, result)
}

func TestRegexExtractor(t *testing.T) {
	regexDoc := `name: Bob, email: bob@example.com, age: 35, salary: 35000.12, birth: 1984-10-12, work: 40h, fulltime: true`

	extractor := NewRegexExtractor(regexDoc)
	extractor.Regex = "name: (?P<name>[a-zA-Z0-9 ]*)"
	extractor.VarType = String
	result, err := extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "Bob", result)

	extractor.Regex = "email: ([a-zA-Z0-9@.]*)"
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, "bob@example.com", result)

	extractor.Regex = "age: (?P<age>\\d+)"
	extractor.VarType = Int
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, 35, result)

	extractor.Regex = "salary: (?P<salary>\\d+\\.\\d+)"
	extractor.VarType = Float
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, 35000.12, result)

	extractor.Regex = "birth: (?P<birth>\\d{4}-\\d{2}-\\d{2})"
	extractor.VarType = Time
	result, err = extractor.Extract()
	assert.Nil(t, err)
	expected, _ := tryParseTime("1984-10-12")
	assert.Equal(t, expected, result)

	extractor.Regex = "work: (?P<work>\\d+[hms])"
	extractor.VarType = Duration
	result, err = extractor.Extract()
	assert.Nil(t, err)
	duration, _ := time.ParseDuration("40h")
	assert.Equal(t, duration, result)

	extractor.Regex = "fulltime: (?P<fulltime>true|false)"
	extractor.VarType = Bool
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, true, result)

	// no Submatch
	extractor.Regex = "name: "
	extractor.VarType = String
	result, err = extractor.Extract()
	assert.Nil(t, err)
	assert.Equal(t, extractor.Regex, result)

	// no match
	extractor.Regex = "mismatch"
	extractor.VarType = String
	result, err = extractor.Extract()
	assert.NotNil(t, err)
	assert.Equal(t, "", result)
}

func TestFailed(t *testing.T) {
	doc := "<div>hello world</div>"
	extractor := NewHTMLExtractor(doc)
	extractor.XPath = "///div"
	extractor.VarType = Int
	result, err := extractor.Extract()
	assert.NotNil(t, err)
	assert.Equal(t, 0, result)

	extractor.VarType = Float
	result, err = extractor.Extract()
	assert.NotNil(t, err)
	assert.Equal(t, 0.0, result)

	extractor.VarType = Bool
	result, err = extractor.Extract()
	assert.NotNil(t, err)
	assert.Equal(t, false, result)

	extractor.VarType = Time
	result, err = extractor.Extract()
	assert.NotNil(t, err)
	assert.Equal(t, time.Time{}, result)

	extractor.VarType = Duration
	result, err = extractor.Extract()
	assert.NotNil(t, err)
	assert.Equal(t, time.Duration(0), result)

	extractor.VarType = Unknown
	result, err = extractor.Extract()
	assert.NotNil(t, err)
	assert.Nil(t, result)

	monkey.Patch(html.Parse, func(io.Reader) (*html.Node, error) {
		return nil, errors.New("parse error")
	})

	extractor.VarType = String
	result, err = extractor.Extract()
	assert.NotNil(t, err)
	assert.Equal(t, "parse error", err.Error())

	monkey.Unpatch(html.Parse)
}
