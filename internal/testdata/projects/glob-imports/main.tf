data "qbec_jsonnet_eval" "template" {
  file = "${path.module}/load.jsonnet"
}

output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}
