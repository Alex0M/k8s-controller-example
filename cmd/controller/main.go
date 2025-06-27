package main

import (
	"flag"
	"log"
	"net/url"
	"os"

	frontendpagev1alpha1 "github.com/Alex0M/k8s-controller-example/apis/frontendpage/v1alpha1"
	frontendsyncv1alpha1 "github.com/Alex0M/k8s-controller-example/apis/frontendsync/v1alpha1"
	"github.com/Alex0M/k8s-controller-example/internal/controller"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	setupLog = ctrl.Log.WithName("setup")
	apiUrl   string
	syncTime int //sec
)

func init() {
	flag.StringVar(&apiUrl, "apiUrl", os.Getenv("API_URL"), "API URL to get data")
	flag.IntVar(&syncTime, "syncTime", 30, "sync time in seconds")
	flag.Parse()

	if _, err := url.Parse(apiUrl); err != nil {
		log.Fatalf("API URL is not valid url %v", err)
	}
}

func main() {
	ctrl.SetLogger(zap.New())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Metrics: server.Options{
			BindAddress: "localhost:8082",
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	err = frontendpagev1alpha1.AddToScheme(mgr.GetScheme())
	if err != nil {
		setupLog.Error(err, "unable to add frontendpage scheme")
		os.Exit(1)
	}

	err = frontendsyncv1alpha1.AddToScheme(mgr.GetScheme())
	if err != nil {
		setupLog.Error(err, "unable to add frontendpage scheme")
		os.Exit(1)
	}

	err = controller.NewFrontPageController(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	err = controller.NewFrontSyncController(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
