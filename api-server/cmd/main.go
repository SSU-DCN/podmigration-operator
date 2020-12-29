package main

import (
	"os"

	apiserver "github.com/SSU-DCN/podmigration-operator/api-server"
	podmigrationv1 "github.com/SSU-DCN/podmigration-operator/api/v1"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	kubelog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var (
	runLog = kubelog.Log.WithName("podmigration-cp").WithName("run")
	scheme = runtime.NewScheme()
)

func init() {
	// Initialize the scheme so that kubernetes dynamic client knows
	// how to work with new CRD and native kubernetes types
	_ = clientgoscheme.AddToScheme(scheme)
	_ = podmigrationv1.AddToScheme(scheme)
}

func main() {
	kubelog.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := apiserver.NewManager(ctrl.GetConfigOrDie(), apiserver.Options{
		Scheme:         scheme,
		Port:           5000,
		AllowedDomains: []string{},
	})
	if err != nil {
		runLog.Error(err, "unable to create api-server manager")
		os.Exit(1)
	}

	runLog.Info("starting api-server manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		runLog.Error(err, "problem running api-server manager")
		os.Exit(1)
	}
}
