data "qbec_jsonnet_eval" "template" {
  code      = <<EOT
    {
      foo: 'foo',
      bar: 'bar',
      lib: import 'foobar.libsonnet',
    }
  EOT
  lib_paths = ["${path.module}/../../lib"]
}

output "result" {
  value = jsondecode(data.qbec_jsonnet_eval.template.rendered)
}
