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

package cmd

import (
	"github.com/ksonnet/ksonnet/actions"
	"github.com/pkg/errors"
)

type initName int

const (
	actionApply initName = iota
	actionInit
	actionValidate
)

type actionFn func(map[string]interface{}) error

var (
	actionFns = map[initName]actionFn{
		actionApply: actions.RunApply,
	}
)

func runAction(name initName, args map[string]interface{}) error {
	fn, ok := actionFns[name]
	if !ok {
		return errors.Errorf("invalid action")
	}

	return fn(args)
}

var (
	actionMap = map[initName]interface{}{
		actionInit:     actions.RunInit,
		actionValidate: actions.RunValidate,
	}
)
