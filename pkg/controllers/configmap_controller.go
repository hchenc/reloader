package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/hchenc/reloader/pkg/constants"
	"github.com/hchenc/reloader/pkg/controllers/filters"
	"github.com/hchenc/reloader/pkg/utils/crypto"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
)

var (
	configmapAction = "ConfigMapReloader"
)

func init() {
	RegisterReconciler(configmapAction, SetUpConfigMapReconcile)
}

type ConfigMapOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (c *ConfigMapOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	configmap := &v1.ConfigMap{}

	err := c.Get(ctx, req.NamespacedName, configmap)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			c.Log.Error(err, fmt.Sprintf("failed to reconcile secret <%s>", req.Name))
			return ctrl.Result{}, err
		}
	}

	if configmap.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	configmapConfig := crypto.GetConfigmapConfig(configmap)

	//preprocess deployment list
	deploymentList := &appsv1.DeploymentList{}

	//if err := c.List(ctx, deploymentList, client.InNamespace(req.Namespace), client.MatchingLabels{constants.ReloaderAutoAnnotation: "true"}); err != nil {
	//	c.Log.Error(err, fmt.Sprintf("Failed to list deployments %v", err))
	//}

	//cancel label selector
	if err := c.List(ctx, deploymentList, client.InNamespace(req.Namespace)); err != nil {
		c.Log.Error(err, fmt.Sprintf("Failed to list deployments %v", err))
	}

	for _, dp := range deploymentList.Items {
		if value, exist := dp.Annotations[constants.ReloaderAutoAnnotation]; exist {
			if noAction, _ := strconv.ParseBool(value); !noAction {
				continue
			}
		}
		// find correct annotation and update the resource
		result := constants.NotUpdated

		result = updateContainerEnvVars(dp, configmapConfig, true)

		if result == constants.Updated {
			if err := c.Update(ctx, &dp, &client.UpdateOptions{FieldManager: "Reloader"}); err != nil {
				logrus.Errorf("Update for '%s' in namespace '%s' failed with error %v", dp.Name, dp.Namespace, err)
				return reconcile.Result{}, err
			} else {
				logrus.Infof("Changes detected in '%s' in namespace '%s'", dp.Name, dp.Namespace)
				logrus.Infof("Updated '%s' in namespace '%s'", dp.Name, dp.Namespace)
			}
		}
	}

	return reconcile.Result{}, nil
}

func (c *ConfigMapOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ConfigMap{}).
		WithEventFilter(
			predicate.And(
				&filters.NamespaceUpdatePredicate{
					IncludeNamespaces: filters.DefaultIncludeNamespaces,
				},
				&filters.ConfigMapDataUpdatePredicate{},
			),
		).
		Complete(c)
}

func SetUpConfigMapReconcile(mgr manager.Manager) {
	if err := (&ConfigMapOperatorReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName(configmapAction),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create configmap controller")
	}
}
