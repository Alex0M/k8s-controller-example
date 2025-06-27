package controller

import (
	frontendpagev1alpha1 "github.com/Alex0M/k8s-controller-example/apis/frontendpage/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildConfigMapObject(page *frontendpagev1alpha1.FrontendPage) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      page.Name,
			Namespace: page.Namespace,
		},
		Data: map[string]string{
			"content": page.Spec.Contents,
		},
	}
}
