package controllers

import (
	"context"
	"github.com/hchenc/reloader/pkg/constants"
	"github.com/hchenc/reloader/pkg/utils/crypto"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	reconcilerMap = make(map[string]Reconciler)
)

type Reconciler interface {
	SetUp(mgr manager.Manager)
}

type Reconcile func(mgr manager.Manager)

func (r Reconcile) SetUp(mgr manager.Manager) {
	r(mgr)
}

func RegisterReconciler(name string, f Reconcile) {
	reconcilerMap[name] = f
}

type Controller struct {
	ReconcilerMap map[string]Reconciler

	manager manager.Manager
}

func NewControllerOrDie(mgr manager.Manager) *Controller {
	sche := mgr.GetScheme()
	runtime.Must(v1.AddToScheme(sche))
	runtime.Must(appsv1.AddToScheme(mgr.GetScheme()))

	for _, reconciler := range reconcilerMap {
		reconciler.SetUp(mgr)
	}

	return &Controller{
		ReconcilerMap: reconcilerMap,
		manager:       mgr,
	}
}

func (c *Controller) Reconcile(ctx context.Context) error {
	if err := c.manager.Start(ctx); err != nil {
		ctrl.Log.Error(err, "start reconciler failed")
		return err
	}
	return nil
}

func updateContainerEnvVars(deployment appsv1.Deployment, config crypto.Config, autoReload bool) constants.Result {
	var result constants.Result
	envVar := constants.EnvVarPrefix + crypto.ConvertToEnvVarName(config.ResourceName) + "_" + config.Type
	container := getContainerUsingResource(deployment, config, autoReload)

	if container == nil {
		return constants.NoContainerFound
	}

	//update if env var exists
	result = updateEnvVar(deployment.Spec.Template.Spec.Containers, envVar, config.SHAValue)

	// if no existing env var exists lets create one
	if result == constants.NoEnvVarFound {
		e := v1.EnvVar{
			Name:  envVar,
			Value: config.SHAValue,
		}
		container.Env = append(container.Env, e)
		result = constants.Updated
	}
	return result
}

func getContainerUsingResource(deployment appsv1.Deployment, config crypto.Config, autoReload bool) *v1.Container {
	volumes := deployment.Spec.Template.Spec.Volumes
	containers := deployment.Spec.Template.Spec.Containers
	initContainers := deployment.Spec.Template.Spec.InitContainers
	var container *v1.Container
	// Get the volumeMountName to find volumeMount in container
	volumeMountName := getVolumeMountName(volumes, config.Type, config.ResourceName)
	// Get the container with mounted configmap/secret
	if volumeMountName != "" {
		container = getContainerWithVolumeMount(containers, volumeMountName)
		if container == nil && len(initContainers) > 0 {
			container = getContainerWithVolumeMount(initContainers, volumeMountName)
			if container != nil {
				// if configmap/secret is being used in init container then return the first Pod container to save reloader env
				return &containers[0]
			}
		} else if container != nil {
			return container
		}
	}

	// Get the container with referenced secret or configmap as env var
	container = getContainerWithEnvReference(containers, config.ResourceName, config.Type)
	if container == nil && len(initContainers) > 0 {
		container = getContainerWithEnvReference(initContainers, config.ResourceName, config.Type)
		if container != nil {
			// if configmap/secret is being used in init container then return the first Pod container to save reloader env
			return &containers[0]
		}
	}

	// Get the first container if the annotation is related to specified configmap or secret i.e. configmap.reloader.stakater.com/reload
	if container == nil && !autoReload {
		return &containers[0]
	}

	return container
}

func getVolumeMountName(volumes []v1.Volume, mountType string, volumeName string) string {
	for i := range volumes {
		if mountType == constants.ConfigmapEnvVarPostfix {
			if volumes[i].ConfigMap != nil && volumes[i].ConfigMap.Name == volumeName {
				return volumes[i].Name
			}

			if volumes[i].Projected != nil {
				for j := range volumes[i].Projected.Sources {
					if volumes[i].Projected.Sources[j].ConfigMap != nil && volumes[i].Projected.Sources[j].ConfigMap.Name == volumeName {
						return volumes[i].Name
					}
				}
			}
		} else if mountType == constants.SecretEnvVarPostfix {
			if volumes[i].Secret != nil && volumes[i].Secret.SecretName == volumeName {
				return volumes[i].Name
			}

			if volumes[i].Projected != nil {
				for j := range volumes[i].Projected.Sources {
					if volumes[i].Projected.Sources[j].Secret != nil && volumes[i].Projected.Sources[j].Secret.Name == volumeName {
						return volumes[i].Name
					}
				}
			}
		}
	}

	return ""
}

func getContainerWithVolumeMount(containers []v1.Container, volumeMountName string) *v1.Container {
	for i := range containers {
		volumeMounts := containers[i].VolumeMounts
		for j := range volumeMounts {
			if volumeMounts[j].Name == volumeMountName {
				return &containers[i]
			}
		}
	}

	return nil
}

func getContainerWithEnvReference(containers []v1.Container, resourceName string, resourceType string) *v1.Container {
	for i := range containers {
		envs := containers[i].Env
		for j := range envs {
			envVarSource := envs[j].ValueFrom
			if envVarSource != nil {
				if resourceType == constants.SecretEnvVarPostfix && envVarSource.SecretKeyRef != nil && envVarSource.SecretKeyRef.LocalObjectReference.Name == resourceName {
					return &containers[i]
				} else if resourceType == constants.ConfigmapEnvVarPostfix && envVarSource.ConfigMapKeyRef != nil && envVarSource.ConfigMapKeyRef.LocalObjectReference.Name == resourceName {
					return &containers[i]
				}
			}
		}

		envsFrom := containers[i].EnvFrom
		for j := range envsFrom {
			if resourceType == constants.SecretEnvVarPostfix && envsFrom[j].SecretRef != nil && envsFrom[j].SecretRef.LocalObjectReference.Name == resourceName {
				return &containers[i]
			} else if resourceType == constants.ConfigmapEnvVarPostfix && envsFrom[j].ConfigMapRef != nil && envsFrom[j].ConfigMapRef.LocalObjectReference.Name == resourceName {
				return &containers[i]
			}
		}
	}
	return nil
}

func updateEnvVar(containers []v1.Container, envVar string, shaData string) constants.Result {
	for i := range containers {
		envs := containers[i].Env
		for j := range envs {
			if envs[j].Name == envVar {
				if envs[j].Value != shaData {
					envs[j].Value = shaData
					return constants.Updated
				}
				return constants.NotUpdated
			}
		}
	}
	return constants.NoEnvVarFound
}
