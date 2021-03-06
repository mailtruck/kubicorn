// Copyright © 2017 The Kubicorn Authors
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

package local

import (
	"os"
	"os/user"
	"strings"

	"github.com/kubicorn/kubicorn/pkg/logger"
)

const (
	TestHome string = "KUBICORN_TEST_HOME_DIRECTORY"
)

func Home() string {
	if os.Getenv(TestHome) != "" {
		return os.Getenv(TestHome)
	}

	home := os.Getenv("HOME")
	if strings.Contains(home, "root") {
		return "/root"
	}
	usr, err := user.Current()
	if err != nil {
		logger.Warning("unable to find user: %v", err)
		return ""
	}
	return usr.HomeDir
}

func Expand(path string) string {
	if strings.Contains(path, "~") {
		return strings.Replace(path, "~", Home(), 1)
	}
	return path
}
