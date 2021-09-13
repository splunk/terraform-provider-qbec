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
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/splunk/qbec/vm"
)

const (
	dsName         = "qbec_jsonnet_eval"
	fldFile        = "file"
	fldCode        = "code"
	fldLibPaths    = "lib_paths"
	fldExtStrVars  = "ext_str_vars"
	fldExtCodeVars = "ext_code_vars"
	fldTLAStrVars  = "tla_str_vars"
	fldTLACodeVars = "tla_code_vars"
	fldDataSources = "data_sources"
	fldRendered    = "rendered"
)

// Provider returns the provider schema.
func Provider() *schema.Provider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			dsName: evalResource(),
		},
	}
}

func evalResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			fldRendered: {
				Type:     schema.TypeString,
				Computed: true,
			},
			fldFile: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("jsonnet file to evaluate (one of %s or %s must be present)", fldFile, fldCode),
			},
			fldCode: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("inline jsonnet code to evaluate (one of %s or %s must be present)", fldFile, fldCode),
			},
			fldLibPaths: {
				Type:        schema.TypeList,
				Description: "library paths to use",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			fldExtStrVars: {
				Type:        schema.TypeMap,
				Description: "external variables to set as strings",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			fldExtCodeVars: {
				Type:        schema.TypeMap,
				Description: "external variables to set as code variables",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			fldTLAStrVars: {
				Type:        schema.TypeMap,
				Description: "TLA variables to set as strings",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			fldTLACodeVars: {
				Type:        schema.TypeMap,
				Description: "TLA variables to set as code variables",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			fldDataSources: {
				Type:        schema.TypeList,
				Description: "data source URIs to configure",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
		ReadContext: evalJsonnet,
	}
}

func evalJsonnet(_ context.Context, data *schema.ResourceData, _ interface{}) (res diag.Diagnostics) {
	file, code, err := getCode(data)
	if err != nil {
		return diag.FromErr(err)
	}
	libPaths, err := getStringArray(fldLibPaths, data)
	if err != nil {
		return diag.FromErr(err)
	}
	vars, err := getVariables(data)
	if err != nil {
		return diag.FromErr(err)
	}
	dsURIs, err := getStringArray(fldDataSources, data)
	if err != nil {
		return diag.FromErr(err)
	}
	dataSources, closer, err := vm.CreateDataSources(dsURIs, vm.ConfigProviderFromVariables(vars))
	defer func() {
		if closer == nil {
			return
		}
		err := closer.Close()
		if err != nil {
			res = append(res, diag.FromErr(err)...)
		}
	}()
	if err != nil {
		return diag.FromErr(err)
	}
	jvm := vm.New(vm.Config{
		LibPaths:    libPaths,
		DataSources: dataSources,
	})
	var output string
	if file != "" {
		output, err = jvm.EvalFile(file, vars)
	} else {
		output, err = jvm.EvalCode("<inline-code>", vm.MakeCode(code), vars)
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if err := data.Set(fldRendered, output); err != nil {
		return diag.FromErr(err)
	}
	data.SetId(outputHash(output))
	return
}

func getCode(data *schema.ResourceData) (file, code string, err error) {
	file0, fileFound := data.GetOk("file")
	code0, codeFound := data.GetOk("code")

	switch {
	case !fileFound && !codeFound:
		return "", "", fmt.Errorf("one of '%s' or '%s' attributes must be set", fldCode, fldFile)
	case fileFound && codeFound:
		return "", "", fmt.Errorf("cannot set both '%s' and '%s'", fldCode, fldFile)
	case fileFound:
		file = file0.(string)
	default:
		code = code0.(string)
		if strings.Trim(code, " \t\n\r") == "" {
			return "", "", fmt.Errorf("%s string is empty", fldCode)
		}
	}
	return file, code, nil
}

func getStringArray(name string, data *schema.ResourceData) ([]string, error) {
	var ret []string
	in0, ok := data.GetOk(name)
	if ok {
		in1, ok := in0.([]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected type for %s, want []interface{} got %v", name, reflect.TypeOf(in0))
		}
		for _, s := range in1 {
			ret = append(ret, fmt.Sprint(s))
		}
	}
	return ret, nil
}

func getVariables(data *schema.ResourceData) (res vm.VariableSet, _ error) {
	var extVars []vm.Var
	var tlaVars []vm.Var

	add := func(fldName string, fn func(name, value string) vm.Var, tla bool) error {
		vars0, ok := data.GetOk(fldName)
		if !ok {
			return nil
		}
		vars, ok := vars0.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected type for %s, want map[string]interface{} got %v", fldName, reflect.TypeOf(vars0))
		}
		for k, v := range vars {
			val := fmt.Sprint(v)
			userVar := fn(k, val)
			if tla {
				tlaVars = append(tlaVars, userVar)
			} else {
				extVars = append(extVars, userVar)
			}
		}
		return nil
	}

	err := add(fldExtStrVars, vm.NewVar, false)
	if err != nil {
		return res, err
	}
	err = add(fldExtCodeVars, vm.NewCodeVar, false)
	if err != nil {
		return res, err
	}
	err = add(fldTLAStrVars, vm.NewVar, true)
	if err != nil {
		return res, err
	}
	err = add(fldTLACodeVars, vm.NewCodeVar, true)
	if err != nil {
		return res, err
	}
	return vm.VariableSet{}.WithVars(extVars...).WithTopLevelVars(tlaVars...), nil
}

func outputHash(s string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(s))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
