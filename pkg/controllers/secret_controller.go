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
	secretAction = "SecretReloader"
)

func init() {
	RegisterReconciler(secretAction, SetUpSecretReconcile)
}

type SecretOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (s *SecretOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	secret := &v1.Secret{}

	err := s.Get(ctx, req.NamespacedName, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			s.Log.Error(err, fmt.Sprintf("failed to reconcile secret <%s>", req.Name))
			return ctrl.Result{}, err
		}
	}

	if secret.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}
	secretConfig := crypto.GetSecretConfig(secret)

	//preprocess deployment list
	// Ensure we always have pod annotations to add to
	deploymentList := &appsv1.DeploymentList{}
	//deploymentListOption := client.MatchingLabels{constants.ReloaderAutoAnnotation: "true"}
	//
	//if err := s.List(ctx, deploymentList, client.InNamespace(req.Namespace), deploymentListOption); err != nil {
	//	s.Log.Error(err, fmt.Sprintf("Failed to list deployments %v", err))
	//}

	//cancel label selector
	if err := s.List(ctx, deploymentList, client.InNamespace(req.Namespace)); err != nil {
		s.Log.Error(err, fmt.Sprintf("Failed to list deployments %v", err))
	}

	for _, dp := range deploymentList.Items {
		// find correct annotation and update the resource
		if value, exist := dp.Annotations[constants.ReloaderAutoAnnotation]; exist {
			if noAction, _ := strconv.ParseBool(value); !noAction {
				continue
			}
		}
		result := constants.NotUpdated

		result = updateContainerEnvVars(dp, secretConfig, true)

		if result == constants.Updated {
			if err := s.Update(ctx, &dp, &client.UpdateOptions{FieldManager: "Reloader"}); err != nil {
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

func (s *SecretOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Secret{}).
		WithEventFilter(
			predicate.And(
				&filters.NamespaceUpdatePredicate{
					IncludeNamespaces: filters.DefaultIncludeNamespaces,
				},
				&filters.SecretDataUpdatePredicate{},
			),
		).
		Complete(s)
}

func SetUpSecretReconcile(mgr manager.Manager) {
	if err := (&SecretOperatorReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName(secretAction),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create secret controller")
	}
}
