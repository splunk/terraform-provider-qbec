
data "qbec_jsonnet_eval" "template" {
  code = <<EOT
    importstr 'data://echo'
  EOT
  data_sources = [
    "exec://echo?configVar=echoConfig"
  ]
}

output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}
