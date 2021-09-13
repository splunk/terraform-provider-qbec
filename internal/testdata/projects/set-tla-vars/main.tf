data "qbec_jsonnet_eval" "template" {
  code = <<EOT
    function (fooStr, fooCode, barStr, barCode) {
      fooStr: fooStr,
      barStr: barStr,
      fooCode: fooCode,
      barCode: barCode,
    }
  EOT
  tla_str_vars = {
    fooStr = "hello"
    barStr = "world"
  }
  tla_code_vars = {
    fooCode = "true"
    barCode = "[ 'a', 'b']"
  }
}

output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}
