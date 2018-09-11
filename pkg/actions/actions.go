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
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/ksonnet/ksonnet/pkg/upgrade"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// OptionApp is app option.
	OptionApp = "app"
	// OptionAppRoot is the root directory of the application.
	OptionAppRoot = "app-root"
	// OptionArguments is arguments option. Used for passing arguments to prototypes.
	OptionArguments = "arguments"
	// OptionAsString is asString. Used for setting values as strings.
	OptionAsString = "as-string"
	// OptionClientConfig is clientConfig option.
	OptionClientConfig = "client-config"
	// OptionComponentName is a componentName option.
	OptionComponentName = "component-name"
	// OptionComponentNames is componentNames option.
	OptionComponentNames = "component-names"
	// OptionCreate is create option.
	OptionCreate = "create"
	// OptionDryRun is dryRun option.
	OptionDryRun = "dry-run"
	// OptionEnvName is envName option.
	OptionEnvName = "env-name"
	// OptionEnvName1 is envName1. Used for param diff.
	OptionEnvName1 = "env-name-1"
	// OptionEnvName2 is envName1. Used for param diff.
	OptionEnvName2 = "env-name-2"
	// OptionExtVarFiles is jsonnet ext var files.
	OptionExtVarFiles = "ext-vars-files"
	// OptionExtVars is jsonnet ext vars.
	OptionExtVars = "ext-vars"
	// OptionForce is force option.
	OptionForce = "force"
	// OptionFormat is format option.
	OptionFormat = "format"
	// OptionFs is fs option.
	OptionFs = "fs"
	// OptionGcTag is gcTag option.
	OptionGcTag = "gc-tag"
	// OptionGlobal is global option.
	OptionGlobal = "global"
	// OptionGracePeriod is gracePeriod option.
	OptionGracePeriod = "grace-period"
	// OptionHTTPClient is the http.Client for outbound network requests.
	OptionHTTPClient = "http-client"
	// OptionInstalled is for listing installed packages.
	OptionInstalled = "only-installed"
	// OptionJPaths is jsonnet paths.
	OptionJPaths = "jpaths"
	// OptionPkgName is (an optionally qualified) name of a package.
	OptionPkgName = "pkg-name"
	// OptionName is name option.
	OptionName = "name"
	// OptionModule is component module option.
	OptionModule = "module"
	// OptionNamespace is a cluster namespace option
	OptionNamespace = "namespace"
	// OptionNewRoot is init new root path option.
	OptionNewRoot = "root-path"
	// OptionNewEnvName is newEnvName option. Used for renaming environments.
	OptionNewEnvName = "new-env-name"
	// OptionOutput is output option.
	OptionOutput = "output"
	// OptionOverride is override option.
	OptionOverride = "override"
	// OptionPackageName is packageName option.
	OptionPackageName = "package-name"
	// OptionPath is path option.
	OptionPath = "path"
	// OptionQuery is query option.
	OptionQuery = "query"
	// OptionResolveImage is resolve image option. It is used to resolve docker image references
	// when setting parameters.
	OptionResolveImage = "resolve-image"
	// OptionServer is server option.
	OptionServer = "server"
	// OptionServerURI is serverURI option.
	OptionServerURI = "server-uri"
	// OptionSkipCheckUpgrade tells app not to emit upgrade warnings, probably because the user is already upgrading.
	OptionSkipCheckUpgrade = "skip-check-upgrade"
	// OptionSkipDefaultRegistries is skipDefaultRegistries option. Used by init.
	OptionSkipDefaultRegistries = "skip-default-registries"
	// OptionSkipGc is skipGc option.
	OptionSkipGc = "skip-gc"
	// OptionSpecFlag is specFlag option. Used for setting k8s spec.
	OptionSpecFlag = "spec-flag"
	// OptionSrc1 is src1 option.
	OptionSrc1 = "src-1"
	// OptionSrc2 is src2 option.
	OptionSrc2 = "src-2"
	// OptionTlaVarFiles is jsonnet tla var files.
	OptionTlaVarFiles = "tla-var-files"
	// OptionTlaVars is jsonnet tla vars.
	OptionTlaVars = "tla-vars"
	// OptionTLSSkipVerify specifies that tls server certifactes should not be verified.
	OptionTLSSkipVerify = "tls-skip-verify"
	// OptionUnset is unset option.
	OptionUnset = "unset"
	// OptionURI is uri option. Used for setting registry URI.
	OptionURI = "URI"
	// OptionWithoutModules is without modules option.
	OptionWithoutModules = "without-modules"
	// OptionValue is value option.
	OptionValue = "value"
	// OptionVersion is version option.
	OptionVersion = "version"
)

const (
	// OutputWide is wide output
	OutputWide = "wide"
	// OutputJSON is JSON output
	OutputJSON = "json"
)

var (
	// ErrNotInApp is an error stating the user is not in a ksonnet application directory
	// hierarchy.
	ErrNotInApp = errors.Errorf("this command has to be run within a ksonnet application")
)

type missingOptionError struct {
	name string
}

func newMissingOptionError(name string) *missingOptionError {
	return &missingOptionError{
		name: name,
	}
}

func (e *missingOptionError) Error() string {
	return fmt.Sprintf("missing required %s option", e.name)
}

type invalidOptionError struct {
	name string
}

func newInvalidOptionError(name string) *invalidOptionError {
	return &invalidOptionError{
		name: name,
	}
}

func (e *invalidOptionError) Error() string {
	return fmt.Sprintf("invalid type for option %s", e.name)
}

