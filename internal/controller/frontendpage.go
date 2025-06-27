package controller

import (
	frontendv1alpha1 "github.com/Alex0M/k8s-controller-example/apis/frontendpage/v1alpha1"
	frontendsyncv1alpha1 "github.com/Alex0M/k8s-controller-example/apis/frontendsync/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildFrontendPageObject(feSync *frontendsyncv1alpha1.FrontendSync, feData *FrontendPageData) *frontendv1alpha1.FrontendPage {
	return &frontendv1alpha1.FrontendPage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      feSync.Name,
			Namespace: feSync.Namespace,
		},
		Spec: frontendv1alpha1.FrontendPageSpec{
			Contents: feData.Content,
			Image:    feData.Image,
			Replicas: feData.Replicas,
		},
	}
}
