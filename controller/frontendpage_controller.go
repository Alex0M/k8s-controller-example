package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	frontendv1alpha1 "github.com/Alex0M/k8s-controller-example/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type FrontendPageReconciler struct {
	client.Client
	scheme *runtime.Scheme
}

type FrontendPageData struct {
	Image    string
	Replicas int
	Content  string
}

func (r *FrontendPageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("frontendpage", req.NamespacedName)
	log.V(1).Info("reconciling FrontendPage")

	var frontendPage frontendv1alpha1.FrontendPage

	if err := r.Get(ctx, req.NamespacedName, &frontendPage); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "unable to get FrontendPage")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	log.Info("get data from API")

	resp, err := http.Get(frontendPage.Spec.Url)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer resp.Body.Close()

	var feData FrontendPageData
	if err := json.NewDecoder(resp.Body).Decode(&feData); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("reconciling configMap")
	cm := buildConfigMapObject(&frontendPage, &feData)
	if err := ctrl.SetControllerReference(&frontendPage, cm, r.Scheme()); err != nil {
		return ctrl.Result{}, err
	}

	var existingConfigMap corev1.ConfigMap

	if err := r.Get(ctx, req.NamespacedName, &existingConfigMap); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "unable to get ConfigMap")
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, cm); err != nil {
			return ctrl.Result{}, err
		}
	} else if !reflect.DeepEqual(existingConfigMap.Data, cm.Data) {
		existingConfigMap.Data = cm.Data
		if err := r.Update(ctx, &existingConfigMap); err != nil {
			return ctrl.Result{}, err
		}
	}

	//2. Ensure Deployment exists and up to date
	log.Info("reconciling deployemnt")
	dp := buildDeploymentObject(&frontendPage, &feData)
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
	} else {
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
	}

	return ctrl.Result{RequeueAfter: time.Second * time.Duration(frontendPage.Spec.SyncInterval)}, nil
}

func NewFrontPageController(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frontendv1alpha1.FrontendPage{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Complete(&FrontendPageReconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})
}
