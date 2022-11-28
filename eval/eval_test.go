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

func assertResult(t *testing.T, eval *Evaluator, success bool) {
	result, err := eval.Evaluate()
	if success {
		assert.Nil(t, err)
		assert.True(t, result)
	} else {
		assert.NotNil(t, err)
		assert.False(t, result)
	}
}

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

	// ---- test message ----
	eval := NewEvaluator(htmlDoc, HTML, "message == 'service is ok'")
	v := NewVariable("message", String, `//div[@id="message"]`)
	eval.AddVariable(v)
	assertResult(t, eval, true)

	eval = NewEvaluator(htmlDoc, HTML, "x_str('//div[@id=\\'message\\']') == 'service is ok'")
	assertResult(t, eval, true)

	// ---- test title ----
	eval.CleanVariable()
	eval.AddVariable(NewVariable("title", String, "//h1"))
	eval.Expression = "title =~ 'World'"
	assertResult(t, eval, true)

	eval = NewEvaluator(htmlDoc, HTML, "x_str('//h1') =~ 'World'")
	assertResult(t, eval, true)

	// ---- test memory ----
	eval = NewEvaluator(htmlDoc, HTML, "(mem_used / mem_total) < 0.8")
	eval.AddVariable(NewVariable("mem_used", Int, "//div[@id='mem_used']"))
	eval.AddVariable(NewVariable("mem_total", Int, "//div[@id='mem_total']"))
	assertResult(t, eval, true)

	eval = NewEvaluator(htmlDoc, HTML, "x_int('//div[@id=\\'mem_used\\']') / x_int('//div[@id=\\'mem_total\\']') < 0.8")
	assertResult(t, eval, true)

	// ---- test time ----
	pass := time.Now().Add(-10 * time.Second)
	eval = NewEvaluator(htmlDoc, HTML, "time > '"+pass.Format(time.RFC3339)+"'")
	eval.AddVariable(NewVariable("time", Time, "//div[@id='time']"))
	assertResult(t, eval, true)

	eval = NewEvaluator(htmlDoc, HTML, "x_time('//div[@id=\\'time\\']') > '"+pass.Format(time.RFC3339)+"'")
	assertResult(t, eval, true)

	// ---- test live ----
	eval = NewEvaluator(htmlDoc, HTML, "!live")
	eval.AddVariable(NewVariable("live", Bool, "//div[@id='live']"))
	assertResult(t, eval, true)

	eval = NewEvaluator(htmlDoc, HTML, "!x_bool('//div[@id=\\'live\\']')")
	assertResult(t, eval, true)

	// test  strlen() function
	eval = NewEvaluator(htmlDoc, HTML, "strlen(title) > 10")
	eval.AddVariable(NewVariable("title", String, "//title"))
	assertResult(t, eval, true)

	eval = NewEvaluator(htmlDoc, HTML, "strlen(x_str('//title')) > 10")
	assertResult(t, eval, true)

	// test now() function
	htmlDoc = fmt.Sprintf(htmlTemp, pass.Format(time.RFC3339))
	eval = NewEvaluator(htmlDoc, HTML, "now() - time > 5")
	eval.AddVariable(NewVariable("time", Time, "//div[@id='time']"))
	assertResult(t, eval, true)

	eval = NewEvaluator(htmlDoc, HTML, "now() - x_time('//div[@id=\\'time\\']') > 5")
	assertResult(t, eval, true)

	// test duration() function
	eval = NewEvaluator(htmlDoc, HTML, "duration(rt) < duration('1s')")
	eval.AddVariable(NewVariable("rt", String, "//div[@id='resp_time']"))
	assertResult(t, eval, true)

	eval = NewEvaluator(htmlDoc, HTML, "x_duration('//div[@id=\\'resp_time\\']') < duration('1s')")
	assertResult(t, eval, true)

	// test duration() function error
	eval = NewEvaluator(htmlDoc, HTML, "duration(rt) < '1000'")
	eval.AddVariable(NewVariable("rt", String, "//div[@id='time']"))
	assertResult(t, eval, false)

	eval = NewEvaluator(htmlDoc, HTML, "duration(x_str('//div[@id=\\'time\\']')) < '1000'")
	assertResult(t, eval, false)
}

