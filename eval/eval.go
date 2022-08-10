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
	"time"

	"github.com/Knetic/govaluate"
)

// Variable is the variable type
type Variable struct {
	Name       string  `yaml:"name"`
	Type       VarType `yaml:"type"`
	Query      string  `yaml:"query"`
	TimeFormat string  `yaml:"time_format"`
	Value      interface{}
}

// NewVariable is the function to create a variable
func NewVariable(name string, t VarType, query string) *Variable {
	return &Variable{
		Name:       name,
		Type:       t,
		Query:      query,
		TimeFormat: time.RFC3339,
		Value:      nil,
	}
}

// Evaluator is the structure of evaluator
type Evaluator struct {
	Variables  []Variable `yaml:"variables"`
	DocType    DocType    `yaml:"doc"`
	Expression string     `yaml:"expression"`
	Document   string     `yaml:"-"`
}

// NewEvaluator is the function to create a evaluator
func NewEvaluator(doc string, t DocType, exp string) *Evaluator {
	return &Evaluator{
		Variables:  make([]Variable, 0),
		DocType:    t,
		Expression: exp,
		Document:   doc,
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

	functions := map[string]govaluate.ExpressionFunction{
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
			return (float64)(d.Milliseconds()), nil
		},
	}

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(e.Expression, functions)
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
	var extractor Extractor
	switch e.DocType {
	case HTML:
		extractor = NewHTMLExtractor(e.Document)
	case XML:
		extractor = NewXMLExtractor(e.Document)
	case JSON:
		extractor = NewJSONExtractor(e.Document)
	case TEXT:
		extractor = NewRegexExtractor(e.Document)
	default:
		return fmt.Errorf("Unsupported document type: %s", e.DocType)
	}
	for i := 0; i < len(e.Variables); i++ {
		v := &e.Variables[i]
		extractor.SetQuery(v.Query)
		extractor.SetVarType(v.Type)
		extractor.SetTimeFormat(v.TimeFormat)
		value, err := extractor.Extract()
		if err != nil {
			return err
		}
		if v.Type == Time {
			v.Value = value.(time.Time).Unix()
		} else {
			v.Value = value
		}
	}
	return nil
}
