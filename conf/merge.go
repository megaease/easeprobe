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
	"bufio"
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

// mergeYamlFiles merges yaml files using yq(https://github.com/mikefarah/yq)
func mergeYamlFiles(path string) ([]byte, error) {
	files, err := filepath.Glob(filepath.Join(path, "*.yaml"))
	if err != nil {
		return nil, err
	}
	moreFiles, err := filepath.Glob(filepath.Join(path, "*.yml"))
	if err == nil {
		files = append(files, moreFiles...)
	}

	if len(files) <= 0 {
		return nil, fmt.Errorf("yaml files not found for %v", path)
	}
	var buf bytes.Buffer
	preference := yqlib.YamlPreferences{
		LeadingContentPreProcessing: true,
		PrintDocSeparators:          true,
		UnwrapScalar:                false,
		EvaluateTogether:            false,
	}
	decoder := yqlib.NewYamlDecoder(preference)
	encoder := yqlib.NewYamlEncoder(preference)
	printer := yqlib.NewPrinter(encoder, yqlib.NewSinglePrinterWriter(bufio.NewWriter(&buf)))
	// use evaluate merge, reference https://mikefarah.gitbook.io/yq/operators/multiply-merge
	err = yqlib.NewAllAtOnceEvaluator().EvaluateFiles(". as $item ireduce ({}; . *+ $item )", files, printer, decoder)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
