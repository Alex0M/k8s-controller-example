package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	frontendpagev1alpha1 "github.com/Alex0M/k8s-controller-example/apis/frontendpage/v1alpha1"
	frontendsyncv1alpha1 "github.com/Alex0M/k8s-controller-example/apis/frontendsync/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type FrontendSyncReconciler struct {
	client.Client
	scheme *runtime.Scheme
}

type FrontendPageData struct {
	Image    string `json:"image"`
	Replicas int    `json:"replicas"`
	Content  string `json:"content"`
}

func (r *FrontendSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("fontendsync", req.NamespacedName)
	log.Info("reconciling FrontendSync")

	var frontendsync frontendsyncv1alpha1.FrontendSync

	if err := r.Get(ctx, req.NamespacedName, &frontendsync); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.Info("get FrontendPage data from API")

	resp, err := http.Get(frontendsync.Spec.Url)
	if err != nil {
		return ctrl.Result{}, err
	}
	defer resp.Body.Close()

	var feData FrontendPageData
	if err := json.NewDecoder(resp.Body).Decode(&feData); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("reconciling FrontendPage")

	fe := buildFrontendPageObject(&frontendsync, &feData)
	log.Info("traing to set reference FrontendPage")
	if err := ctrl.SetControllerReference(&frontendsync, fe, r.scheme); err != nil {
		return ctrl.Result{}, err
	}

	var existingFePage frontendpagev1alpha1.FrontendPage

	log.Info("treying to get existing FrontendPage")
	if err := r.Get(ctx, req.NamespacedName, &existingFePage); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "unable to get FrondentPage")
			return ctrl.Result{}, err
		}

		log.Info("treying to create existing FrontendPage")
		if err := r.Create(ctx, fe); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		log.Info("treying to update existing FrontendPage")
		update := false

		if existingFePage.Spec.Contents != fe.Spec.Contents {
			update = true
			existingFePage.Spec.Contents = fe.Spec.Contents
		}

		if existingFePage.Spec.Image != fe.Spec.Image {
			update = true
			existingFePage.Spec.Image = fe.Spec.Image
		}

		if existingFePage.Spec.Replicas != fe.Spec.Replicas {
			update = true
			existingFePage.Spec.Replicas = fe.Spec.Replicas
		}

		if update {
			if err := r.Update(ctx, &existingFePage); err != nil {
				if apierrors.IsConflict(err) {
					return ctrl.Result{Requeue: true}, err
				}
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{RequeueAfter: time.Second * time.Duration(frontendsync.Spec.SyncInterval)}, nil
}

func NewFrontSyncController(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&frontendsyncv1alpha1.FrontendSync{}).
		Owns(&frontendpagev1alpha1.FrontendPage{}).
		Complete(&FrontendSyncReconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})
}
