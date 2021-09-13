data "qbec_jsonnet_eval" "template" {
}

output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}
