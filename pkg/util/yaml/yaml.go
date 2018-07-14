// Copyright 2018 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package yaml

import (
	"bufio"
	"bytes"
	"io"
)

const (
	docSeparator = "---"
)

// Decode decodes YAML into one or more readers.
func Decode(f io.Reader) ([]io.Reader, error) {

	buffer := make([]bytes.Buffer, 1)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()
		if t == docSeparator {
			buffer = append(buffer, bytes.Buffer{})
			continue
		}

		if _, err := buffer[len(buffer)-1].WriteString(t); err != nil {
			return nil, err
		}

		if err := buffer[len(buffer)-1].WriteByte('\n'); err != nil {
			return nil, err
		}
	}

	var readers []io.Reader
	for i := range buffer {
		readers = append(readers, &buffer[i])
	}

	return readers, nil
}
