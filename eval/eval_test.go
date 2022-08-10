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
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/Knetic/govaluate"
	"github.com/stretchr/testify/assert"
)

func TestHTMLEval(t *testing.T) {
	htmlTemp := `
	<html>
	<head>
		<title>Hello World</title>
	</head>
	<body>
		<h1>Hello World</h1>
		<p>This is a simple example of a HTML document.</p>
		<div id=time>%s</div>
		<div id=message>service is ok</div>
		<div id=mem_used>512</div>
		<div id=mem_total>1024</div>
		<div id=resp_time>500ms</div>
		<div id=live>false</div>
	</body>
	</html>`
	htmlDoc := fmt.Sprintf(htmlTemp, time.Now().Format(time.RFC3339))

	eval := NewEvaluator(htmlDoc, HTML, "message == 'service is ok'")
	v := NewVariable("message", String, "//div[@id='message']")
	eval.AddVariable(v)
	result, err := eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	eval.CleanVariable()
	eval.AddVariable(NewVariable("title", String, "//h1"))
	eval.Expression = "title =~ 'World'"
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	eval = NewEvaluator(htmlDoc, HTML, "(mem_used / mem_total) < 0.8")
	eval.AddVariable(NewVariable("mem_used", Int, "//div[@id='mem_used']"))
	eval.AddVariable(NewVariable("mem_total", Int, "//div[@id='mem_total']"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	pass := time.Now().Add(-10 * time.Second)
	eval = NewEvaluator(htmlDoc, HTML, "time > '"+pass.Format(time.RFC3339)+"'")
	eval.AddVariable(NewVariable("time", Time, "//div[@id='time']"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	eval = NewEvaluator(htmlDoc, HTML, "!live")
	eval.AddVariable(NewVariable("live", Bool, "//div[@id='live']"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	// test  strlen() function
	eval = NewEvaluator(htmlDoc, HTML, "strlen(title) > 10")
	eval.AddVariable(NewVariable("title", String, "//title"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	// test now() function
	htmlDoc = fmt.Sprintf(htmlTemp, pass.Format(time.RFC3339))
	eval = NewEvaluator(htmlDoc, HTML, "now() - time > 5")
	eval.AddVariable(NewVariable("time", Time, "//div[@id='time']"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	// test duration() function
	eval = NewEvaluator(htmlDoc, HTML, "duration(rt) < 1000")
	eval.AddVariable(NewVariable("rt", String, "//div[@id='resp_time']"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	// test duration() function error
	eval = NewEvaluator(htmlDoc, HTML, "duration(rt) < '1000'")
	eval.AddVariable(NewVariable("rt", String, "//div[@id='time']"))
	result, err = eval.Evaluate()
	assert.NotNil(t, err)
	assert.False(t, result)
}

func TestJSONEval(t *testing.T) {
	json := `{
		"name": "Server",
		"time": "` + time.Now().Format(time.RFC3339) + `",
		"mem_used": 512,
		"mem_total": 1024,
		"resp_time": "500ms"
	}`

	eval := NewEvaluator(json, JSON, "name == 'Server'")
	v := NewVariable("name", String, "//name")
	eval.AddVariable(v)
	result, err := eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	eval = NewEvaluator(json, JSON, `time > '`+time.Now().Add(-10*time.Second).Format(time.RFC3339)+`'`)
	eval.AddVariable(NewVariable("time", Time, "//time"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	eval = NewEvaluator(json, JSON, "(mem_used / mem_total) < 0.8")
	eval.AddVariable(NewVariable("mem_used", Int, "//mem_used"))
	eval.AddVariable(NewVariable("mem_total", Int, "//mem_total"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	eval = NewEvaluator(json, JSON, "duration(rt) ")
	eval.AddVariable(NewVariable("rt", String, "//resp_time"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	eval = NewEvaluator(json, JSON, "name + ' ' + time")
	eval.AddVariable(NewVariable("name", String, "//name"))
	eval.AddVariable(NewVariable("time", String, "//time"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	eval = NewEvaluator(json, JSON, "mem_total - mem_used")
	eval.AddVariable(NewVariable("mem_used", Int, "//mem_used"))
	eval.AddVariable(NewVariable("mem_total", Int, "//mem_total"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestXMLEval(t *testing.T) {
	xmlDoc := `
	<root>
		<name>Server</name>
		<time>` + time.Now().Format(time.RFC3339) + `</time>
		<cpu>0.88</cpu>
		<mem_used>512</mem_used>
		<mem_total>1024</mem_total>
		<resp_time>500ms</resp_time>
		<live>true</live>
	</root>`

	eval := NewEvaluator(xmlDoc, XML, "name == 'Server'")
	v := NewVariable("name", String, "//name")
	eval.AddVariable(v)
	result, err := eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	eval = NewEvaluator(xmlDoc, XML, "live && (mem_used / mem_total) < 0.8 && cpu < 0.9")
	eval.AddVariable(NewVariable("mem_used", Int, "//mem_used"))
	eval.AddVariable(NewVariable("mem_total", Int, "//mem_total"))
	eval.AddVariable(NewVariable("cpu", Float, "//cpu"))
	eval.AddVariable(NewVariable("live", Bool, "//live"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestRegexEval(t *testing.T) {
	text := `name: Server, cpu: 0.8, mem_used: 512, mem_total: 1024, resp_time: 256ms, live: true`

	eval := NewEvaluator(text, TEXT, "name == 'Server'")
	v := NewVariable("name", String, "name: (?P<name>[a-zA-Z0-9 ]*)")
	eval.AddVariable(v)
	result, err := eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)

	eval = NewEvaluator(text, TEXT, "live && (mem_used / mem_total) < 0.8 && cpu < 0.9")
	eval.AddVariable(NewVariable("mem_used", Int, "mem_used: (?P<mem_used>[0-9]*)"))
	eval.AddVariable(NewVariable("mem_total", Int, "mem_total: (?P<mem_total>[0-9]*)"))
	eval.AddVariable(NewVariable("cpu", Float, "cpu: (?P<cpu>[0-9.]*)"))
	eval.AddVariable(NewVariable("live", Bool, "live: (?P<live>true|false)"))
	result, err = eval.Evaluate()
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestFailure(t *testing.T) {
	htmlDoc := `<html></html>`
	eval := NewEvaluator(htmlDoc, HTML, "name == 'Server'")
	v := NewVariable("name", String, "///name")
	eval.AddVariable(v)
	result, err := eval.Evaluate()
	assert.NotNil(t, err)
	assert.False(t, result)

	eval = NewEvaluator("", Unsupported, "")
	result, err = eval.Evaluate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unsupported")
	assert.False(t, result)

	eval = NewEvaluator(htmlDoc, HTML, "name == 'Server'")
	v = NewVariable("name", String, "//name")
	eval.AddVariable(v)

	var expression govaluate.EvaluableExpression
	monkey.PatchInstanceMethod(reflect.TypeOf(expression), "Evaluate", func(govaluate.EvaluableExpression, map[string]interface{}) (interface{}, error) {
		n := 10
		return n, nil
	})

	result, err = eval.Evaluate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Unsupported")
	assert.False(t, result)

	monkey.Patch(govaluate.NewEvaluableExpressionWithFunctions, func(expression string, functions map[string]govaluate.ExpressionFunction) (*govaluate.EvaluableExpression, error) {
		return nil, errors.New("error")
	})

	result, err = eval.Evaluate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error")
	assert.False(t, result)

	monkey.UnpatchAll()
}