func TestJSONEval(t *testing.T) {
	json := `{
		"name": "Server",
		"time": "` + time.Now().Format(time.RFC3339) + `",
		"mem_used": 512,
		"mem_total": 1024,
		"resp_time": "500ms"
	}`

	// ---- test name ----
	eval := NewEvaluator(json, JSON, "name == 'Server'")
	v := NewVariable("name", String, "//name")
	eval.AddVariable(v)
	assertResult(t, eval, true)

	eval = NewEvaluator(json, JSON, "x_str('//name') == 'Server'")
	assertResult(t, eval, true)

	// ---- test time ----
	eval = NewEvaluator(json, JSON, `time > '`+time.Now().Add(-10*time.Second).Format(time.RFC3339)+`'`)
	eval.AddVariable(NewVariable("time", Time, "//time"))
	assertResult(t, eval, true)

	eval = NewEvaluator(json, JSON, "x_time('//time') > '"+time.Now().Add(-10*time.Second).Format(time.RFC3339)+"'")
	assertResult(t, eval, true)

	// ---- test memory ----
	eval = NewEvaluator(json, JSON, "(mem_used / mem_total) < 0.8")
	eval.AddVariable(NewVariable("mem_used", Int, "//mem_used"))
	eval.AddVariable(NewVariable("mem_total", Int, "//mem_total"))
	assertResult(t, eval, true)

	eval = NewEvaluator(json, JSON, "x_int('//mem_used') / x_int('//mem_total') < 0.8")
	assertResult(t, eval, true)

	// ---- test resp_time ----
	eval = NewEvaluator(json, JSON, "duration(rt) ")
	eval.AddVariable(NewVariable("rt", String, "//resp_time"))
	assertResult(t, eval, true)

	eval = NewEvaluator(json, JSON, "x_duration('//resp_time') < duration('1s')")
	assertResult(t, eval, true)

	// ---- test string concat ----
	eval = NewEvaluator(json, JSON, "name + ' ' + time")
	eval.AddVariable(NewVariable("name", String, "//name"))
	eval.AddVariable(NewVariable("time", String, "//time"))
	assertResult(t, eval, true)

	eval = NewEvaluator(json, JSON, "strlen(x_str('//name') + ' ' + x_str('//time')) > 10")
	assertResult(t, eval, true)

	// ----- test minus ----
	eval = NewEvaluator(json, JSON, "mem_total - mem_used")
	eval.AddVariable(NewVariable("mem_used", Int, "//mem_used"))
	eval.AddVariable(NewVariable("mem_total", Int, "//mem_total"))
	assertResult(t, eval, true)

	eval = NewEvaluator(json, JSON, "x_int('//mem_total') - x_int('//mem_used')")
	assertResult(t, eval, true)
}

