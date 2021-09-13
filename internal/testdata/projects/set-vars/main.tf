data "qbec_jsonnet_eval" "template" {
  code = <<EOT
    {
      fooStr: std.extVar('fooStr'),
      barStr: std.extVar('barStr'),
      fooCode: std.extVar('fooCode'),
      barCode: std.extVar('barCode'),
    }
  EOT
  ext_str_vars = {
    fooStr = "hello"
    barStr = "world"
  }
  ext_code_vars = {
    fooCode = "true"
    barCode = "[ 'a', 'b']"
  }
}

output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}
