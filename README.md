# chart-migrate

## Installation

Release archives are available from the [chart-migrate releases
page](https://github.com/Kong/chart-migrate/releases) for common platforms and
architectures. `chart-migrate` is a self-contained application: you can extract
the archive and execute the `chart-migrate` executable without any additional
installation steps.

## Basic example invocations

### Kong chart

If coming from the `kong` chart, run only the `migrate` command. Keys whose
names have changed will be moved to their new location. For example:

```
./chart-migrate migrate -f /tmp/values.yaml > /tmp/migrated.yaml
```

If you inspect the original and output values, you can see that migrated fields
are now present at new locations:

```
$ cat /tmp/values.yaml | python -c 'import json, sys, yaml ; y=yaml.safe_load(sys.stdin.read()) ; print(json.dumps(y))' | jq .ingressController.image
{
  "repository": "kong/kubernetes-ingress-controller",
  "tag": "3.0",
  "effectiveSemver": null
}


$ ./chart-migrate migrate -f /tmp/values.yaml --output-format=json | jq .ingressController.image,.ingressController.deployment.pod.container.image
null
{
  "effectiveSemver": null,
  "repository": "kong/kubernetes-ingress-controller",
  "tag": "3.0"
}
```

### Ingress chart

If coming from the `ingress` chart, first run the `migrate` command with `-s
ingress` and then the `merge` command on the output.

```
./chart-migrate migrate -f /tmp/values.yaml -s ingress > /tmp/migrated.yaml
./chart-migrate merge -f /tmp/migrated.yaml > /tmp/merged.yaml
```

The `migrate` command will move settings from `controller` that apply to the
Deployment/Pod/etc. to the appropriate `ingressController` subsection.

The `merge` command will move all remaining sections from under the
`controller` and `gateway` sections into their equivalent root-level settings.

If a key is present under both `ingress` and `gateway`, it will use the
`gateway` value and print an alert. This should not occur, as `ingress` keys
that would collide with `gateway` keys should have all moved to new locations
during the first step.

For example, the initial `migrate` command  will create a root-level
`ingressController` section with the migrated controller keys:

```
$ cat /tmp/values.yaml | python -c 'import json, sys, yaml ; y=yaml.safe_load(sys.stdin.read()) ; print(json.dumps(y))' | jq .ingressController,.controller.podAnnotations
null
{
  "kuma.io/gateway": "enabled",
  "traffic.kuma.io/exclude-outbound-ports": "8444",
  "traffic.sidecar.istio.io/excludeOutboundPorts": "8444"
}

$ ./chart-migrate migrate -f /tmp/values.yaml --output-format=json -s ingress | jq .ingressController,.ingress.podAnnotations
{
  "enabled": true,
  "gatewayDiscovery": {
    "enabled": true,
    "generateAdminApiService": true
  },
  "deployment": {
    "pod": {
      "annotations": {
        "kuma.io/gateway": "enabled",
        "traffic.kuma.io/exclude-outbound-ports": "8444",
        "traffic.sidecar.istio.io/excludeOutboundPorts": "8444"
      }
    }
  }
}
null
```

`migrate` alone will leave most keys under `gateway` and `controller` at their
original location. For example, the `env` key will still be under `gateway`.
Running `merge` moves these keys to their root-level sections:

```
$ ./chart-migrate migrate -f /tmp/values.yaml -s ingress > /tmp/migrated.yaml

$ ./chart-migrate migrate -f /tmp/values.yaml -s ingress --output-format=json | jq .gateway.env,.env
{
  "database": "off",
  "role": "traditional"
}
null

$ ./chart-migrate merge -f /tmp/migrated.yaml | python -c 'import json, sys, yaml ; y=yaml.safe_load(sys.stdin.read()) ; print(json.dumps(y))' | jq .gateway.env,.env
null
{
  "database": "off",
  "role": "traditional"
}
```
