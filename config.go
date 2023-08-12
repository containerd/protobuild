/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"io/ioutil"
	"log"

	"github.com/pelletier/go-toml"
)

type config struct {
	Version    string
	Generators []string

	// Parameters are custom parameters to be passed to the generators.
	// The parameter key must be the generator name with a table value
	// of keys and string values to be passed.
	// Example:
	// [parameters.go-ttrpc]
	// customkey = "somevalue"
	Parameters map[string]map[string]string

	Includes struct {
		Before   []string
		Vendored []string
		Packages []string
		After    []string
	}

	Packages map[string]string

	Overrides []struct {
		Prefixes []string
		Generators []string
		Parameters map[string]map[string]string

		// TODO(stevvooe): We could probably support overriding of includes and
		// package maps, but they don't seem to be as useful. Likely,
		// overriding the package map is more useful but includes happen
		// project-wide.
	}

	Descriptors []struct {
		Prefix      string
		Target      string
		IgnoreFiles []string `toml:"ignore_files"`
	}
}

func newDefaultConfig() config {
	return config{
		Includes: struct {
			Before   []string
			Vendored []string
			Packages []string
			After    []string
		}{
			Before: []string{"."},
			After:  []string{"/usr/local/include", "/usr/include"},
		},
	}
}

func readConfig(path string) (config, error) {
	p, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	return readConfigFrom(p)
}

func readConfigFrom(p []byte) (config, error) {
	c := newDefaultConfig()
	if err := toml.Unmarshal(p, &c); err != nil {
		log.Fatalln(err)
	}

	if len(c.Generators) == 0 {
		c.Generators = []string{"go"}
	}

	return c, nil
}
