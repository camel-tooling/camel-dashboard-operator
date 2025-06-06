/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

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
	"os"
	"strings"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/util"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, `Use "license-check <file> <license>`)
		os.Exit(1)
	}

	fileName := os.Args[1]
	licenseName := os.Args[2]

	fileBin, err := util.ReadFile(fileName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot read file %s: %v\n", fileName, err)
		os.Exit(1)
	}
	file := string(fileBin)

	licenseBin, err := util.ReadFile(licenseName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cannot read file %s: %v\n", licenseName, err)
		os.Exit(1)
	}
	license := string(licenseBin)

	if !strings.Contains(file, license) {
		_, _ = fmt.Fprintf(os.Stderr, "file %s does not contain license\n", fileName)
		os.Exit(1)
	}
}
