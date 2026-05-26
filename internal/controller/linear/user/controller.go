// SPDX-FileCopyrightText: 2026 Starlight Romero
//
// SPDX-License-Identifier: GPL-3.0-or-later

package user

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	tjcontroller "github.com/crossplane/upjet/v2/pkg/controller"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/avodah-inc/provider-linear/apis/linear/v1alpha1"
	"github.com/avodah-inc/provider-linear/internal/clients"
	"github.com/avodah-inc/provider-linear/internal/features"
)

const linearAPI = "https://api.linear.app/graphql"

// Setup adds the User observe-only controller to the manager.
func Setup(mgr ctrl.Manager, o tjcontroller.Options) error {
	name := managed.ControllerName(v1alpha1.User_GroupVersionKind.String())
	opts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), log: o.Logger}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithPollInterval(o.PollInterval),
	}
	if o.Features.Enabled(features.EnableBetaManagementPolicies) {
		opts = append(opts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr, xpresource.ManagedKind(v1alpha1.User_GroupVersionKind), opts...)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(xpresource.DesiredStateChanged()).
		For(&v1alpha1.User{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// SetupGated registers the controller via the gating mechanism.
func SetupGated(mgr ctrl.Manager, o tjcontroller.Options) error {
	return Setup(mgr, o)
}

type connector struct {
	kube client.Client
	log  logging.Logger
}

func (c *connector) Connect(ctx context.Context, mg xpresource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.User)
	if !ok {
		return nil, errors.New("managed resource is not a User")
	}

	pcRef := cr.Spec.ProviderConfigReference
	if pcRef == nil {
		return nil, errors.New("no providerConfigRef set")
	}

	token, err := clients.ResolveTokenFromProviderConfig(ctx, c.kube, pcRef.Name)
	if err != nil {
		return nil, err
	}

	return &external{token: token}, nil
}

type external struct {
	token string
}

func (e *external) Observe(ctx context.Context, mg xpresource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.User)
	if !ok {
		return managed.ExternalObservation{}, errors.New("managed resource is not a User")
	}

	params := cr.Spec.ForProvider

	userID := ""
	if params.ID != nil && *params.ID != "" {
		userID = *params.ID
	} else if extName := meta.GetExternalName(cr); extName != "" && extName != cr.Name {
		userID = extName
	}

	var user *linearUser
	var err error

	if userID != "" {
		user, err = e.getUser(ctx, userID)
	} else {
		user, err = e.findUser(ctx, params)
	}
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	if user == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider = v1alpha1.UserObservation{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Active:      user.Active,
		Admin:       user.Admin,
		URL:         user.URL,
	}

	meta.SetExternalName(cr, user.ID)
	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
}

func (e *external) Create(_ context.Context, _ xpresource.Managed) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(_ context.Context, _ xpresource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(_ context.Context, _ xpresource.Managed) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, nil
}

func (e *external) Disconnect(_ context.Context) error {
	return nil
}

type linearUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Active      bool   `json:"active"`
	Admin       bool   `json:"admin"`
	URL         string `json:"url"`
}

func (e *external) getUser(ctx context.Context, id string) (*linearUser, error) {
	query := `query($id: String!) { user(id: $id) { id name displayName email active admin url } }`
	vars := map[string]any{"id": id}

	var resp struct {
		Data struct {
			User linearUser `json:"user"`
		} `json:"data"`
	}
	if err := e.graphql(ctx, query, vars, &resp); err != nil {
		return nil, err
	}
	if resp.Data.User.ID == "" {
		return nil, nil
	}
	return &resp.Data.User, nil
}

func (e *external) findUser(ctx context.Context, params v1alpha1.UserParameters) (*linearUser, error) {
	query := `query { users { nodes { id name displayName email active admin url } } }`

	var resp struct {
		Data struct {
			Users struct {
				Nodes []linearUser `json:"nodes"`
			} `json:"users"`
		} `json:"data"`
	}
	if err := e.graphql(ctx, query, nil, &resp); err != nil {
		return nil, err
	}

	for _, u := range resp.Data.Users.Nodes {
		if params.Name != nil && u.Name != *params.Name {
			continue
		}
		if params.Email != nil && u.Email != *params.Email {
			continue
		}
		if params.DisplayName != nil && u.DisplayName != *params.DisplayName {
			continue
		}
		return &u, nil
	}
	return nil, nil
}

func (e *external) graphql(ctx context.Context, query string, variables map[string]any, result any) error {
	body := map[string]any{"query": query}
	if variables != nil {
		body["variables"] = variables
	}
	payload, _ := json.Marshal(body)

	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, linearAPI, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", e.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Linear API returned %d: %s", resp.StatusCode, string(respBody))
	}
	return json.Unmarshal(respBody, result)
}
