package app

import (
	"context"
	"github.com/hchenc/reloader/cmd/app/options"
	"github.com/hchenc/reloader/pkg/controllers"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"time"
)

var (
	leaderElection = &leaderelection.LeaderElectionConfig{}
	config         = options.NewKubernetesConfig()
)

// NewReloaderCommand starts the reloader controller
func NewReloaderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reloader",
		Short: "A watcher for your Kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if errs := validateFlags(); len(errs) != 0 {
				klog.Error(utilerrors.NewAggregate(errs))
				os.Exit(1)
			}

			if err := run(config, signals.SetupSignalHandler()); err != nil {
				klog.Error(err)
				os.Exit(1)
			}
		},
	}

	// options
	cmd.PersistentFlags().BoolVar(&options.LeaderElect, "leader-elect", true, "")

	return cmd
}

func validateFlags() []error {
	var errs []error
	errs = append(errs, config.Validate()...)
	// Ensure the reload strategy is one of the following...
	if options.LeaderElect {
		leaderElection.LeaseDuration = 30 * time.Second
		leaderElection.RenewDeadline = 15 * time.Second
		leaderElection.RetryPeriod = 5 * time.Second
	}
	return errs
}

func run(config *options.KubernetesOptions, ctx context.Context) error {
	scheme := runtime.NewScheme()

	mgrOptions := manager.Options{
		Scheme: scheme,
		Port:   9443,
	}
	//if options.LeaderElect {
	//	mgrOptions.LeaderElection = options.LeaderElect
	//	mgrOptions.LeaderElectionNamespace = constants.DevopsNamespace
	//	mgrOptions.LeaderElectionID = "reloader-controller-manager-leader-election"
	//	mgrOptions.LeaseDuration = &leaderElection.LeaseDuration
	//	mgrOptions.RetryPeriod = &leaderElection.RetryPeriod
	//	mgrOptions.RenewDeadline = &leaderElection.RenewDeadline
	//}
	klog.V(0).Info("setting up manager")
	ctrl.SetLogger(klogr.New())

	// Use 8443 instead of 443 cause we need root permission to bind port 443
	mgr, err := manager.New(config.KubeConfig, mgrOptions)
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}

	controller := controllers.NewControllerOrDie(mgr)
	if err = controller.Reconcile(ctx); err != nil {
		klog.Fatalf("unable to run the manager: %v", err)
	}
	return nil
	//collectors := metrics.SetupPrometheusEndpoint()

}

func getStringSliceFromFlags(cmd *cobra.Command, flag string) ([]string, error) {
	slice, err := cmd.Flags().GetStringSlice(flag)
	if err != nil {
		return nil, err
	}

	return slice, nil
}
