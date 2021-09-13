/*
   Copyright 2021 Splunk Inc.

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

package internal

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/splunk/qbec/vm"
	"github.com/stretchr/testify/require"
)

func TestProvider(t *testing.T) {
	err := Provider().InternalValidate()
	require.NoError(t, err)
}

// setWD sets the current working dir to the one supplied and returns a function that restores the old value.
func setWD(t *testing.T, dir string) func() {
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	return func() {
		require.NoError(t, os.Chdir(wd))
	}
}

// readTF reads the main.tf file from the current directory
func readTF(t *testing.T) string {
	b, err := ioutil.ReadFile("main.tf")
	require.NoError(t, err)
	return string(b)
}

// readResult reads the expected output from a result.jsonnet file in the current directory
func readResult(t *testing.T) interface{} {
	b, err := ioutil.ReadFile("result.jsonnet")
	require.NoError(t, err)
	jvm := vm.New(vm.Config{})
	out, err := jvm.EvalCode("result-file", vm.MakeCode(string(b)), vm.VariableSet{})
	require.NoError(t, err)
	var res interface{}
	require.NoError(t, json.Unmarshal([]byte(out), &res))
	return res
}

// readError reads the expected error message from an error.txt file in the current directory
func readError(t *testing.T) string {
	b, err := ioutil.ReadFile("error.txt")
	require.NoError(t, err)
	return strings.Trim(string(b), "\r\n")
}

func TestDataSource(t *testing.T) {
	dirs, err := filepath.Glob("testdata/projects/*")
	require.NoError(t, err)
	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			defer setWD(t, dir)()
			config := readTF(t)
			result := readResult(t)
			resource.UnitTest(t, resource.TestCase{
				IsUnitTest: true,
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"qbec": func() (*schema.Provider, error) { return Provider(), nil },
				},
				Steps: []resource.TestStep{
					{
						Config: config,
						Check: func(state *terraform.State) error {
							outState := state.RootModule().Outputs["result"]
							require.NotNil(t, outState)
							require.EqualValues(t, result, outState.Value)
							return nil
						},
					},
				},
			})
		})
	}
}

func TestDataSourceNegative(t *testing.T) {
	dirs, err := filepath.Glob("testdata/negative/*")
	require.NoError(t, err)
	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			defer setWD(t, dir)()
			config := readTF(t)
			errmsg := readError(t)
			re, err := regexp.Compile(regexp.QuoteMeta(errmsg))
			require.NoError(t, err)
			resource.UnitTest(t, resource.TestCase{
				IsUnitTest: true,
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"qbec": func() (*schema.Provider, error) { return Provider(), nil },
				},
				Steps: []resource.TestStep{
					{
						Config:      config,
						ExpectError: re,
					},
				},
			})
		})
	}
}
