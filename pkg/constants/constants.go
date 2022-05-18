package constants

const (
	DevopsNamespace = "devops-system"

	// ConfigmapEnvVarPostfix is a postfix for configmap envVar
	ConfigmapEnvVarPostfix = "CONFIGMAP"
	// SecretEnvVarPostfix is a postfix for secret envVar
	SecretEnvVarPostfix = "SECRET"
	// EnvVarPrefix is a Prefix for environment variable
	EnvVarPrefix = "EFUNDS_"

	// ReloaderAnnotationPrefix is a Prefix for all reloader annotations
	ReloaderAnnotationPrefix = "reloader.efunds.com"
	// LastReloadedFromAnnotation is an annotation used to describe the last resource that triggered a reload
	LastReloadedFromAnnotation = "last-reloaded-from"

	// 	ReloadStrategyFlag The reload strategy flag name
	ReloadStrategyFlag = "reload-strategy"
	// EnvVarsReloadStrategy instructs Reloader to add container environment variables to facilitate a restart
	EnvVarsReloadStrategy = "env-vars"
	// AnnotationsReloadStrategy instructs Reloader to add pod template annotations to facilitate a restart
	AnnotationsReloadStrategy = "annotations"

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
	ReloadStrategy = EnvVarsReloadStrategy
	// ReloadOnCreate Adds support to watch create events
	ReloadOnCreate = "false"

	AllowedSecretType = "Opaque"
)