func TestXMLEval(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	xmlDoc := `
	<root>
		<name>Server</name>
		<time>` + now + `</time>
		<cpu>0.88</cpu>
		<mem_used>512</mem_used>
		<mem_total>1024</mem_total>
		<resp_time>500ms</resp_time>
		<live>true</live>
	</root>`

	// ---- test name ----
	eval := NewEvaluator(xmlDoc, XML, "name == 'Server'")
	v := NewVariable("name", String, "//name")
	eval.AddVariable(v)
	assertResult(t, eval, true)

	eval = NewEvaluator(xmlDoc, XML, "x_str('//name') == 'Server'")
	assertResult(t, eval, true)

	eval.CleanVariable()
	eval.Expression = "x_str('//name') == 'Server'"
	assertResult(t, eval, true)

	// ---- test time ----
	eval = NewEvaluator(xmlDoc, XML, `t == '`+now+`'`)
	eval.AddVariable(NewVariable("t", Time, "//time"))
	assertResult(t, eval, true)

	eval.CleanVariable()
	eval.Expression = "x_time('//time') == '" + now + "'"
	assertResult(t, eval, true)

	// ---- test cpu , memory and live ----
	eval = NewEvaluator(xmlDoc, XML, "live && (mem_used / mem_total) < 0.8 && cpu < 0.9")
	eval.AddVariable(NewVariable("mem_used", Int, "//mem_used"))
	eval.AddVariable(NewVariable("mem_total", Int, "//mem_total"))
	eval.AddVariable(NewVariable("cpu", Float, "//cpu"))
	eval.AddVariable(NewVariable("live", Bool, "//live"))
	assertResult(t, eval, true)

	eval = NewEvaluator(xmlDoc, XML, "x_bool('//live') && x_float('//cpu') < 0.9 && x_int('//mem_used') / x_int('//mem_total') < 0.8")
	assertResult(t, eval, true)
}

func TestRegexEval(t *testing.T) {
	text := `name: Server, cpu: 0.8, mem_used: 512, mem_total: 1024, resp_time: 256ms, live: true`

	// ---- test name ----
	eval := NewEvaluator(text, TEXT, "name == 'Server'")
	v := NewVariable("name", String, "name: (?P<name>[a-zA-Z0-9 ]*)")
	eval.AddVariable(v)
	assertResult(t, eval, true)

	eval = NewEvaluator(text, TEXT, "x_str('name: (?P<name>[a-zA-Z0-9 ]*)') == 'Server'")
	assertResult(t, eval, true)

	// ---- test live memory cpu ----
	eval = NewEvaluator(text, TEXT, "live && (mem_used / mem_total) < 0.8 && cpu < 0.9")
	eval.AddVariable(NewVariable("mem_used", Int, "mem_used: (?P<mem_used>[0-9]*)"))
	eval.AddVariable(NewVariable("mem_total", Int, "mem_total: (?P<mem_total>[0-9]*)"))
	eval.AddVariable(NewVariable("cpu", Float, "cpu: (?P<cpu>[0-9.]*)"))
	eval.AddVariable(NewVariable("live", Bool, "live: (?P<live>true|false)"))
	assertResult(t, eval, true)

	eval = NewEvaluator(text, TEXT, "x_bool('live: (?P<live>true|false)') && x_float('cpu: (?P<cpu>[0-9.]*)') < 0.9 && x_int('mem_used: (?P<mem_used>[0-9]*)') / x_int('mem_total: (?P<mem_total>[0-9]*)') < 0.8")
	assertResult(t, eval, true)

	// ---- test mix usage ----
	// - retrieve name and mem_used from extract function
	// - set live, cpu and mem_total as the variables
	eval = NewEvaluator(text, TEXT, "x_str('name: (?P<name>[a-zA-Z0-9 ]*)')  == 'Server' && live && (x_int('mem_used: (?P<mem_used>[0-9]*)') / mem_total) < 0.8 && cpu < 0.9")
	eval.AddVariable(NewVariable("mem_total", Int, "mem_total: (?P<mem_total>[0-9]*)"))
	eval.AddVariable(NewVariable("cpu", Float, "cpu: (?P<cpu>[0-9.]*)"))
	eval.AddVariable(NewVariable("live", Bool, "live: (?P<live>true|false)"))
	assertResult(t, eval, true)
}

