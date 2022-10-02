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

package conf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type (
	// YAML has three fundamental types. When unmarshaled into interface{},
	// they're represented like this.
	mapping = map[string]interface{}
	array   = []interface{}
	scalar  = interface{}
)

// mergeYamlFiles deep-merges any number of mergeYamlFiles sources, with later sources taking
// priority over earlier ones.
//
// Maps are deep-merged. For example,
//
//	{"one": 1, "two": 2} + {"one": 42, "three": 3}
//	== {"one": 42, "two": 2, "three": 3}
//
// Arrays are appended. For example,
//
//	{"foo": [1, 2, 3]} + {"foo": [4, 5, 6]}
//	== {"foo": [1, 2, 3, 4, 5, 6]}
func mergeYamlFiles(path string) ([]byte, error) {
	var merged interface{}
	var hasContent bool

	apath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	files, err := filepath.Glob(apath + "/*.yaml")
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		log.Infof("Find file: %v", f)
		b, err := os.ReadFile(f)
		if err != nil {
			log.Errorf("Failed to read file: %v by: %v", f, err)
			continue
		}
		d := yaml.NewDecoder(bytes.NewReader(b))

		var contents interface{}
		if err := d.Decode(&contents); err == io.EOF {
			// Skip empty and comment-only sources, which we should handle
			// differently from explicit nils.
			continue
		} else if err != nil {
			log.Errorf("Failed to decode file: %v by: %v", f, err)
			continue
		}

		hasContent = true
		pair, err := merge(merged, contents)
		if err != nil {
			return nil, err // error is already descriptive enough
		}
		merged = pair
		log.Debugf("Merged result: %v", merged)
	}

	buf := &bytes.Buffer{}
	if !hasContent {
		// No sources had any content. To distinguish this from a source with just
		// an explicit top-level null, return an empty buffer.
		log.Infof("yaml content is empty")
		return buf.Bytes(), nil
	}
	enc := yaml.NewEncoder(buf)
	if err := enc.Encode(merged); err != nil {
		return nil, fmt.Errorf("couldn't re-serialize merged YAML: %v", err)
	}
	return buf.Bytes(), nil
}

// merge merges different types using different rules
func merge(into, from interface{}) (interface{}, error) {
	// It's possible to handle this with a mass of reflection, but we only need
	// to merge whole YAML files. Since we're always unmarshaling into
	// interface{}, we only need to handle a few types. This ends up being
	// cleaner if we just handle each case explicitly.
	log.Debugf("Merge %v into %v", from, into)
	if into == nil {
		return from, nil
	}
	if from == nil {
		return into, nil
	}
	if IsArray(into) && IsArray(from) {
		return mergeArray(into.(array), from.(array))
	}
	if IsScalar(into) && IsScalar(from) {
		return from, nil
	}
	if IsMapping(into) && IsMapping(from) {
		return mergeMapping(into.(mapping), from.(mapping))
	}
	return nil, fmt.Errorf("can't merge a %s into a %s", describe(from), describe(into))
}

// mergeArray appends arrays in from to into
func mergeArray(into, from []interface{}) ([]interface{}, error) {
	var arr []interface{}
	for _, i := range from {
		arr = append(into, i)
	}
	return arr, nil
}

// mergeMapping merges by key and merged value
func mergeMapping(into, from mapping) (mapping, error) {
	merged := make(mapping, len(into))
	for k, v := range into {
		merged[k] = v
	}
	for k := range from {
		m, err := merge(merged[k], from[k])
		if err != nil {
			return nil, err
		}
		merged[k] = m
	}
	return merged, nil
}

// IsMapping reports whether a type is a mapping in YAML, represented as a
// map[interface{}]interface{}.
func IsMapping(i interface{}) bool {
	_, is := i.(mapping)
	return is
}

// IsArray reports whether a type is a Array in YAML, represented as an
// []interface{}.
func IsArray(i interface{}) bool {
	_, is := i.(array)
	return is
}

// IsScalar reports whether a type is a scalar value in YAML.
func IsScalar(i interface{}) bool {
	return !IsMapping(i) && !IsArray(i)
}

// describe describes data type
func describe(i interface{}) string {
	if IsMapping(i) {
		return "mapping"
	}
	if IsArray(i) {
		return "array"
	}
	return "scalar"
}
