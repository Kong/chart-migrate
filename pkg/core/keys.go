package core

// TODO this is a basic single set, random order collection of migrations, since there's a readily available type for
// that. should probably convert to a struct we can stick in an ordered array and then allow running different sets
// in the order specified, to handle migrations that depend on one another.

// TODO there are some keys that only migrate if you're coming from the ingress chart because they were using existing
// global settings to modify the controller deployment. AFAIK podAnnotations and proxy are the only two of these. proxy
// shouldn't matter since we can infer the correct setting in the split kong chart automatically.
// controller.podAnnotations would move to ingressController.deployment.pod.annotations, but building a second set of
// conditional maps complicates things. We may just ignore this and use the default annotations in kong 3.x to handle
// most users and ask the remainder to migrate manually

// TODO the kong chart is sort of affected by a concern similar to the above: annotations previously applied to the
// single Deployment via podAnnotations and the like _may_ be relevant for the new controller Deployment, but there's
// no way to know whether they were in place for the controller or for the proxy

// getKeyReMaps returns a map of strings to strings. Keys are the original locations of a key in values.yaml and values
// are their new locations. Both are in dotted string format: "foo.bar.baz" indicates a YAML structure like:
// foo:
//
//	bar:
//	  baz: {}

func getControllerKeys() map[string]string {
	return map[string]string{
		"ingressController.image":          "ingressController.deployment.pod.container.image",
		"ingressController.args":           "ingressController.deployment.pod.container.args",
		"ingressController.env":            "ingressController.deployment.pod.container.env",
		"ingressController.customEnv":      "ingressController.deployment.pod.container.customEnv",
		"ingressController.livenessProbe":  "ingressController.deployment.pod.container.livenessProbe",
		"ingressController.readinessProbe": "ingressController.deployment.pod.container.readinessProbe",
		"ingressController.resources":      "ingressController.deployment.pod.container.resources",
	}
}

func getGatewayKeys() map[string]string {
	return map[string]string{}
}

// TODO weird keys
// proxy.* probably no longer necessary since it's only required to deal with the subchart boundary

func getIngressControllerKeys() map[string]string {
	return map[string]string{
		// moved keys, basically getControllerKeys() with a controller. prefix
		"controller.ingressController.image":          "ingressController.deployment.pod.container.image",
		"controller.ingressController.args":           "ingressController.deployment.pod.container.args",
		"controller.ingressController.env":            "ingressController.deployment.pod.container.env",
		"controller.ingressController.customEnv":      "ingressController.deployment.pod.container.customEnv",
		"controller.ingressController.livenessProbe":  "ingressController.deployment.pod.container.livenessProbe",
		"controller.ingressController.readinessProbe": "ingressController.deployment.pod.container.readinessProbe",
		"controller.ingressController.resources":      "ingressController.deployment.pod.container.resources",
		// remapped keys, i.e. formerly global keys that can apply to the controller
		"controller.deployment.serviceAccount":                 "ingressController.serviceAccount",
		"controller.deployment.hostNetwork":                    "ingressController.deployment.pod.hostNework",
		"controller.deployment.hostname":                       "ingressController.deployment.pod.hostname",
		"controller.deployment.tmpDir":                         "ingressController.deployment.pod.tmpDir",
		"controller.ingressController.enabled":                 "ingressController.enabled",
		"controller.ingressController.gatewayDiscovery":        "ingressController.gatewayDiscovery",
		"controller.ingressController.watchNamespaces":         "ingressController.watchNamespaces",
		"controller.ingressController.admissionWebhook":        "ingressController.admissionWebhook",
		"controller.ingressController.ingressClass":            "ingressController.ingressClass",
		"controller.ingressController.ingressClassAnnotations": "ingressController.ingressClassAnnotations",
		"controller.ingressController.rbac":                    "ingressController.rbac",
		"controller.ingressController.konnect":                 "ingressController.konnect",
		"controller.ingressController.adminApi":                "ingressController.adminApi",
		"controller.updateStrategy":                            "ingressController.deployment.updateStrategy",
		"controller.resources":                                 "ingressController.deployment.pod.container.resources",
		"controller.terminationGracePeriodSeconds":             "ingressController.deployment.pod.terminationGracePeriodSeconds",
		"controller.tolerations":                               "ingressController.deployment.pod.tolerations",
		"controller.nodeSelector":                              "ingressController.deployment.pod.nodeSelector",
		"controller.podAnnotations":                            "ingressController.deployment.pod.annotations",
		"controller.podLabels":                                 "ingressController.deployment.pod.labels",
		"controller.deploymentAnnotations":                     "ingressController.deployment.annotations",
		"controller.replicaCount":                              "ingressController.deployment.replicaCount",
		"controller.priorityClassName":                         "ingressController.deployment.pod.priorityClassName",
		"controller.securityContext":                           "ingressController.deployment.pod.securityContext",
		"controller.containerSecurityContext":                  "ingressController.deployment.pod.container.securityContext",
	}
}

func getIngressGatewayKeys() map[string]string {
	return map[string]string{}
}

type mapFunc func() map[string]string
