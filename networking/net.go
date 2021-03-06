// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networking

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"

	"github.com/coreos/rocket/networking/util"
	rktpath "github.com/coreos/rocket/path"
)

// Net encodes a network plugin.
type Net struct {
	util.Net
	args string
}

// Absolute path where users place their net configs
const UserNetPath = "/etc/rkt/net.d"

// Default net path relative to stage1 root
const DefaultNetPath = "etc/rkt/net.d/99-default.conf"

func listFiles(dir string) ([]string, error) {
	dirents, err := ioutil.ReadDir(dir)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return nil, nil
	default:
		return nil, err
	}

	files := []string{}
	for _, dent := range dirents {
		if dent.IsDir() {
			continue
		}

		files = append(files, dent.Name())
	}

	return files, nil
}

func loadUserNets() ([]Net, error) {
	files, err := listFiles(UserNetPath)
	if err != nil {
		return nil, err
	}

	sort.Strings(files)

	nets := make([]Net, 0, len(files))

	for _, filename := range files {
		filepath := path.Join(UserNetPath, filename)
		n := Net{}
		if err := util.LoadNet(filepath, &n); err != nil {
			return nil, fmt.Errorf("error loading %v: %v", filepath, err)
		}

		nets = append(nets, n)
	}

	return nets, nil
}

// Loads nets specified by user and default one from stage1
func (e *containerEnv) loadNets() ([]Net, error) {
	nets, err := loadUserNets()
	if err != nil {
		return nil, err
	}

	defPath := path.Join(rktpath.Stage1RootfsPath(e.rktRoot), DefaultNetPath)
	defNet := Net{}
	if err := util.LoadNet(defPath, &defNet); err != nil {
		return nil, fmt.Errorf("error loading net: %v", err)
	}

	return append(nets, defNet), nil
}
