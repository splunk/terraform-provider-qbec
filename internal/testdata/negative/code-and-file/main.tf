data "qbec_jsonnet_eval" "template" {
  file      = "${path.module}/simple.jsonnet"
  code      = "{}"
  lib_paths = ["${path.module}/../../lib"]
}

output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}
