package core

// NOTE For the kong chart, several keys (e.g. podAnnotations) affected both the controller and proxy. This tool _does
// not_ copy these from the root-level (now proxy) configuration to the controller Deployment/Pod/etc. configuration.
// There's unfortunately no way to know if individual annotations and the like were in place for the controller or
// proxy. Leaving them as-is (so that they now apply to the proxy) isn't an ideal solution, but it is the simplest.

// getControllerKeys returns a map of existing key locations to their new locations for the "kong" chart.
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

// getIngressControllerKeys returns a map of existing key locations to their new locations for the "ingress" chart.
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
