// SPDX-FileCopyrightText: 2026 Starlight Romero
//
// SPDX-License-Identifier: GPL-3.0-or-later

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// UserParameters defines the filter criteria for looking up a single Linear user.
type UserParameters struct {
	// Linear user ID. If set, other filters are ignored.
	// +optional
	ID *string `json:"id,omitempty"`

	// Filter by user name (exact match).
	// +optional
	Name *string `json:"name,omitempty"`

	// Filter by email (exact match).
	// +optional
	Email *string `json:"email,omitempty"`

	// Filter by display name (exact match).
	// +optional
	DisplayName *string `json:"displayName,omitempty"`
}

// UserObservation contains the observed state of a Linear user.
type UserObservation struct {
	// The Linear user ID.
	ID string `json:"id,omitempty"`

	// The user's full name.
	Name string `json:"name,omitempty"`

	// The user's display name.
	DisplayName string `json:"displayName,omitempty"`

	// The user's email address.
	Email string `json:"email,omitempty"`

	// Whether the user is active.
	Active bool `json:"active,omitempty"`

	// Whether the user is an admin.
	Admin bool `json:"admin,omitempty"`

	// The user's profile URL.
	URL string `json:"url,omitempty"`
}

// UserSpec defines the desired state of User.
type UserSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     UserParameters `json:"forProvider"`
}

// UserStatus defines the observed state of User.
type UserStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        UserObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,linear}

// User is an observe-only resource that looks up a single Linear user.
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UserSpec   `json:"spec"`
	Status            UserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserList contains a list of Users.
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

// Repository type metadata.
var (
	User_Kind             = "User"
	User_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: User_Kind}.String()
	User_KindAPIVersion   = User_Kind + "." + CRDGroupVersion.String()
	User_GroupVersionKind = CRDGroupVersion.WithKind(User_Kind)
)

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}
