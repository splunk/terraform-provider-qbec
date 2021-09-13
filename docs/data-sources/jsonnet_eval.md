# qbec_jsonnet_eval (data source)

This data source evaluates a file or inline code using a jsonnet VM as configured in [qbec](https://qbec.io/).
Read the provider docs for more information on how this VM differs from a standard jsonnet VM.

### Example usage
A simple example of evaluating inline code:

```hcl
// evaluate inline code passing in some variables
data "qbec_jsonnet_eval" "template" {
  code = <<EOT
    {
      foo: std.extVar('foo'),
      bar: std.extVar('bar'),
    }
  EOT
  ext_str_vars = {
    foo = "hello"
    bar = "world"
  }
}

// result contains the json object, 
// in this case { foo: 'hello', bar: 'world' }
output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}
```

More complex examples can be found under internal/testdata/projects in the source tree.

### Required attributes

Exactly one of these two attributes must be set.

- **file** (String) jsonnet file to evaluate
- **code** (String) inline jsonnet code to evaluate

### Optional attributes

- **lib_paths** (List of String) library paths to use
- **ext_str_vars** (Map of String) external variables to set as strings
- **ext_code_vars** (Map of String) external variables to set as code variables
- **tla_str_vars** (Map of String) TLA variables to set as strings
- **tla_code_vars** (Map of String) TLA variables to set as code variables
- **data_sources** (List of String) Data source URIs to configure for the evaluation. 
  See the [qbec docs](https://qbec.io/reference/jsonnet-external-data/) for more details.

### Computed attributes

- **rendered** (String) the output of evaluation as a JSON string. Note that the jsonnet code itself could return a string;
  in that case the rendered attribute is a json-quoted string, and you need to call `jsondecode` on the result 
  to extract its raw value.

-> Paths to the main jsonnet file to evaluate as well as library paths should be specified as absolute paths constructed
using `${path.module}`. There is no way to reliably specify relative paths.
