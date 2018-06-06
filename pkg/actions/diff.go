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

package actions

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/ksonnet/ksonnet/pkg/diff"
	"github.com/pkg/errors"
)

var (
	// ErrDiffFound is an error returned when differences are found.
	ErrDiffFound = errors.New("differences found")

	diffAddColor    = color.New(color.FgGreen)
	diffRemoveColor = color.New(color.FgRed)
)

// RunDiff runs `diff`
func RunDiff(m map[string]interface{}) error {
	d, err := NewDiff(m)
	if err != nil {
		return err
	}

	return d.Run()
}

// Diff sets targets for an environment.
type Diff struct {
	app          app.App
	clientConfig *client.Config
	src1         string
	src2         string
	components   []string

	diffFn func(app.App, *client.Config, []string, *diff.Location, *diff.Location) (io.Reader, error)

	out io.Writer
}

// NewDiff creates an instance of Diff.
func NewDiff(m map[string]interface{}) (*Diff, error) {
	ol := newOptionLoader(m)

	d := &Diff{
		app:          ol.LoadApp(),
		clientConfig: ol.LoadClientConfig(),
		src1:         ol.LoadString(OptionSrc1),
		src2:         ol.LoadOptionalString(OptionSrc2),
		components:   ol.LoadStringSlice(OptionComponentNames),

		diffFn: diff.DefaultDiff,

		out: os.Stdout,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return d, nil
}

// Run assigns targets to an environment.
func (d *Diff) Run() error {
	location1 := diff.NewLocation(d.src1)

	if d.src2 == "" {
		d.src2 = fmt.Sprintf("%s:%s", "remote", location1.EnvName())
	}
	location2 := diff.NewLocation(d.src2)

	r, err := d.diffFn(d.app, d.clientConfig, d.components, location1, location2)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		t := scanner.Text()

		switch {
		case strings.HasPrefix(t, "+"):
			_, err = diffAddColor.Fprintln(&buf, t)
			if err != nil {
				return err
			}
		case strings.HasPrefix(t, "-"):
			diffRemoveColor.Fprintln(&buf, t)
			if err != nil {
				return err
			}
		default:
			fmt.Fprintln(&buf, t)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if s := buf.String(); s != "" {
		fmt.Fprintln(d.out, s)
		return ErrDiffFound
	}

	return nil
}
