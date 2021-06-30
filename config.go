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
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

const configVersion = "unstable"

type config struct {
	Version   string
	Generator string
	Plugins   []string
	Includes  struct {
		Before   []string
		Vendored []string
		Packages []string
		After    []string
	}

	Packages map[string]string

	Overrides []struct {
		Prefixes  []string
		Generator string
		Plugins   *[]string

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
		Generator: "go",
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

func readConfig(path string) (config, string, error) {
	configFile, err := filepath.Abs(path)
	if err != nil {
		log.Fatalln(err)
	}

	p, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalln(err)
	}
	c := newDefaultConfig()
	if err := toml.Unmarshal(p, &c); err != nil {
		log.Fatalln(err)
	}

	if c.Version != configVersion {
		return config{}, "", fmt.Errorf("unknown file version %v; please upgrade to %v", c.Version, configVersion)
	}

	return c, filepath.Dir(configFile), nil
}
