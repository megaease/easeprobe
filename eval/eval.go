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

// Package eval is the package to eval the expression and extract the value from the document
package eval

import (
	"fmt"
	"time"

	"github.com/Knetic/govaluate"
	log "github.com/sirupsen/logrus"
)

// Variable is the variable type
type Variable struct {
	Name  string      `yaml:"name" json:"name" jsonschema:"required,title=Variable Name,description=Variable Name"`
	Type  VarType     `yaml:"type" json:"type" jsonschema:"required,type=string,enum=int,enum=string,enum=bool,enum=float,enum=bool,enum=time,enum=duration,title=Variable Type,description=Variable Type"`
	Query string      `yaml:"query" json:"query" jsonschema:"required,title=Query,description=XPath/Regex Expression to extract the value"`
	Value interface{} `yaml:"-" json:"-"`
}

// NewVariable is the function to create a variable
func NewVariable(name string, t VarType, query string) *Variable {
	return &Variable{
		Name:  name,
		Type:  t,
		Query: query,
		Value: nil,
	}
}

// Evaluator is the structure of evaluator
type Evaluator struct {
	Variables  []Variable                              `yaml:"variables,omitempty" json:"variables,omitempty" jsonschema:"title=Variables Definition,description=define the variables used in the expression"`
	DocType    DocType                                 `yaml:"doc" json:"doc" jsonschema:"required,type=string,enum=html,enum=xml,enum=json,enum=text,title=Document Type,description=Document Type"`
	Expression string                                  `yaml:"expression" json:"expression" jsonschema:"required,title=Expression,description=Expression need to be evaluated"`
	Document   string                                  `yaml:"-" json:"-"`
	Extractor  Extractor                               `yaml:"-" json:"-"`
	EvalFuncs  map[string]govaluate.ExpressionFunction `yaml:"-" json:"-"`

	ExtractedValues map[string]interface{} `yaml:"-" json:"-"`
}

// NewEvaluator is the function to create a evaluator
func NewEvaluator(doc string, t DocType, exp string) *Evaluator {
	e := &Evaluator{
		Variables:  make([]Variable, 0),
		DocType:    t,
		Expression: exp,
		Document:   doc,
	}
	e.Config()
	return e
}

// Config is the function to config the evaluator
func (e *Evaluator) Config() error {
	e.configExtractor()
	e.configEvalFunctions()
	return nil
}

func (e *Evaluator) configExtractor() {
	e.ExtractedValues = make(map[string]interface{})
	switch e.DocType {
	case HTML:
		e.Extractor = NewHTMLExtractor(e.Document)
	case XML:
		e.Extractor = NewXMLExtractor(e.Document)
	case JSON:
		e.Extractor = NewJSONExtractor(e.Document)
	case TEXT:
		e.Extractor = NewRegexExtractor(e.Document)
	default:
		e.Extractor = nil
		log.Errorf("Unsupported document type: %s", e.DocType)
	}
}

func (e *Evaluator) configEvalFunctions() {

	extract := func(t VarType, query string, failed interface{}) (interface{}, error) {
		v := Variable{
			Type:  t,
			Query: query,
		}
		if err := e.ExtractValue(&v); err != nil {
			return failed, err
		}
		return v.Value, nil
	}

	e.EvalFuncs = map[string]govaluate.ExpressionFunction{

		// Extract value by XPath/Regex Expression
		"x_str": func(args ...interface{}) (interface{}, error) {
			return extract(String, args[0].(string), "")
		},
		"x_float": func(args ...interface{}) (interface{}, error) {
			return extract(Float, args[0].(string), 0.0)
		},
		"x_int": func(args ...interface{}) (interface{}, error) {
			v, e := extract(Int, args[0].(string), 0)
			return float64(v.(int)), e
		},
		"x_bool": func(args ...interface{}) (interface{}, error) {
			return extract(Bool, args[0].(string), false)
		},
		"x_time": func(args ...interface{}) (interface{}, error) {
			v := Variable{
				Type:  Time,
				Query: args[0].(string),
			}

			if err := e.ExtractValue(&v); err != nil {
				return (time.Time{}), err
			}
			return (float64)(v.Value.(int64)), nil
		},
		"x_duration": func(args ...interface{}) (interface{}, error) {
			v, e := extract(Duration, args[0].(string), 0)
			return (float64)(v.(time.Duration)), e
		},

		// Functional functions
		"strlen": func(args ...interface{}) (interface{}, error) {
			length := len(args[0].(string))
			return (float64)(length), nil
		},
		"now": func(args ...interface{}) (interface{}, error) {
			return (float64)(time.Now().Unix()), nil
		},
		"duration": func(args ...interface{}) (interface{}, error) {
			str := args[0].(string)
			d, err := time.ParseDuration(str)
			if err != nil {
				return nil, err
			}
			return (float64)(d), nil
		},
	}
}

// SetDocument is the function to set the document
func (e *Evaluator) SetDocument(t DocType, doc string) {
	if e.DocType != t {
		e.DocType = t
		e.Document = doc
		e.configExtractor()
	} else {
		e.Document = doc
		e.Extractor.SetDocument(doc)
	}
}

// AddVariable is the function to add a variable
func (e *Evaluator) AddVariable(v *Variable) {
	e.Variables = append(e.Variables, *v)
}

// CleanVariable is the function to clean the variable
func (e *Evaluator) CleanVariable() {
	e.Variables = make([]Variable, 0)
}

// Evaluate is the function to evaluate the expression
func (e *Evaluator) Evaluate() (bool, error) {

	if err := e.Extract(); err != nil {
		return false, err
	}

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(e.Expression, e.EvalFuncs)
	if err != nil {
		return false, err
	}

	variables := make(map[string]interface{})
	for _, v := range e.Variables {
		variables[v.Name] = v.Value
	}

	result, err := expression.Evaluate(variables)
	if err != nil {
		return false, err
	}
	switch result.(type) {
	case bool:
		return result.(bool), nil
	case float64:
		return result.(float64) != 0, nil
	case string:
		return result.(string) != "", nil
	}
	return false, fmt.Errorf("Unsupported type: %T", result)
}

// Extract is the function to extract the value from the document
func (e *Evaluator) Extract() error {
	for i := 0; i < len(e.Variables); i++ {
		if err := e.ExtractValue(&e.Variables[i]); err != nil {
			return err
		}
	}
	return nil
}

// ExtractValue is the function to extract the value from the document
func (e *Evaluator) ExtractValue(v *Variable) error {
	if e.DocType == Unsupported || e.Extractor == nil {
		return fmt.Errorf("Unsupported document type: %s", e.DocType)
	}
	e.Extractor.SetQuery(v.Query)
	e.Extractor.SetVarType(v.Type)
	value, err := e.Extractor.Extract()
	if err != nil {
		return err
	}
	if v.Type == Time {
		v.Value = value.(time.Time).Local().Unix()
	} else {
		v.Value = value
	}
	e.ExtractedValues[v.Query] = v.Value
	return nil
}
