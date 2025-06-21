package main

import (
	"context"
	"os"
	"reflect"

	api "github.com/Alex0M/k8s-controller-example/api/v1alpha1"
	"github.com/Alex0M/k8s-controller-example/internal"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

	if frontendPageFound {
		log.Info("reconciling configMap")

		cm := internal.GetConfigMapObject(req.Name, req.Namespace, frontendPage.Spec.Contents)
		if err := ctrl.SetControllerReference(&frontendPage, cm, r.Scheme()); err != nil {
			return ctrl.Result{}, err
		}

		var existingConfigMap corev1.ConfigMap
		existingConfigMap.Name = req.Name

		if err := r.Get(ctx, req.NamespacedName, &existingConfigMap); err != nil {
			if !apierrors.IsNotFound(err) {
				log.Error(err, "unable to get ConfigMap")
				return ctrl.Result{}, err
			}

			if err := r.Create(ctx, cm); err != nil {
				return ctrl.Result{}, err
			}
		}

		if !reflect.DeepEqual(existingConfigMap.Data, cm.Data) {
			existingConfigMap.Data = cm.Data
			if err := r.Update(ctx, &existingConfigMap); err != nil {
				return ctrl.Result{}, err
			}
		}

		//2. Ensure Deployment exists and up to date
		log.Info("reconciling deployemnt")
		dp := internal.GetDeploymentObject(req.Name, req.Namespace, frontendPage.Spec.Image, frontendPage.Spec.Replicas)
		if err := ctrl.SetControllerReference(&frontendPage, dp, r.Scheme()); err != nil {
			return ctrl.Result{}, err
		}

		var existingDeployemnt appsv1.Deployment
		existingDeployemnt.Name = req.Name

		if err := r.Get(ctx, req.NamespacedName, &existingDeployemnt); err != nil {
			if !apierrors.IsNotFound(err) {
				log.Error(err, "unable to get Deployment")
				return ctrl.Result{}, err
			}

			if err := r.Create(ctx, dp); err != nil {
				log.Error(err, "unable to create Deployment")
				return ctrl.Result{}, err
			}
		}

		dpUpdate := false
		if *existingDeployemnt.Spec.Replicas != *dp.Spec.Replicas {
			existingDeployemnt.Spec.Replicas = dp.Spec.Replicas
			dpUpdate = true
		}

		if existingDeployemnt.Spec.Template.Spec.Containers[0].Image != dp.Spec.Template.Spec.Containers[0].Image {
			existingDeployemnt.Spec.Template.Spec.Containers[0].Image = dp.Spec.Template.Spec.Containers[0].Image
			dpUpdate = true
		}

		if dpUpdate {
			if err := r.Update(ctx, &existingDeployemnt); err != nil {
				if apierrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, err
				}
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	} else {
		log.Info("goging to delete config map")
		var configmap corev1.ConfigMap
		configmap.Namespace = req.Namespace
		configmap.Name = req.Name
		if err := r.Delete(ctx, &configmap); err != nil {
			log.Error(err, "unable to delete configmap")
			return ctrl.Result{}, err
		}
		log.Info("config map has been deleted")
		log.Info("goging to delete config map")
		var deployment appsv1.Deployment
		deployment.Namespace = req.Namespace
		deployment.Name = req.Name
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