func TestFailure(t *testing.T) {
	htmlDoc := `<html></html>`
	eval := NewEvaluator(htmlDoc, HTML, "name == 'Server'")
	v := NewVariable("name", String, "///name")
	eval.AddVariable(v)
	assertResult(t, eval, false)

	eval = NewEvaluator("", Unsupported, "")
	v = NewVariable("name", String, "//name")
	eval.AddVariable(v)
	assertResult(t, eval, false)
	result, err := eval.Evaluate()
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

func TestExtractFunc(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	xmlDoc := `
	<root>
		<name>Server</name>
		<time>` + now + `</time>
		<cpu>0.88</cpu>
		<mem_used>512</mem_used>
		<mem_total>1024</mem_total>
		<resp_time>500ms</resp_time>
		<live>true</live>
		<date>2022-08-11 10:10:10</date>
	</root>`

	eval := NewEvaluator(xmlDoc, XML, "x_str('//name') == 'Server'")
	assertResult(t, eval, true)

	eval = NewEvaluator(xmlDoc, XML, "x_time('//time') == '"+now+"'")
	assertResult(t, eval, true)

	eval = NewEvaluator(xmlDoc, XML, "x_float('//cpu') < 0.9")
	assertResult(t, eval, true)

	eval = NewEvaluator(xmlDoc, XML, "x_int('//mem_used') < 800")
	assertResult(t, eval, true)

	eval = NewEvaluator(xmlDoc, XML, "x_int('//mem_used') / x_int('//mem_total') < 0.8")
	assertResult(t, eval, true)

	eval = NewEvaluator(xmlDoc, XML, "x_duration('//resp_time') < duration('1000ms')")
	assertResult(t, eval, true)

	eval = NewEvaluator(xmlDoc, XML, "x_bool('//live')")
	assertResult(t, eval, true)

	eval = NewEvaluator(xmlDoc, XML, "x_time('//date') == '2022-08-11 10:10:10'")
	assertResult(t, eval, true)

	eval = NewEvaluator(xmlDoc, XML, "x_time('//error') == '2022-08-11 10:10:11'")
	assertResult(t, eval, false)

	eval = NewEvaluator(xmlDoc, XML, "x_int('//error') < 10")
	assertResult(t, eval, false)
}

func TestSetDocument(t *testing.T) {
	text := `name: Server, cpu: 0.8, mem_used: 512, mem_total: 1024, resp_time: 256ms, live: true`

	eval := NewEvaluator(text, TEXT, "x_str('name: (?P<name>[a-zA-Z0-9 ]*)') == 'Server'")
	assertResult(t, eval, true)

	eval.SetDocument(TEXT, "name: Server, cpu: 0.5, mem_used: 512, mem_total: 1024, resp_time: 256ms, live: true")
	eval.Expression = "x_float('cpu: (?P<cpu>[0-9.]*)') < 0.6"
	assertResult(t, eval, true)

	eval.SetDocument(JSON, `{"name": "Server", "cpu": 0.3, "mem_used": 512, "mem_total": 1024, "resp_time": "256ms", "live": true}`)
	eval.Expression = "x_float('//cpu') < 0.4"
	assertResult(t, eval, true)
}

func TestSpringActuator(t *testing.T) {
	errMsg := "com.sun.mail.util.MailConnectException: Couldn't connect to host, port: smtp.example.com, 465; timeout 10"
	jsonDoc := `{
		"status":"DOWN",
		"components":{
		  "diskSpace":{
			"status":"UP",
			"details":{
			  "total":41954803712,
			  "free":8418234368,
			  "threshold":10485760,
			  "exists":true
			}
		  },
		  "hystrix":{
			"status":"UP"
		  },
		  "livenessState":{
			"status":"UP"
		  },
		  "mail":{
			"status":"DOWN",
			"details":{
			  "location":"smtp.example.com:465",
			  "error": "` + errMsg + `"
			}
		  },
		  "ping":{
			"status":"UP"
		  }
		},
		"groups":[
		  "liveness",
		  "readiness"
		]
	  }`

	eval := NewEvaluator(jsonDoc, JSON, "x_str('//status') == 'UP'")
	assertResult(t, eval, true)

	exp := "//components/*/details/error"
	eval = NewEvaluator(jsonDoc, JSON, "x_str('"+exp+"') != ''")
	assertResult(t, eval, true)
	assert.Equal(t, errMsg, eval.ExtractedValues[exp])
}
