// SPDX-FileCopyrightText: 2026 Starlight Romero
//
// SPDX-License-Identifier: GPL-3.0-or-later

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	resource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
)

func (mg *User) GetCondition(ct xpv1.ConditionType) xpv1.Condition { return mg.Status.GetCondition(ct) }
func (mg *User) SetConditions(c ...xpv1.Condition)                 { mg.Status.SetConditions(c...) }
func (mg *User) GetManagementPolicies() xpv1.ManagementPolicies    { return mg.Spec.ManagementPolicies }
func (mg *User) SetManagementPolicies(r xpv1.ManagementPolicies)   { mg.Spec.ManagementPolicies = r }
func (mg *User) GetProviderConfigReference() *xpv1.Reference       { return mg.Spec.ProviderConfigReference }
func (mg *User) SetProviderConfigReference(r *xpv1.Reference)      { mg.Spec.ProviderConfigReference = r }
func (mg *User) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}
func (mg *User) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}
func (mg *User) GetDeletionPolicy() xpv1.DeletionPolicy  { return mg.Spec.DeletionPolicy }
func (mg *User) SetDeletionPolicy(r xpv1.DeletionPolicy) { mg.Spec.DeletionPolicy = r }

var _ resource.Managed = &User{}
