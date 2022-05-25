package constants

const (
	DevopsNamespace = "devops-system"

	// ConfigmapEnvVarPostfix is a postfix for configmap envVar
	ConfigmapEnvVarPostfix = "CONFIGMAP"
	// SecretEnvVarPostfix is a postfix for secret envVar
	SecretEnvVarPostfix = "SECRET"
	// EnvVarPrefix is a Prefix for environment variable
	EnvVarPrefix = "EFUNDS_"

	ConfigmapUpdateOnChangeAnnotation = "configmap.reloader.efunds.com/reload"
	// SecretUpdateOnChangeAnnotation is an annotation to detect changes in
	// secrets specified by name
	SecretUpdateOnChangeAnnotation = "secret.reloader.efunds.com/reload"
	// ReloaderAutoAnnotation is an annotation to detect changes in secrets
	ReloaderAutoAnnotation = "reloader.efunds.com/auto"
	// AutoSearchAnnotation is an annotation to detect changes in
	// configmaps or triggers with the SearchMatchAnnotation
	AutoSearchAnnotation = "reloader.efunds.com/search"
	// SearchMatchAnnotation is an annotation to tag secrets to be found with
	// AutoSearchAnnotation
	SearchMatchAnnotation = "reloader.efunds.com/match"
	// LogFormat is the log format to use (json, or empty string for default)
	LogFormat = ""
	// ReloadStrategy Specify the update strategy
	// ReloadOnCreate Adds support to watch create events
	ReloadOnCreate = "false"

	AllowedSecretType = "Opaque"
)
