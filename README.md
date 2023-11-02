# chart-migrate

Basic example invocations:

```
go run ./pkg/cmd -f /tmp/mutate_values.yaml
go run ./pkg/cmd -f /tmp/mutate_values.yaml --output-format=json
go run ./pkg/cmd -f /tmp/mutate_values.yaml --output-format=json| jq .deployment,.podAnnotations
```

```
17:22:40-0700 esenin $ cat /tmp/mutate_values.yaml | python -c 'import json, sys, yaml ; y=yaml.safe_load(sys.stdin.read()) ; print(json.dumps(y))' | jq .deployment,.podAnnotations
{
  "kong": {
    "enabled": true
  },
  "serviceAccount": {
    "create": true,
    "automountServiceAccountToken": false
  },
  "test": {
    "enabled": false
  },
  "daemonset": false,
  "hostNetwork": false,
  "prefixDir": {
    "sizeLimit": "256Mi"
  },
  "tmpDir": {
    "sizeLimit": "1Gi"
  }
}
{
  "kuma.io/gateway": "enabled",
  "traffic.sidecar.istio.io/includeInboundPorts": ""
}
17:22:47-0700 esenin $ go run ./pkg/cmd -f /tmp/mutate_values.yaml --output-format=json | jq .deployment,.podAnnotations                                                                                       
{
  "daemonset": false,
  "hostNetwork": false,
  "kong": {
    "enabled": true,
    "pod": {
      "annotations": {
        "kuma.io/gateway": "enabled",
        "traffic.sidecar.istio.io/includeInboundPorts": ""
      }
    },
    "annotations": {}
  },
  "prefixDir": {
    "sizeLimit": "256Mi"
  },
  "serviceAccount": {
    "automountServiceAccountToken": false,
    "create": true
  },
  "test": {
    "enabled": false
  },
  "tmpDir": {
    "sizeLimit": "1Gi"
  },
  "controller": {
    "pod": {
      "container": {
        "env": {
          "kong_admin_tls_skip_verify": true
        },
        "image": {
          "effectiveSemver": null,
          "repository": "kong/kubernetes-ingress-controller",
          "tag": "2.12"
        }
      }
    }
  }
}
null
```