// optionLoader loads typed option from a configuration map. If an error
// occurs in any of the load processes, the loader will return default
// values for type.
type optionLoader struct {
	// err is the error state for the optionLoader. If this is not nil, all
	// subsequent calls to load will return nil.
	err error
	m   map[string]interface{}
}

func newOptionLoader(m map[string]interface{}) *optionLoader {
	return &optionLoader{
		m: m,
	}
}

func (o *optionLoader) LoadFs() afero.Fs {
	i := o.loadOptional(OptionFs)
	if i == nil {
		return afero.NewOsFs()
	}

	a, ok := i.(afero.Fs)
	if !ok {
		o.err = newInvalidOptionError(OptionFs)
		return nil
	}

	return a
}

func (o *optionLoader) LoadBool(name string) bool {
	i := o.load(name)
	if i == nil {
		return false
	}

	a, ok := i.(bool)
	if !ok {
		o.err = newInvalidOptionError(name)
		return false
	}

	return a
}

func (o *optionLoader) LoadOptionalBool(name string) bool {
	i := o.loadOptional(name)
	if i == nil {
		return false
	}

	a, ok := i.(bool)
	if !ok {
		return false
	}

	return a
}

func (o *optionLoader) LoadInt(name string) int {
	i := o.load(name)
	if i == nil {
		return 0
	}

	a, ok := i.(int)
	if !ok {
		o.err = newInvalidOptionError(name)
		return 0
	}

	return a
}

func (o *optionLoader) LoadInt64(name string) int64 {
	i := o.load(name)
	if i == nil {
		return 0
	}

	a, ok := i.(int64)
	if !ok {
		o.err = newInvalidOptionError(name)
		return 0
	}

	return a
}

func (o *optionLoader) LoadOptionalInt(name string) int {
	i := o.loadOptional(name)
	if i == nil {
		return 0
	}

	a, ok := i.(int)
	if !ok {
		return 0
	}

	return a
}

func (o *optionLoader) LoadString(name string) string {
	i := o.load(name)
	if i == nil {
		return ""
	}

	a, ok := i.(string)
	if !ok {
		o.err = newInvalidOptionError(name)
		return ""
	}

	return a
}

func (o *optionLoader) LoadOptionalString(name string) string {
	i := o.loadOptional(name)
	if i == nil {
		return ""
	}

	a, ok := i.(string)
	if !ok {
		return ""
	}

	return a
}

func (o *optionLoader) LoadStringSlice(name string) []string {
	i := o.load(name)
	if i == nil {
		return nil
	}

	a, ok := i.([]string)
	if !ok {
		o.err = newInvalidOptionError(name)
		return nil
	}

	return a
}

func (o *optionLoader) LoadClientConfig() *client.Config {
	i := o.load(OptionClientConfig)
	if i == nil {
		return nil
	}

	a, ok := i.(*client.Config)
	if !ok {
		o.err = newInvalidOptionError(OptionClientConfig)
		return nil
	}

	return a
}

// LoadApp returns an app.App reference - either as passed via OptionApp,
// or newly constructed.
func (o *optionLoader) LoadApp() app.App {
	i := o.loadOptional(OptionApp)
	a, ok := i.(app.App)
	if i != nil && !ok {
		// App was provided but was invalid type
		o.err = newInvalidOptionError(OptionApp)
		return nil
	}
	if a != nil {
		// Return app if a valid app.App was provided
		return a
	}

	var fs = o.LoadFs()
	if fs == nil {
		o.err = errors.New("missing required fs reference")
		return nil
	}
	var httpClient = o.LoadHTTPClient()
	if httpClient == nil {
		o.err = errors.New("initializing http client")
		return nil
	}
	var appRoot = o.LoadOptionalString(OptionAppRoot)

	appRoot, err := app.FindRoot(fs, appRoot)
	if err != nil {
		o.err = errors.Wrapf(err, "finding app root from starting path: %s", appRoot)
		return nil
	}

	a, err = app.Load(fs, httpClient, appRoot)
	if err != nil {
		o.err = errors.New("initializing app")
		return nil
	}

	if !o.LoadOptionalBool(OptionSkipCheckUpgrade) {
		pm := registry.NewPackageManager(a)
		if _, err := upgrade.CheckUpgrade(a, new(bytes.Buffer), pm, false); err != nil {
			o.err = errors.Wrap(err, "checking for app upgrades")
			return nil
		}
	}

	return a
}

// LoadHTTPClient loads an HTTP client based on common configuration for certificates, tls verification, timeouts, etc.
func (o *optionLoader) LoadHTTPClient() *http.Client {
	i := o.loadOptional(OptionHTTPClient)
	if c, ok := i.(*http.Client); ok {
		return c
	}

	// Construct a client if none was passed
	tlsSkipVerify := o.LoadOptionalBool(OptionTLSSkipVerify)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: tlsSkipVerify,
	}

	timeoutSeconds := 10

	var defaultTransport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSClientConfig:       tlsConfig,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	c := &http.Client{
		Timeout:   time.Duration(timeoutSeconds) * time.Second,
		Transport: defaultTransport,
	}

	return c
}

func (o *optionLoader) load(key string) interface{} {
	if o.err != nil {
		return nil
	}

	i, ok := o.m[key]
	if !ok {
		o.err = newMissingOptionError(key)
	}

	return i
}

func (o *optionLoader) loadOptional(key string) interface{} {
	if o.err != nil {
		return nil
	}

	i, ok := o.m[key]
	if !ok {
		return nil
	}

	return i
}
