data "qbec_jsonnet_eval" "template" {
  file = "${path.module}/bad.jsonnet"
}

output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}
