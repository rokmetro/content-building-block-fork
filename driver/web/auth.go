// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	"content/core"
	"log"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v3/authorization"
	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// Authorization is an interface for auth types
type Authorization interface {
	check(req *http.Request) (int, *tokenauth.Claims, error)
	start()
}

// Auth handler
type Auth struct {
	coreAuth *CoreAuth
	bbs      tokenauth.Handlers
	tps      tokenauth.Handlers
	logger   *logs.Logger
}

// NewAuth creates new auth handler
func NewAuth(app *core.Application, serviceRegManager *authservice.ServiceRegManager, logger *logs.Logger) *Auth {
	coreAuth := NewCoreAuth(app, serviceRegManager)

	bbsStandardHandler, err := newBBsStandardHandler(serviceRegManager)
	if err != nil {
		return nil
	}
	bbsHandlers := tokenauth.NewHandlers(bbsStandardHandler) //add permissions, user and authenticated

	tpsStandardHandler, err := newTPsStandardHandler(serviceRegManager)
	if err != nil {
		return nil
	}
	tpsHandlers := tokenauth.NewHandlers(tpsStandardHandler) //add permissions, user and authenticated

	auth := Auth{coreAuth: coreAuth, bbs: bbsHandlers, tps: tpsHandlers, logger: logger}
	return &auth
}

// CoreAuth implementation
type CoreAuth struct {
	app       *core.Application
	tokenAuth *tokenauth.TokenAuth

	permissionsAuth *PermissionsAuth
	userAuth        *UserAuth
	standardAuth    *StandardAuth
}

// NewCoreAuth creates new CoreAuth
func NewCoreAuth(app *core.Application, serviceRegManager *authservice.ServiceRegManager) *CoreAuth {
	adminPermissionAuth := authorization.NewCasbinAuthorization("driver/web/authorization_model.conf", "driver/web/authorization_policy.csv")
	tokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, adminPermissionAuth, nil)
	if err != nil {
		log.Fatalf("Error intitializing token auth: %v", err)
	}
	permissionsAuth := newPermissionsAuth(tokenAuth)
	usersAuth := newUserAuth(tokenAuth)
	standardAuth := newStandardAuth(tokenAuth)

	auth := CoreAuth{app: app, tokenAuth: tokenAuth, permissionsAuth: permissionsAuth,
		userAuth: usersAuth, standardAuth: standardAuth}
	return &auth
}

// BBs auth ///////////
func newBBsStandardHandler(serviceRegManager *authservice.ServiceRegManager) (*tokenauth.StandardHandler, error) {
	bbsPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/authorization_bbs_permission_policy.csv")
	bbsTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, bbsPermissionAuth, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "bbs token auth", nil, err)
	}

	check := func(claims *tokenauth.Claims, req *http.Request) (int, error) {
		if !claims.Service {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "service claim", nil)
		}

		if !claims.FirstParty {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "first party claim", nil)
		}

		return http.StatusOK, nil
	}

	auth := tokenauth.NewStandardHandler(bbsTokenAuth, check)
	return auth, nil
}

// TPs auth ///////////
func newTPsStandardHandler(serviceRegManager *authservice.ServiceRegManager) (*tokenauth.StandardHandler, error) {
	tpsPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/authorization_tps_permission_policy.csv")
	tpsTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, tpsPermissionAuth, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "tps token auth", nil, err)
	}

	check := func(claims *tokenauth.Claims, req *http.Request) (int, error) {
		if !claims.Service {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "service claim", nil)
		}

		if claims.FirstParty {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "first party claim", nil)
		}

		return http.StatusOK, nil
	}

	auth := tokenauth.NewStandardHandler(tpsTokenAuth, check)
	return auth, nil
}

// PermissionsAuth entity
// This enforces that the user has permissions matching the policy
type PermissionsAuth struct {
	tokenAuth *tokenauth.TokenAuth
}

func (a *PermissionsAuth) start() {}

func (a *PermissionsAuth) check(req *http.Request) (int, *tokenauth.Claims, error) {
	claims, err := a.tokenAuth.CheckRequestToken(req)
	if err != nil {
		return http.StatusUnauthorized, nil, errors.WrapErrorAction("typeCheckServicesAuthRequestToken", logutils.TypeToken, nil, err)
	}

	if err == nil && claims != nil {
		err = a.tokenAuth.AuthorizeRequestPermissions(claims, req)
		if err != nil {
			return http.StatusForbidden, nil, errors.WrapErrorAction("check permissions", logutils.TypeRequest, nil, err)
		}
	}

	return http.StatusOK, claims, err
}

func newPermissionsAuth(tokenAuth *tokenauth.TokenAuth) *PermissionsAuth {
	permissionsAuth := PermissionsAuth{tokenAuth: tokenAuth}
	return &permissionsAuth
}

// UserAuth entity
// This enforces that the user is not anonymous
type UserAuth struct {
	tokenAuth *tokenauth.TokenAuth
}

func (a *UserAuth) start() {}

func (a *UserAuth) check(req *http.Request) (int, *tokenauth.Claims, error) {
	claims, err := a.tokenAuth.CheckRequestToken(req)
	if err != nil {
		return http.StatusUnauthorized, nil, errors.WrapErrorAction("typeCheckServicesAuthRequestToken", logutils.TypeToken, nil, err)
	}

	if err == nil && claims != nil {
		if claims.Anonymous {
			return http.StatusForbidden, nil, errors.New("token must not be anonymous")
		}
	}

	return http.StatusOK, claims, err
}

func newUserAuth(tokenAuth *tokenauth.TokenAuth) *UserAuth {
	userAuth := UserAuth{tokenAuth: tokenAuth}
	return &userAuth
}

// StandardAuth entity
// This enforces standard auth check
type StandardAuth struct {
	tokenAuth *tokenauth.TokenAuth
}

func (a *StandardAuth) start() {}

func (a *StandardAuth) check(req *http.Request) (int, *tokenauth.Claims, error) {
	claims, err := a.tokenAuth.CheckRequestToken(req)
	if err != nil {
		return http.StatusUnauthorized, nil, errors.WrapErrorAction("typeCheckServicesAuthRequestToken", logutils.TypeToken, nil, err)
	}

	return http.StatusOK, claims, err
}

func newStandardAuth(tokenAuth *tokenauth.TokenAuth) *StandardAuth {
	standartAuth := StandardAuth{tokenAuth: tokenAuth}
	return &standartAuth
}
