package crypto

import (
	"github.com/hchenc/reloader/pkg/constants"
	v1 "k8s.io/api/core/v1"
)

//Config contains rolling upgrade configuration parameters
type Config struct {
	Namespace           string
	ResourceName        string
	ResourceAnnotations map[string]string
	Annotation          string
	SHAValue            string
	Type                string
}

// GetConfigmapConfig provides utility config for configmap
func GetConfigmapConfig(configmap *v1.ConfigMap) Config {
	return Config{
		Namespace:           configmap.Namespace,
		ResourceName:        configmap.Name,
		ResourceAnnotations: configmap.Annotations,
		Annotation:          constants.ConfigmapUpdateOnChangeAnnotation,
		SHAValue:            GetSHAfromConfigmap(configmap),
		Type:                constants.ConfigmapEnvVarPostfix,
	}
}

// GetSecretConfig provides utility config for secret
func GetSecretConfig(secret *v1.Secret) Config {
	return Config{
		Namespace:           secret.Namespace,
		ResourceName:        secret.Name,
		ResourceAnnotations: secret.Annotations,
		Annotation:          constants.SecretUpdateOnChangeAnnotation,
		SHAValue:            GetSHAfromSecret(secret.Data),
		Type:                constants.SecretEnvVarPostfix,
	}
}
