locals {
  echoConfig = {
    command = "echo"
    args = [ "hello", "world" ]
  }
}

data "qbec_jsonnet_eval" "template" {
  code = <<EOT
    importstr 'data://echo'
  EOT
  data_sources = [
    "exec://echo?configVar=echoConfig"
  ]
  ext_code_vars = {
    echoConfig = jsonencode(local.echoConfig)
  }
}

output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}
