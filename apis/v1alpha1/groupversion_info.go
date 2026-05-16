// SPDX-FileCopyrightText: 2026 Starlight Romero
//
// SPDX-License-Identifier: GPL-3.0-or-later

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	CRDGroup   = "linear.crossplane.io"
	CRDVersion = "v1alpha1"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}
	SchemeBuilder      = &scheme.Builder{GroupVersion: SchemeGroupVersion}
	AddToScheme        = SchemeBuilder.AddToScheme
)
