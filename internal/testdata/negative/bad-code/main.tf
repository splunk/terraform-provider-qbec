data "qbec_jsonnet_eval" "template" {
  code = "{{}}"
}

output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}
