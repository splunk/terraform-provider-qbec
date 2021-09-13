data "qbec_jsonnet_eval" "template" {
  code = <<EOT
  "hello world"
  EOT
}

output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}

