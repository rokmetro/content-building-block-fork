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
	"content/driver/web/rest"
	"content/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/core-auth-library-go/v3/webauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Adapter entity
type Adapter struct {
	host string
	port string
	auth *Auth

	apisHandler      rest.ApisHandler
	adminApisHandler rest.AdminApisHandler

	app *core.Application

	corsAllowedOrigins []string
	corsAllowedHeaders []string

	logger *logs.Logger
}

// @title Rokwire Content Building Block API
// @description Rokwire Content Building Block API Documentation.
// @version 1.2.3
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost
// @BasePath /content
// @schemes https

// @securityDefinitions.apikey RokwireAuth
// @in header
// @name ROKWIRE-API-KEY

// @securityDefinitions.apikey UserAuth
// @in header (add Bearer prefix to the Authorization value)
// @name Authorization

// @securityDefinitions.apikey AdminUserAuth
// @in header (add Bearer prefix to the Authorization value)
// @name Authorization

// @securityDefinitions.apikey AdminGroupAuth
// @in header
// @name GROUP

// Start starts the module
func (we Adapter) Start() {

	router := mux.NewRouter().StrictSlash(true)

	// handle apis
	contentRouter := router.PathPrefix("/content").Subrouter()
	contentRouter.PathPrefix("/doc/ui").Handler(we.serveDocUI())
	contentRouter.HandleFunc("/doc", we.serveDoc)
	contentRouter.HandleFunc("/version", we.wrapFunc(we.apisHandler.Version)).Methods("GET")

	contentRouter.HandleFunc("/profile_photo/{user-id}", we.coreAuthWrapFunc(we.apisHandler.GetProfilePhoto, we.auth.coreAuth.userAuth)).Methods("GET")
	contentRouter.HandleFunc("/profile_photo", we.coreAuthWrapFunc(we.apisHandler.GetUserProfilePhoto, we.auth.coreAuth.userAuth)).Methods("GET")
	contentRouter.HandleFunc("/profile_photo", we.coreAuthWrapFunc(we.apisHandler.StoreProfilePhoto, we.auth.coreAuth.userAuth)).Methods("POST")
	contentRouter.HandleFunc("/profile_photo", we.coreAuthWrapFunc(we.apisHandler.DeleteProfilePhoto, we.auth.coreAuth.userAuth)).Methods("DELETE")

	// handle student guide client apis
	contentRouter.HandleFunc("/student_guides", we.coreAuthWrapFunc(we.apisHandler.GetStudentGuides, we.auth.coreAuth.standardAuth)).Methods("GET")
	contentRouter.HandleFunc("/student_guides/{id}", we.coreAuthWrapFunc(we.apisHandler.GetStudentGuide, we.auth.coreAuth.standardAuth)).Methods("GET")
	contentRouter.HandleFunc("/health_locations", we.coreAuthWrapFunc(we.apisHandler.GetHealthLocations, we.auth.coreAuth.standardAuth)).Methods("GET")
	contentRouter.HandleFunc("/health_locations/{id}", we.coreAuthWrapFunc(we.apisHandler.GetHealthLocation, we.auth.coreAuth.standardAuth)).Methods("GET")
	contentRouter.HandleFunc("/content_items", we.coreAuthWrapFunc(we.apisHandler.GetContentItems, we.auth.coreAuth.standardAuth)).Methods("GET")
	contentRouter.HandleFunc("/content_items/{id}", we.coreAuthWrapFunc(we.apisHandler.GetContentItem, we.auth.coreAuth.standardAuth)).Methods("GET")
	contentRouter.HandleFunc("/content_item/categories", we.coreAuthWrapFunc(we.apisHandler.GetContentItemsCategories, we.auth.coreAuth.standardAuth)).Methods("GET")
	contentRouter.HandleFunc("/image", we.coreAuthWrapFunc(we.apisHandler.UploadImage, we.auth.coreAuth.userAuth)).Methods("POST")
	contentRouter.HandleFunc("/twitter/users/{user_id}/tweets", we.coreAuthWrapFunc(we.apisHandler.GetTweeterPosts, we.auth.coreAuth.standardAuth)).Methods("GET")

	// handle student guide admin apis
	adminSubRouter := contentRouter.PathPrefix("/admin").Subrouter()

	//deprecated
	adminSubRouter.HandleFunc("/student_guides", we.coreAuthWrapFunc(we.adminApisHandler.GetStudentGuides, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/student_guides", we.coreAuthWrapFunc(we.adminApisHandler.CreateStudentGuide, we.auth.coreAuth.permissionsAuth)).Methods("POST")
	adminSubRouter.HandleFunc("/student_guides/{id}", we.coreAuthWrapFunc(we.adminApisHandler.GetStudentGuide, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/student_guides/{id}", we.coreAuthWrapFunc(we.adminApisHandler.UpdateStudentGuide, we.auth.coreAuth.permissionsAuth)).Methods("PUT")
	adminSubRouter.HandleFunc("/student_guides/{id}", we.coreAuthWrapFunc(we.adminApisHandler.DeleteStudentGuide, we.auth.coreAuth.permissionsAuth)).Methods("DELETE")
	//end deprecated

	//deprecated
	adminSubRouter.HandleFunc("/health_locations", we.coreAuthWrapFunc(we.adminApisHandler.GetHealthLocations, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/health_locations", we.coreAuthWrapFunc(we.adminApisHandler.CreateHealthLocation, we.auth.coreAuth.permissionsAuth)).Methods("POST")
	adminSubRouter.HandleFunc("/health_location/{id}", we.coreAuthWrapFunc(we.adminApisHandler.GetHealthLocation, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/health_location/{id}", we.coreAuthWrapFunc(we.adminApisHandler.UpdateHealthLocation, we.auth.coreAuth.permissionsAuth)).Methods("PUT")
	adminSubRouter.HandleFunc("/health_location/{id}", we.coreAuthWrapFunc(we.adminApisHandler.DeleteHealthLocation, we.auth.coreAuth.permissionsAuth)).Methods("DELETE")
	//end deprecated

	adminSubRouter.HandleFunc("/v2/health_locations", we.coreAuthWrapFunc(we.adminApisHandler.GetHealthLocationsV2, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/v2/health_locations", we.coreAuthWrapFunc(we.adminApisHandler.CreateHealthLocationV2, we.auth.coreAuth.permissionsAuth)).Methods("POST")
	adminSubRouter.HandleFunc("/v2/health_locations/{id}", we.coreAuthWrapFunc(we.adminApisHandler.UpdateHealthLocationV2, we.auth.coreAuth.permissionsAuth)).Methods("PUT")
	adminSubRouter.HandleFunc("/v2/health_locations/{id}", we.coreAuthWrapFunc(we.adminApisHandler.DeleteHealthLocationV2, we.auth.coreAuth.permissionsAuth)).Methods("DELETE")

	adminSubRouter.HandleFunc("/v2/student_guides", we.coreAuthWrapFunc(we.adminApisHandler.GetStudentGuidesV2, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/v2/student_guides", we.coreAuthWrapFunc(we.adminApisHandler.CreateStudentGuidesV2, we.auth.coreAuth.permissionsAuth)).Methods("POST")
	adminSubRouter.HandleFunc("/v2/student_guides/{id}", we.coreAuthWrapFunc(we.adminApisHandler.UpdateStudentGuidesV2, we.auth.coreAuth.permissionsAuth)).Methods("PUT")
	adminSubRouter.HandleFunc("/v2/student_guides/{id}", we.coreAuthWrapFunc(we.adminApisHandler.DeleteStudentGuidesV2, we.auth.coreAuth.permissionsAuth)).Methods("DELETE")

	adminSubRouter.HandleFunc("/wellness_tips", we.coreAuthWrapFunc(we.adminApisHandler.GetWellnessTips, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/wellness_tips", we.coreAuthWrapFunc(we.adminApisHandler.CreateWellnessTips, we.auth.coreAuth.permissionsAuth)).Methods("POST")
	adminSubRouter.HandleFunc("/wellness_tips/{id}", we.coreAuthWrapFunc(we.adminApisHandler.UpdateWellnessTips, we.auth.coreAuth.permissionsAuth)).Methods("PUT")
	adminSubRouter.HandleFunc("/wellness_tips/{id}", we.coreAuthWrapFunc(we.adminApisHandler.DeleteWellnessTips, we.auth.coreAuth.permissionsAuth)).Methods("DELETE")

	adminSubRouter.HandleFunc("/campus_reminders", we.coreAuthWrapFunc(we.adminApisHandler.GetCampusReminders, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/campus_reminders", we.coreAuthWrapFunc(we.adminApisHandler.CreateCampusReminder, we.auth.coreAuth.permissionsAuth)).Methods("POST")
	adminSubRouter.HandleFunc("/campus_reminders/{id}", we.coreAuthWrapFunc(we.adminApisHandler.UpdateCampusReminder, we.auth.coreAuth.permissionsAuth)).Methods("PUT")
	adminSubRouter.HandleFunc("/campus_reminders/{id}", we.coreAuthWrapFunc(we.adminApisHandler.DeleteCampusReminder, we.auth.coreAuth.permissionsAuth)).Methods("DELETE")

	adminSubRouter.HandleFunc("/gies_onboarding_checklists", we.coreAuthWrapFunc(we.adminApisHandler.GetGiesOnboardingChecklists, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/gies_onboarding_checklists", we.coreAuthWrapFunc(we.adminApisHandler.CreateGiesOnboardingChecklist, we.auth.coreAuth.permissionsAuth)).Methods("POST")
	adminSubRouter.HandleFunc("/gies_onboarding_checklists/{id}", we.coreAuthWrapFunc(we.adminApisHandler.UpdateGiesOnboardingChecklist, we.auth.coreAuth.permissionsAuth)).Methods("PUT")
	adminSubRouter.HandleFunc("/gies_onboarding_checklists/{id}", we.coreAuthWrapFunc(we.adminApisHandler.DeleteGiesOnboardingChecklist, we.auth.coreAuth.permissionsAuth)).Methods("DELETE")

	adminSubRouter.HandleFunc("/uiuc_onboarding_checklists", we.coreAuthWrapFunc(we.adminApisHandler.GetUIUCOnboardingChecklists, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/uiuc_onboarding_checklists", we.coreAuthWrapFunc(we.adminApisHandler.CreateUIUCOnboardingChecklist, we.auth.coreAuth.permissionsAuth)).Methods("POST")
	adminSubRouter.HandleFunc("/uiuc_onboarding_checklists/{id}", we.coreAuthWrapFunc(we.adminApisHandler.UpdateUIUCOnboardingChecklist, we.auth.coreAuth.permissionsAuth)).Methods("PUT")
	adminSubRouter.HandleFunc("/uiuc_onboarding_checklists/{id}", we.coreAuthWrapFunc(we.adminApisHandler.DeleteUIUCOnboardingChecklist, we.auth.coreAuth.permissionsAuth)).Methods("DELETE")

	adminSubRouter.HandleFunc("/gies_post_templates", we.coreAuthWrapFunc(we.adminApisHandler.GetGiesPostTemplates, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/gies_post_templates", we.coreAuthWrapFunc(we.adminApisHandler.CreateGiesPostTemplate, we.auth.coreAuth.permissionsAuth)).Methods("POST")
	adminSubRouter.HandleFunc("/gies_post_templates/{id}", we.coreAuthWrapFunc(we.adminApisHandler.UpdateGiesPostTemplate, we.auth.coreAuth.permissionsAuth)).Methods("PUT")
	adminSubRouter.HandleFunc("/gies_post_templates/{id}", we.coreAuthWrapFunc(we.adminApisHandler.DeleteGiesPostTemplate, we.auth.coreAuth.permissionsAuth)).Methods("DELETE")

	adminSubRouter.HandleFunc("/content_items", we.coreAuthWrapFunc(we.adminApisHandler.GetContentItems, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/content_items", we.coreAuthWrapFunc(we.adminApisHandler.CreateContentItem, we.auth.coreAuth.permissionsAuth)).Methods("POST")
	adminSubRouter.HandleFunc("/content_items/{id}", we.coreAuthWrapFunc(we.adminApisHandler.GetContentItem, we.auth.coreAuth.permissionsAuth)).Methods("GET")
	adminSubRouter.HandleFunc("/content_items/{id}", we.coreAuthWrapFunc(we.adminApisHandler.UpdateContentItem, we.auth.coreAuth.permissionsAuth)).Methods("PUT")
	adminSubRouter.HandleFunc("/content_items/{id}", we.coreAuthWrapFunc(we.adminApisHandler.DeleteContentItem, we.auth.coreAuth.permissionsAuth)).Methods("DELETE")
	adminSubRouter.HandleFunc("/content_item/categories", we.coreAuthWrapFunc(we.adminApisHandler.GetContentItemsCategories, we.auth.coreAuth.permissionsAuth)).Methods("GET")

	adminSubRouter.HandleFunc("/image", we.coreAuthWrapFunc(we.adminApisHandler.UploadImage, we.auth.coreAuth.permissionsAuth)).Methods("POST")

	var handler http.Handler = router
	if len(we.corsAllowedOrigins) > 0 {
		handler = webauth.SetupCORS(we.corsAllowedOrigins, we.corsAllowedHeaders, router)
	}
	we.logger.Fatalf("Error serving: %v", http.ListenAndServe(":"+we.port, handler))
}

func (we Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")
	http.ServeFile(w, r, "./docs/swagger.yaml")
}

func (we Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/content/doc", we.host)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

func (we Adapter) wrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		handler(w, req)
	}
}

type coreAuthFunc = func(*tokenauth.Claims, http.ResponseWriter, *http.Request)

func (we Adapter) coreAuthWrapFunc(handler coreAuthFunc, authorization Authorization) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		responseStatus, claims, err := authorization.check(req)
		if err != nil {
			log.Printf("error authorization check - %s", err)
			http.Error(w, http.StatusText(responseStatus), responseStatus)
			return
		}
		handler(claims, w, req)
	}
}

// NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(host string, port string, app *core.Application, serviceRegManager *authservice.ServiceRegManager,
	corsAllowedOrigins []string, corsAllowedHeaders []string, logger *logs.Logger) Adapter {
	auth := NewAuth(app, serviceRegManager)

	apisHandler := rest.NewApisHandler(app)
	adminApisHandler := rest.NewAdminApisHandler(app)
	return Adapter{host: host, port: port, auth: auth, apisHandler: apisHandler, adminApisHandler: adminApisHandler, app: app,
		corsAllowedOrigins: corsAllowedOrigins, corsAllowedHeaders: corsAllowedHeaders, logger: logger}
}

// AppListener implements core.ApplicationListener interface
type AppListener struct {
	adapter *Adapter
}
