package main

import (
	"context"
	"os"

	api "github.com/Alex0M/k8s-controller-example/api/v1alpha1"
	"github.com/Alex0M/k8s-controller-example/internal"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

type reconciler struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("frontendpage", req.NamespacedName)
	log.V(1).Info("reconciling FrontendPage")

	var frontendPage api.FrontendPage
	frontendPageFound := true

	if err := r.Get(ctx, req.NamespacedName, &frontendPage); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "unable to get FrontendPage")
			return ctrl.Result{}, err
		}
		frontendPageFound = false
	}

	var deployment appsv1.Deployment
	var configmap corev1.ConfigMap
	deployment.Name = req.Name
	configmap.Name = req.Name

	if frontendPageFound {
		log.Info("checking deploymnet")

		if err := r.Get(ctx, req.NamespacedName, &configmap); err != nil {
			if !apierrors.IsNotFound(err) {
				log.Error(err, "unable to get configmap")
				return ctrl.Result{}, err
			}
			log.Info("configmap not found")
			if err := r.Create(ctx, internal.GetConfigMapObject(req.Name, req.Namespace, frontendPage.Spec.Contents)); err != nil {
				log.Error(err, "unable to create configmap")
				return ctrl.Result{}, err
			}
		}

		if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
			if !apierrors.IsNotFound(err) {
				log.Error(err, "unable to get deployemnt")
				return ctrl.Result{}, err
			}

			log.Info("deployemnt not found")
			if err := r.Create(ctx, internal.GetDeploymentObject(req.Name, req.Namespace, frontendPage.Spec.Image, frontendPage.Spec.Replicas)); err != nil {
				log.Error(err, "unable to create deployemnt")
				return ctrl.Result{}, err
			}
		}

		if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
			if !apierrors.IsNotFound(err) {
				log.Error(err, "unable to get deployemnt")
				return ctrl.Result{}, err
			}
		}

		if *deployment.Spec.Replicas != int32(frontendPage.Spec.Replicas) || deployment.Spec.Template.Spec.Containers[0].Image != frontendPage.Spec.Image {
			deployment.Spec.Replicas = ptr.To[int32](int32(frontendPage.Spec.Replicas))
			deployment.Spec.Template.Spec.Containers[0].Image = frontendPage.Spec.Image

			if err := r.Update(ctx, &deployment); err != nil {
				log.Error(err, "unable to update deployment")
				return ctrl.Result{}, err
			}

			log.Info("deployemnt is updated")
			return ctrl.Result{}, nil
		}

		log.Info("deployemnt is up to date")
		return ctrl.Result{}, nil
	} else {
		log.Info("goging to delete config map")
		configmap.Namespace = req.Namespace
		if err := r.Delete(ctx, &configmap); err != nil {
			log.Error(err, "unable to delete configmap")
			return ctrl.Result{}, err
		}
		log.Info("config map has been deleted")
		log.Info("goging to delete config map")
		deployment.Namespace = req.Namespace
		if err := r.Delete(ctx, &deployment); err != nil {
			log.Error(err, "unable to delete deployment")
			return ctrl.Result{}, err
		}

		log.Info("deployment is deleted")
		return ctrl.Result{}, nil
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

	err = api.AddToScheme(mgr.GetScheme())
	if err != nil {
		setupLog.Error(err, "unable to add scheme")
		os.Exit(1)
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&api.FrontendPage{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Complete(&reconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})
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
