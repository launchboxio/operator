package main

import (
	"flag"
	"log"
	"os"

	crossplanev1 "github.com/crossplane/crossplane/apis/pkg/v1"
	corev1alpha1 "github.com/launchboxio/operator/api/v1alpha1"
	"github.com/launchboxio/operator/controllers"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	rootCmd = &cobra.Command{
		Use:   "operator",
		Short: "LaunchboxHQ Operator",
		Run: func(cmd *cobra.Command, args []string) {
			var metricsAddr string
			var enableLeaderElection bool
			var probeAddr string
			flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
			flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
			flag.BoolVar(&enableLeaderElection, "leader-elect", false,
				"Enable leader election for controller manager. "+
					"Enabling this will ensure there is only one active controller manager.")
			opts := zap.Options{
				Development: true,
			}
			opts.BindFlags(flag.CommandLine)
			flag.Parse()

			ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

			mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
				Scheme: scheme,
				//MetricsBindAddress:     metricsAddr,
				//Port:                   9443,
				HealthProbeBindAddress: probeAddr,
				LeaderElection:         enableLeaderElection,
				LeaderElectionID:       "de4bbe6f.launchboxhq.io",
				// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
				// when the Manager ends. This requires the binary to immediately end when the
				// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
				// speeds up voluntary leader transitions as the new leader don't have to wait
				// LeaseDuration time first.
				//
				// In the default scaffold provided, the program ends immediately after
				// the manager stops, so would be fine to enable this option. However,
				// if you are doing or is intended to do any operation such as perform cleanups
				// after the manager stops then its usage might be unsafe.
				// LeaderElectionReleaseOnCancel: true,
			})
			if err != nil {
				setupLog.Error(err, "unable to start manager")
				os.Exit(1)
			}

			if err = (&controllers.ProjectReconciler{
				Client: mgr.GetClient(),
				Scheme: mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", "Project")
				os.Exit(1)
			}

			if err = (&controllers.ClusterReconciler{
				Client: mgr.GetClient(),
				Scheme: mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", "Cluster")
				os.Exit(1)
			}
			if err = (&controllers.AddonReconciler{
				Client: mgr.GetClient(),
				Scheme: mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", "Addon")
				os.Exit(1)
			}
			//+kubebuilder:scaffold:builder

			if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
				setupLog.Error(err, "unable to set up health check")
				os.Exit(1)
			}
			if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
				setupLog.Error(err, "unable to set up ready check")
				os.Exit(1)
			}

			setupLog.Info("starting manager")
			if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
				setupLog.Error(err, "problem running manager")
				os.Exit(1)
			}
		},
	}
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(corev1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme

	utilruntime.Must(crossplanev1.AddToScheme(scheme))
	//utilruntime.Must(crossplanehelm.AddToScheme(scheme))
	//utilruntime.Must(crossplanek8s.AddToScheme(scheme))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
