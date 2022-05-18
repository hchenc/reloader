package filters

import (
	"github.com/hchenc/reloader/pkg/constants"
	"github.com/hchenc/reloader/pkg/utils/crypto"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type NamespaceUpdatePredicate struct {
	filterPredicate
	//include namespaces has higher priority
	IncludeNamespaces []string
	ExcludeNamespaces []string
}

func (r NamespaceUpdatePredicate) Update(e event.UpdateEvent) bool {
	namespace := e.ObjectNew.GetNamespace()

	if exists, verified := checkIndexKey(r.IncludeNamespaces, namespace); verified {
		return exists
	}

	if exists, verified := checkIndexKey(r.ExcludeNamespaces, namespace); verified {
		return !exists
	}

	return false

}

type SecretDataUpdatePredicate struct {
	filterPredicate
}

func (s SecretDataUpdatePredicate) Create(e event.CreateEvent) bool {
	return false
}

func (s SecretDataUpdatePredicate) Update(e event.UpdateEvent) bool {
	oldSecret := e.ObjectOld.(*v1.Secret)
	newSecret := e.ObjectNew.(*v1.Secret)
	if oldSecret.Type != constants.AllowedSecretType || newSecret.Type != constants.AllowedSecretType {
		return false
	}
	oldSecretConfig := crypto.GetSecretConfig(e.ObjectOld.(*v1.Secret))
	newSecretConfig := crypto.GetSecretConfig(e.ObjectNew.(*v1.Secret))
	if newSecretConfig.SHAValue == oldSecretConfig.SHAValue {
		return false
	}
	return true
}

type ConfigMapDataUpdatePredicate struct {
	filterPredicate
}

func (c ConfigMapDataUpdatePredicate) Update(e event.UpdateEvent) bool {
	oldSecretConfig := crypto.GetConfigmapConfig(e.ObjectOld.(*v1.ConfigMap))
	newSecretConfig := crypto.GetConfigmapConfig(e.ObjectNew.(*v1.ConfigMap))
	if newSecretConfig.SHAValue == oldSecretConfig.SHAValue {
		return false
	}
	return true
}
