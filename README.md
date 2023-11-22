# chart-migrate

## Basic example invocations

### Kong chart

If coming from the `kong` chart, run only the root command. Keys whose names have changed will be moved to their new location.

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

### Ingress chart

If coming from the `ingress` chart, first run the root command and then the `merge` command on the output.

The root command will move settings from `ingress` that apply to the Deployment/Pod/etc. and will now live under
the `kong` chart's `ingressController` section to their appropriate `kong` key.

The `merge` command will move any remaining settings from `ingress` and `gateway` sections to the root of values.yaml.
If a key is present under both `ingress` and `gateway`, it will use the `gateway` value and print an alert. This should
not occur, as `ingress` keys that would collide with `gateway` keys should have all moved to new locations during the first step.

```
$ go run ./pkg/cmd -f /tmp/ingvalues.yaml -s ingress 2>/dev/null | tee /tmp/movvalues.yaml 

controller:
  deployment:
    kong:
      enabled: false
  enabled: true
  ingressController: {}
  proxy:
    nameOverride: '{{ .Release.Name }}-gateway-proxy'
gateway:
  admin:
    clusterIP: None
    enabled: true
    type: ClusterIP
  deployment:
    kong:
      enabled: true
  enabled: true
  env:
    database: "off"
    role: traditional
  ingressController:
    enabled: false
  podAnnotations:
    example.com/gateway: bongo
ingressController:
  deployment:
    annotations:
      example.com/example: whatever
    pod:
      annotations:
        kuma.io/gateway: enabled
        traffic.kuma.io/exclude-outbound-ports: "8444"
        traffic.sidecar.istio.io/excludeOutboundPorts: "8444"
      container:
        env:
          dump_config: true
        image:
          repository: traines/kic
      hostNework: true
      terminationGracePeriodSeconds: 20
      tmpDir:
        sizeLimit: 2Gi
  enabled: true
  gatewayDiscovery:
    enabled: true
    generateAdminApiService: true
```

```
$ go run ./pkg/cmd -f /tmp/movvalues.yaml -s ingress merge                                

admin:
  clusterIP: None
  enabled: true
  type: ClusterIP
deployment:
  kong:
    enabled: true
env:
  database: "off"
  role: traditional
ingressController:
  deployment:
    annotations:
      example.com/example: whatever
    pod:
      annotations:
        kuma.io/gateway: enabled
        traffic.kuma.io/exclude-outbound-ports: "8444"
        traffic.sidecar.istio.io/excludeOutboundPorts: "8444"
      container:
        env:
          dump_config: true
        image:
          repository: traines/kic
      hostNework: true
      terminationGracePeriodSeconds: 20
      tmpDir:
        sizeLimit: 2Gi
  enabled: true
  gatewayDiscovery:
    enabled: true
    generateAdminApiService: true
podAnnotations:
  example.com/gateway: bongo
```
