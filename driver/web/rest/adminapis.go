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

package rest

import (
	"content/core"
	"content/core/model"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"go.mongodb.org/mongo-driver/bson"
)

// AdminApisHandler handles the rest Admin APIs implementation
type AdminApisHandler struct {
	app *core.Application
}

// GetStudentGuides Retrieves  all student guides
// @Description Retrieves  all student guides
// @Param ids query string false "Coma separated IDs of the desired records"
// @Tags Admin
// @ID AdminGetStudentGuides
// @Accept json
// @Success 200
// @Security AdminUserAuth
// @Deprecated true
// @Router /admin/student_guides [get]
func (h AdminApisHandler) GetStudentGuides(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {

	IDs := []string{}
	IDskeys, ok := r.URL.Query()["ids"]
	if ok && len(IDskeys[0]) > 0 {
		extIDs := IDskeys[0]
		IDs = strings.Split(extIDs, ",")
	}

	resData, err := h.app.Services.GetStudentGuides(claims.AppID, claims.OrgID, IDs)
	if err != nil {
		log.Printf("Error on getting guide items by id - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resData == nil {
		resData = []bson.M{}
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal all student guides")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetStudentGuide Retrieves a student guide by id
// @Description Retrieves  all items
// @Tags Admin
// @ID AdminGetStudentGuide
// @Accept json
// @Produce json
// @Success 200
// @Security AdminUserAuth
// @Deprecated true
// @Router /admin/student_guides/{id} [get]
func (h AdminApisHandler) GetStudentGuide(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guideID := vars["id"]

	resData, err := h.app.Services.GetStudentGuide(claims.AppID, claims.OrgID, guideID)
	if err != nil {
		log.Printf("Error on getting student guide id - %s\n %s", guideID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal the student guide")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// UpdateStudentGuide Updates a student guide with the specified id
// @Description Updates a student guide with the specified id
// @Tags Admin
// @ID AdminUpdateStudentGuide
// @Accept json
// @Produce json
// @Success 200
// @Security AdminUserAuth
// @Deprecated true
// @Router /admin/student_guides/{id} [put]
func (h AdminApisHandler) UpdateStudentGuide(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guideID := vars["id"]

	var item bson.M
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the create student guide request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resData, err := h.app.Services.UpdateStudentGuide(claims.AppID, claims.OrgID, guideID, item)
	if err != nil {
		log.Printf("Error on updating student guide with id - %s\n %s", guideID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal the updated student guide")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// CreateStudentGuide Creates a student guide item
// @Description Creates a student guide item
// @Tags Admin
// @ID AdminCreateStudentGuide
// @Accept json
// @Success 200
// @Security AdminUserAuth
// @Deprecated true
// @Router /admin/student_guides [post]
func (h AdminApisHandler) CreateStudentGuide(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	var item bson.M
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the create student guide request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdItem, err := h.app.Services.CreateStudentGuide(claims.AppID, claims.OrgID, item)
	if err != nil {
		log.Printf("Error on creating student guide: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(createdItem)
	if err != nil {
		log.Println("Error on marshal the new item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// DeleteStudentGuide Deletes a student guide item with the specified id
// @Description Deletes a student guide item with the specified id
// @Tags Admin
// @ID AdminDeleteStudentGuide
// @Success 200
// @Security AdminUserAuth
// @Deprecated true
// @Router /admin/student_guides/{id} [delete]
func (h AdminApisHandler) DeleteStudentGuide(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guideID := vars["id"]

	err := h.app.Services.DeleteStudentGuide(claims.AppID, claims.OrgID, guideID)
	if err != nil {
		log.Printf("Error on deleting student guide with id - %s\n %s", guideID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// GetHealthLocations Retrieves  all health locations
// @Description Retrieves  all health locations
// @Param ids query string false "Coma separated IDs of the desired records"
// @Tags Admin
// @ID AdminGetHealthLocations
// @Accept json
// @Success 200
// @Security AdminUserAuth
// @Deprecated true
// @Router /admin/health_locations [get]
func (h AdminApisHandler) GetHealthLocations(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {

	IDs := []string{}
	IDskeys, ok := r.URL.Query()["ids"]
	if ok && len(IDskeys[0]) > 0 {
		extIDs := IDskeys[0]
		IDs = strings.Split(extIDs, ",")
	}

	resData, err := h.app.Services.GetHealthLocations(claims.AppID, claims.OrgID, IDs)
	if err != nil {
		log.Printf("Error on health location items by id - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resData == nil {
		resData = []bson.M{}
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal all health locations")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetHealthLocation Retrieves a health location by id
// @Description Retrieves a health location by id
// @Tags Admin
// @ID AdminGetHealthLocation
// @Accept json
// @Produce json
// @Success 200
// @Security AdminUserAuth
// @Deprecated true
// @Router /admin/health_locations/{id} [get]
func (h AdminApisHandler) GetHealthLocation(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["id"]

	resData, err := h.app.Services.GetHealthLocation(claims.AppID, claims.OrgID, locationID)
	if err != nil {
		log.Printf("Error on getting health location id - %s\n %s", locationID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal the health location")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// UpdateHealthLocation Updates a health location with the specified id
// @Description Updates a health location with the specified id
// @Tags Admin
// @ID AdminUpdateHealthLocation
// @Accept json
// @Produce json
// @Success 200
// @Security AdminUserAuth
// @Deprecated true
// @Router /admin/health_locations/{id} [put]
func (h AdminApisHandler) UpdateHealthLocation(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["id"]

	var item bson.M
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the create health location request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resData, err := h.app.Services.UpdateHealthLocation(claims.AppID, claims.OrgID, locationID, item)
	if err != nil {
		log.Printf("Error on updating health location with id - %s\n %s", locationID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal the updated health location")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// CreateHealthLocation Create a new health location
// @Description Create a new health location
// @Tags Admin
// @ID AdminCreateHealthLocation
// @Accept json
// @Success 200
// @Security AdminUserAuth
// @Deprecated true
// @Router /admin/health_locations [post]
func (h AdminApisHandler) CreateHealthLocation(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	var item bson.M
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the create health location request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdItem, err := h.app.Services.CreateHealthLocation(claims.AppID, claims.OrgID, item)
	if err != nil {
		log.Printf("Error on creating health location: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(createdItem)
	if err != nil {
		log.Println("Error on marshal the new item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// DeleteHealthLocation Deletes a health location with the specified id
// @Description Deletes a health location with the specified id
// @Tags Admin
// @ID AdminDeleteHealthLocation
// @Success 200
// @Security AdminUserAuth
// @Deprecated true
// @Router /admin/health_location/{id} [delete]
func (h AdminApisHandler) DeleteHealthLocation(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	locationID := vars["id"]

	err := h.app.Services.DeleteHealthLocation(claims.AppID, claims.OrgID, locationID)
	if err != nil {
		log.Printf("Error on deleting health location with id - %s\n %s", locationID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// GetHealthLocationsV2 Retrieves health locations
// @Description Retrieves Retrieves health locations
// @Tags Admin
// @ID AdminGetHealthLocationsV2
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Param ids query string false "Comma separated IDs of the desired records"
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Accept json
// @Success 200 {array} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/v2/health_locations [get]
func (h AdminApisHandler) GetHealthLocationsV2(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.getContentItemsByCategory(claims, w, r, "health_locations")
}

// CreateHealthLocationV2 creates a new health location. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Creates a new health location. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminCreateHealthLocationV2
// @Param data body createContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/v2/health_locations [post]
func (h AdminApisHandler) CreateHealthLocationV2(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.createContentItemByCategory(claims, w, r, "health_locations")
}

// UpdateHealthLocationV2 Updates a health location with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Updates a health location with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminUpdateHealthLocationV2
// @Param data body updateContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Produce json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/v2/health_locations/{id} [put]
func (h AdminApisHandler) UpdateHealthLocationV2(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.updateContentItemByCategory(claims, w, r, "health_locations")
}

// DeleteHealthLocationV2 Deletes a health location with the specified id
// @Description Deletes a health location with the specified id
// @Tags Admin
// @ID AdminDeleteHealthLocationV2
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Success 200
// @Security AdminUserAuth
// @Router /admin/v2/health_locations/{id} [delete]
func (h AdminApisHandler) DeleteHealthLocationV2(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.deleteContentItemByCategory(claims, w, r, "health_locations")
}

// GetStudentGuidesV2 Retrieves student guides
// @Description Retrieves student guides
// @Tags Admin
// @ID AdminGetStudentGuidesV2
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Param ids query string false "Comma separated IDs of the desired records"
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Accept json
// @Success 200 {array} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/v2/student_guides [get]
func (h AdminApisHandler) GetStudentGuidesV2(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.getContentItemsByCategory(claims, w, r, "student_guides")
}

// CreateStudentGuidesV2 creates a new student guide. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Creates a new student guide. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminCreateStudentGuidesV2
// @Param data body createContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/v2/student_guides [post]
func (h AdminApisHandler) CreateStudentGuidesV2(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.createContentItemByCategory(claims, w, r, "student_guides")
}

// UpdateStudentGuidesV2 Updates a student guide with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Updates a student guide with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminUpdateStudentGuidesV2
// @Param data body updateContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Produce json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/v2/student_guides/{id} [put]
func (h AdminApisHandler) UpdateStudentGuidesV2(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.updateContentItemByCategory(claims, w, r, "student_guides")
}

// DeleteStudentGuidesV2 Deletes a student guide with the specified id
// @Description Deletes a student guide with the specified id
// @Tags Admin
// @ID AdminDeleteStudentGuidesV2
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Success 200
// @Security AdminUserAuth
// @Router /admin/v2/student_guides/{id} [delete]
func (h AdminApisHandler) DeleteStudentGuidesV2(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.deleteContentItemByCategory(claims, w, r, "student_guides")
}

// GetWellnessTips Retrieves wellness tip item
// @Description Retrieves wellness tip items
// @Tags Admin
// @ID AdminGetWellnessTip
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Param ids query string false "Comma separated IDs of the desired records"
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Accept json
// @Success 200 {array} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/wellness_tips [get]
func (h AdminApisHandler) GetWellnessTips(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.getContentItemsByCategory(claims, w, r, "wellness_tips")
}

// CreateWellnessTips creates a new wellness tip. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Creates a new wellness tip. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminCreateWellnessTip
// @Param data body createContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/wellness_tips [post]
func (h AdminApisHandler) CreateWellnessTips(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.createContentItemByCategory(claims, w, r, "wellness_tips")
}

// UpdateWellnessTips Updates a wellness tip with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Updates a wellness tip with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminUpdateWellnessTip
// @Param data body updateContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Produce json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/wellness_tips/{id} [put]
func (h AdminApisHandler) UpdateWellnessTips(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.updateContentItemByCategory(claims, w, r, "wellness_tips")
}

// DeleteWellnessTips Deletes a wellness tip with the specified id
// @Description Deletes a wellness tip with the specified id
// @Tags Admin
// @ID AdminDeleteWellnessTip
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Success 200
// @Security AdminUserAuth
// @Router /admin/wellness_tips/{id} [delete]
func (h AdminApisHandler) DeleteWellnessTips(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.deleteContentItemByCategory(claims, w, r, "wellness_tips")
}

// GetCampusReminders Retrieves campus reminders
// @Description Retrieves campus reminders
// @Tags Admin
// @ID AdminGetCampusReminders
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Param ids query string false "Comma separated IDs of the desired records"
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Accept json
// @Success 200 {array} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/campus_reminders [get]
func (h AdminApisHandler) GetCampusReminders(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.getContentItemsByCategory(claims, w, r, "campus_reminders")
}

// CreateCampusReminder creates a new campus reminder. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Creates a new campus reminder. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminCreateCampusReminder
// @Param data body createContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/campus_reminders [post]
func (h AdminApisHandler) CreateCampusReminder(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.createContentItemByCategory(claims, w, r, "campus_reminders")
}

// UpdateCampusReminder Updates a campus reminder with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Updates a campus reminder with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminCampusReminder
// @Param data body updateContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Produce json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/campus_reminders/{id} [put]
func (h AdminApisHandler) UpdateCampusReminder(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.updateContentItemByCategory(claims, w, r, "campus_reminders")
}

// DeleteCampusReminder Deletes a campus reminder with the specified id
// @Description Deletes a campus reminder with the specified id
// @Tags Admin
// @ID AdminDeleteCampusReminder
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Success 200
// @Security AdminUserAuth
// @Router /admin/campus_reminders/{id} [delete]
func (h AdminApisHandler) DeleteCampusReminder(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.deleteContentItemByCategory(claims, w, r, "campus_reminders")
}

// GetGiesOnboardingChecklists Retrieves gies onboarding checklists
// @Description Retrieves gies onboarding checklists
// @Tags Admin
// @ID AdminGetGiesOnboardingChecklists
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Param ids query string false "Comma separated IDs of the desired records"
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Accept json
// @Success 200 {array} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/gies_onboarding_checklists [get]
func (h AdminApisHandler) GetGiesOnboardingChecklists(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.getContentItemsByCategory(claims, w, r, "gies_onboarding_checklists")
}

// CreateGiesOnboardingChecklist creates a new gies onboarding checklist. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Creates a new gies onboarding checklist. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminCreateGiesOnboardingChecklist
// @Param data body createContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/gies_onboarding_checklists [post]
func (h AdminApisHandler) CreateGiesOnboardingChecklist(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.createContentItemByCategory(claims, w, r, "gies_onboarding_checklists")
}

// UpdateGiesOnboardingChecklist Updates a gies onboarding checklist with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Updates a gies onboarding checklist with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminUpdateGiesOnboardingChecklist
// @Param data body updateContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Produce json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/gies_onboarding_checklists/{id} [put]
func (h AdminApisHandler) UpdateGiesOnboardingChecklist(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.updateContentItemByCategory(claims, w, r, "gies_onboarding_checklists")
}

// DeleteGiesOnboardingChecklist Deletes a gies onboarding checklist with the specified id
// @Description Deletes a gies onboarding checklist with the specified id
// @Tags Admin
// @ID AdminDeleteGiesOnboardingChecklist
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Success 200
// @Security AdminUserAuth
// @Router /admin/gies_onboarding_checklists/{id} [delete]
func (h AdminApisHandler) DeleteGiesOnboardingChecklist(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.deleteContentItemByCategory(claims, w, r, "gies_onboarding_checklists")
}

// GetUIUCOnboardingChecklists Retrieves uiuc onboarding checklist items
// @Description Retrieves uiuc onboarding checklist items
// @Tags Admin
// @ID AdminGetUIUCOnboardingChecklists
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Param ids query string false "Comma separated IDs of the desired records"
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Accept json
// @Success 200 {array} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/uiuc_onboarding_checklists [get]
func (h AdminApisHandler) GetUIUCOnboardingChecklists(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.getContentItemsByCategory(claims, w, r, "uiuc_onboarding_checklists")
}

// CreateUIUCOnboardingChecklist creates a new uiuc onboarding checklist. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Creates a new uiuc onboarding checklist. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminCreateUIUCOnboardingChecklist
// @Param data body createContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/uiuc_onboarding_checklists [post]
func (h AdminApisHandler) CreateUIUCOnboardingChecklist(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.createContentItemByCategory(claims, w, r, "uiuc_onboarding_checklists")
}

// UpdateUIUCOnboardingChecklist Updates a uiuc onboarding checklist with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Updates a uiuc onboarding checklist with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminUpdateUIUCOnboardingChecklist
// @Param data body updateContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Produce json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/uiuc_onboarding_checklists/{id} [put]
func (h AdminApisHandler) UpdateUIUCOnboardingChecklist(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.updateContentItemByCategory(claims, w, r, "uiuc_onboarding_checklists")
}

// DeleteUIUCOnboardingChecklist Deletes a uiuc onboarding checklist with the specified id
// @Description Deletes a uiuc onboarding checklist with the specified id
// @Tags Admin
// @ID AdminDeleteUIUCOnboardingChecklist
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Success 200
// @Security AdminUserAuth
// @Router /admin/uiuc_onboarding_checklists/{id} [delete]
func (h AdminApisHandler) DeleteUIUCOnboardingChecklist(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.deleteContentItemByCategory(claims, w, r, "uiuc_onboarding_checklists")
}

// GetGiesPostTemplates Retrieves gies post template items
// @Description Retrieves gies post template items
// @Tags Admin
// @ID AdminGiesPostTemplates
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Param ids query string false "Comma separated IDs of the desired records"
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Accept json
// @Success 200 {array} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/gies_post_templates [get]
func (h AdminApisHandler) GetGiesPostTemplates(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.getContentItemsByCategory(claims, w, r, "gies_post_templates")
}

// CreateGiesPostTemplate creates a new gies post template. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Creates a new gies post template. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminGiesPostTemplate
// @Param data body createContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/gies_post_templates [post]
func (h AdminApisHandler) CreateGiesPostTemplate(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.createContentItemByCategory(claims, w, r, "gies_post_templates")
}

// UpdateGiesPostTemplate Updates a gies post template with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Updates a gies post template with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminUpdateGiesPostTemplate
// @Param data body updateContentItemByCategoryRequestBody true "Params"
// @Accept json
// @Produce json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/gies_post_templates/{id} [put]
func (h AdminApisHandler) UpdateGiesPostTemplate(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.updateContentItemByCategory(claims, w, r, "gies_post_templates")
}

// DeleteGiesPostTemplate Deletes a gies post template with the specified id
// @Description Deletes a gies post template with the specified id
// @Tags Admin
// @ID AdminGiesPostTemplate
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Success 200
// @Security AdminUserAuth
// @Router /admin/gies_post_templates/{id} [delete]
func (h AdminApisHandler) DeleteGiesPostTemplate(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	h.deleteContentItemByCategory(claims, w, r, "gies_post_templates")
}

func (h AdminApisHandler) getContentItemsByCategory(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request, category string) {
	//get all-apps param value
	allApps := false //false by defautl
	allAppsParam := r.URL.Query().Get("all-apps")
	if allAppsParam != "" {
		allApps, _ = strconv.ParseBool(allAppsParam)
	}

	IDs := []string{}
	IDskeys, ok := r.URL.Query()["ids"]
	if ok && len(IDskeys[0]) > 0 {
		extIDs := IDskeys[0]
		IDs = strings.Split(extIDs, ",")
	}

	var offset *int64
	offsets, ok := r.URL.Query()["offset"]
	if ok && len(offsets[0]) > 0 {
		val, err := strconv.ParseInt(offsets[0], 0, 64)
		if err == nil {
			offset = &val
		}
	}

	var limit *int64
	limits, ok := r.URL.Query()["limit"]
	if ok && len(limits[0]) > 0 {
		val, err := strconv.ParseInt(limits[0], 0, 64)
		if err == nil {
			limit = &val
		}
	}

	var order *string
	orders, ok := r.URL.Query()["order"]
	if ok && len(orders[0]) > 0 {
		order = &orders[0]
	}

	categories := []string{category}

	resData, err := h.app.Services.GetContentItems(allApps, claims.AppID, claims.OrgID, IDs, categories, offset, limit, order)
	if err != nil {
		log.Printf("Error on cgetting content items - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resData == nil {
		resData = []model.ContentItemResponse{}
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// createContentItemByCategoryRequestBody Expected body while creating a new content item
type createContentItemByCategoryRequestBody struct {
	AllApps bool        `json:"all_apps"`
	Data    interface{} `json:"data" bson:"data"`
} // @name createContentItemByCategoryRequestBody

func (h AdminApisHandler) createContentItemByCategory(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request, category string) {
	var item createContentItemByCategoryRequestBody
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the create content item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdItem, err := h.app.Services.CreateContentItem(item.AllApps, claims.AppID, claims.OrgID, category, item.Data)
	if err != nil {
		log.Printf("Error on creating content item: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(createdItem)
	if err != nil {
		log.Println("Error on marshal the new content item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// updateContentItemByCategoryRequestBody Expected body while updating a content item
type updateContentItemByCategoryRequestBody struct {
	AllApps bool        `json:"all_apps"`
	Data    interface{} `json:"data"`
} // @name updateContentItemByCategoryRequestBody

func (h AdminApisHandler) updateContentItemByCategory(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request, category string) {
	vars := mux.Vars(r)
	id := vars["id"]

	var item updateContentItemByCategoryRequestBody
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the update content item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if item.Data == nil {
		log.Printf("Unable to update content item: Missing data")
		http.Error(w, "Unable to update content item: Missing data", http.StatusBadRequest)
		return
	}

	resData, err := h.app.Services.UpdateContentItemData(item.AllApps, claims.AppID, claims.OrgID, id, category, item.Data)
	if err != nil {
		log.Printf("Error on updating content item with id - %s\n %s", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal the updated content item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h AdminApisHandler) deleteContentItemByCategory(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request, category string) {
	//get all-apps param value
	allApps := false //false by defautl
	allAppsParam := r.URL.Query().Get("all-apps")
	if allAppsParam != "" {
		allApps, _ = strconv.ParseBool(allAppsParam)
	}

	vars := mux.Vars(r)
	id := vars["id"]

	err := h.app.Services.DeleteContentItemByCategory(allApps, claims.AppID, claims.OrgID, id, category)
	if err != nil {
		log.Printf("Error on deleting content item with id - %s\n %s", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// uploadImageResponse wrapper
type uploadImageResponse struct {
	URL string `json:"url"`
} // @name uploadImageResponse

// UploadImage Uploads an image to AWS S3
// @Description Uploads an image to AWS S3
// @Tags Admin
// @ID AdminUploadImage
// @Param path body string true "path - path within the S3 bucket"
// @Param width body string false "width - width of the image to resize. If width and height are missing - then the new image will use the original size"
// @Param height body string false "height - height of the image to resize. If width and height are missing - then the new image will use the original size"
// @Param quality body string false "quality - quality of the image. Default: 90"
// @Param fileName body string false "fileName - the uploaded file name"
// @Accept multipart/form-data
// @Produce json
// @Success 200 {object} uploadImageResponse
// @Security AdminUserAuth
// @Router /admin/image [post]
func (h AdminApisHandler) UploadImage(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	//validate the image type
	path := r.PostFormValue("path")
	if len(path) <= 0 {
		log.Print("Missing image path\n")
		http.Error(w, "missing 'path' form param", http.StatusBadRequest)
		return
	}

	heightParam := intPostValueFromString(r.PostFormValue("height"))
	widthParam := intPostValueFromString(r.PostFormValue("width"))
	qualityParam := intPostValueFromString(r.PostFormValue("quality"))
	imgSpec := model.ImageSpec{Height: heightParam, Width: widthParam, Quality: qualityParam}

	// validate file size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.Print("File is too big\n")
		http.Error(w, "File is too big", http.StatusBadRequest)
		return
	}

	// parse and validate file and post parameters
	file, _, err := r.FormFile("fileName")
	if err != nil {
		log.Print("Invalid file\n")
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Print("Invalid file\n")
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}

	// check file type, detectcontenttype only needs the first 512 bytes
	filetype := http.DetectContentType(fileBytes)
	switch filetype {
	case "image/jpeg", "image/jpg":
	case "image/gif", "image/png":
	case "image/webp":
		break
	default:
		log.Print("Invalid file type\n")
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// pass the file to be processed by the use case handler
	url, err := h.app.Services.UploadImage(fileBytes, path, imgSpec)
	if err != nil {
		log.Printf("Error converting image: %s\n", err)
		http.Error(w, "Error converting image", http.StatusInternalServerError)
		return
	}

	jsonData := map[string]string{"url": *url}
	jsonBynaryData, err := json.Marshal(jsonData)
	if err != nil {
		log.Println("Error on marshal s3 location data")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBynaryData)
}

type getContentItemsRequestBody struct {
	IDs        []string `json:"ids,omitempty"`        // List of IDs for the filter. Optional and may be null or missing.
	Categories []string `json:"categories,omitempty"` // List of Categories for the filter. Optional and may be null or missing.
} // @name getContentItemsRequestBody

// GetContentItems Retrieves  all content items. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Retrieves  all content items.<b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminGetContentItems
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Param data body getContentItemsRequestBody false "Optional - body json of the all items ids that need to be filtered. NOTE: Bad/broken json will be interpreted as an empty filter and the request will be proceeded further."
// @Accept json
// @Success 200 {array} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/content_items [get]
func (h AdminApisHandler) GetContentItems(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	//get all-apps param value
	allApps := false //false by defautl
	allAppsParam := r.URL.Query().Get("all-apps")
	if allAppsParam != "" {
		allApps, _ = strconv.ParseBool(allAppsParam)
	}

	var offset *int64
	offsets, ok := r.URL.Query()["offset"]
	if ok && len(offsets[0]) > 0 {
		val, err := strconv.ParseInt(offsets[0], 0, 64)
		if err == nil {
			offset = &val
		}
	}

	var limit *int64
	limits, ok := r.URL.Query()["limit"]
	if ok && len(limits[0]) > 0 {
		val, err := strconv.ParseInt(limits[0], 0, 64)
		if err == nil {
			limit = &val
		}
	}

	var order *string
	orders, ok := r.URL.Query()["order"]
	if ok && len(orders[0]) > 0 {
		order = &orders[0]
	}

	var item getContentItemsRequestBody
	bodyErr := json.NewDecoder(r.Body).Decode(&item)
	if bodyErr != nil {
		log.Printf("Warning: bad getContentItemsRequestBody request: %s", bodyErr)
	}

	resData, err := h.app.Services.GetContentItems(allApps, claims.AppID, claims.OrgID, item.IDs, item.Categories, offset, limit, order)
	if err != nil {
		log.Printf("Error on cgetting content items - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resData == nil {
		resData = []model.ContentItemResponse{}
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal all content items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetContentItem Retrieves a content item by id. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Retrieves a content item by id. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminGetContentItem
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Accept json
// @Produce json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/content_items/{id} [get]
func (h AdminApisHandler) GetContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	//get all-apps param value
	allApps := false //false by defautl
	allAppsParam := r.URL.Query().Get("all-apps")
	if allAppsParam != "" {
		allApps, _ = strconv.ParseBool(allAppsParam)
	}

	vars := mux.Vars(r)
	id := vars["id"]

	resData, err := h.app.Services.GetContentItem(allApps, claims.AppID, claims.OrgID, id)
	if err != nil {
		log.Printf("Error on getting content item id - %s\n %s", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal the content item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// updateContentItemRequestBody Expected body while updating a new content item
type updateContentItemRequestBody struct {
	AllApps  bool        `json:"all_apps"`
	Category string      `json:"category"`
	Data     interface{} `json:"data"`
} // @name updateContentItemRequestBody

// UpdateContentItem Updates a content item with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Updates a content item with the specified id. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminUpdateContentItem
// @Accept json
// @Produce json
// @Success 200 {object} model.ContentItem
// @Security AdminUserAuth
// @Router /admin/content_items/{id} [put]
func (h AdminApisHandler) UpdateContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var item updateContentItemRequestBody
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the update content item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(item.Category) == 0 {
		log.Printf("Unable to update content item: Missing category")
		http.Error(w, "Unable to update content item: Missing category", http.StatusBadRequest)
		return
	}

	if item.Data == nil {
		log.Printf("Unable to update content item: Missing data")
		http.Error(w, "Unable to update content item: Missing data", http.StatusBadRequest)
		return
	}

	resData, err := h.app.Services.UpdateContentItem(item.AllApps, claims.AppID, claims.OrgID, id, item.Category, item.Data)
	if err != nil {
		log.Printf("Error on updating content item with id - %s\n %s", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal the updated content item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// createContentItemRequestBody Expected body while creating a new content item
type createContentItemRequestBody struct {
	AllApps  bool        `json:"all_apps"`
	Category string      `json:"category" bson:"category"`
	Data     interface{} `json:"data" bson:"data"`
} // @name createContentItemRequestBody

// CreateContentItem creates a new content item. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Creates a new content item. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Admin
// @ID AdminCreateContentItem
// @Accept json
// @Success 200 {object} createContentItemRequestBody
// @Security AdminUserAuth
// @Router /admin/content_items [post]
func (h AdminApisHandler) CreateContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	var item createContentItemRequestBody
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the create content item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(item.Category) == 0 {
		log.Printf("Unable to create content item: Missing category")
		http.Error(w, "Unable to create content item: Missing category", http.StatusBadRequest)
		return
	}

	createdItem, err := h.app.Services.CreateContentItem(item.AllApps, claims.AppID, claims.OrgID, item.Category, item.Data)
	if err != nil {
		log.Printf("Error on creating content item: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(createdItem)
	if err != nil {
		log.Println("Error on marshal the new content item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// DeleteContentItem Deletes a content item with the specified id
// @Description Deletes a content item with the specified id
// @Tags Admin
// @ID AdminDeleteContentItem
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Success 200
// @Security AdminUserAuth
// @Router /admin/content_items/{id} [delete]
func (h AdminApisHandler) DeleteContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	//get all-apps param value
	allApps := false //false by defautl
	allAppsParam := r.URL.Query().Get("all-apps")
	if allAppsParam != "" {
		allApps, _ = strconv.ParseBool(allAppsParam)
	}

	vars := mux.Vars(r)
	guideID := vars["id"]

	err := h.app.Services.DeleteContentItem(allApps, claims.AppID, claims.OrgID, guideID)
	if err != nil {
		log.Printf("Error on deleting content item with id - %s\n %s", guideID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// GetContentItemsCategories Retrieves  all content item categories that have in the database
// @Description Retrieves  all content item categories that have in the database
// @Tags Admin
// @ID AdminGetContentItemsCategories
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Success 200
// @Security AdminUserAuth
// @Router /admin/content_item/categories [get]
func (h AdminApisHandler) GetContentItemsCategories(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	//get all-apps param value
	allApps := false //false by defautl
	allAppsParam := r.URL.Query().Get("all-apps")
	if allAppsParam != "" {
		allApps, _ = strconv.ParseBool(allAppsParam)
	}

	resData, err := h.app.Services.GetContentItemsCategories(allApps, claims.AppID, claims.OrgID)
	if err != nil {
		log.Printf("Error on cgetting content items - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resData == nil {
		resData = []string{}
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal all content items")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// CreateDataContentItem Creates a new data content type item
// @Description Creates a new data content type item
// @Tags Admin
// @ID AdminCreateDataContentItem
// @Accept json
// @Success 200
// @Security AdminUserAuth
// @Router /admin/data [post]
func (h AdminApisHandler) CreateDataContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	var item model.DataContentItem
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the create data content item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdItem, err := h.app.Services.CreateDataContentItem(claims, &item)
	if err != nil {
		log.Printf("Error on creating data content item: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(createdItem)
	if err != nil {
		log.Println("Error on marshal the new item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// GetDataContentItem Gets a data content type item
// @Description Gets a data content type item
// @Tags Admin
// @ID AdminGetDataContentItem
// @Accept json
// @Produce json
// @Success 200
// @Security AdminUserAuth
// @Router /admin/data/{key} [get]
func (h AdminApisHandler) GetDataContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	resData, err := h.app.Services.GetDataContentItem(claims, key)
	if err != nil {
		log.Printf("Error on getting data content type with key - %s\n %s", key, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal of data content type")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetDataContentItems Gets data content items
// @Descriptions Gets data content items
// @Tags Admin
// @ID AdminGetDataContentItems
// @Param category body string false "category - get all data content items based on category"
// @Accept json
// @Produce json
// @Success 200
// @Security AdminUserAuth
// @Router /admin/data [get]
func (h AdminApisHandler) GetDataContentItems(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if len(category) <= 0 {
		log.Print("Missing category\n")
		http.Error(w, "missing 'category' query param", http.StatusBadRequest)
		return
	}

	resData, err := h.app.Services.GetDataContentItems(claims, category)
	if err != nil {
		log.Printf("Error on getting data content type with id - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal of data content type")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// UpdateDataContentItem Updates a content item.
// @Description Updates a content item
// @Tags Admin
// @ID AdminUpdateDataContentItem
// @Accept json
// @Produce json
// @Success 200 {object} model.DataContentItem
// @Security AdminUserAuth
// @Router /admin/data/ [put]
func (h AdminApisHandler) UpdateDataContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {

	var item model.DataContentItem
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the update data content item request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(item.Category) == 0 {
		log.Printf("Unable to update content item: Missing category")
		http.Error(w, "Unable to update content item: Missing category", http.StatusBadRequest)
		return
	}

	if item.Data == nil {
		log.Printf("Unable to update content item: Missing data")
		http.Error(w, "Unable to update content item: Missing data", http.StatusBadRequest)
		return
	}

	resData, err := h.app.Services.UpdateDataContentItem(claims, &item)
	if err != nil {
		log.Printf("Error on updating content item- %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal the updated content item")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// DeleteDataContentItem Deletes a data content item with a specified key
// @Description Deletes a data content item with the specified key
// @Tags Admin
// @ID AdminDeleteDataContentItem
// @Success 200
// @Security AdminUserAuth
// @Router /admin/data/{key} [delete]
func (h AdminApisHandler) DeleteDataContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := h.app.Services.DeleteDataContentItem(claims, key)
	if err != nil {
		log.Printf("Error on deleting data content item with key - %s\n %s", key, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// CreateCategory Creates a category
// @Description Creates a category
// @Tags Admin
// @ID AdminCreateCategory
// @Accept json
// @Success 200
// @Security AdminUserAuth
// @Router /admin/categories [post]
func (h AdminApisHandler) CreateCategory(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	var item model.Category
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the create a category - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdItem, err := h.app.Services.CreateCategory(claims, &item)
	if err != nil {
		log.Printf("Error on creating category %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(createdItem)
	if err != nil {
		log.Println("Error on marshal the new category")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// GetCategory Gets a category
// @Description Gets a category
// @Tags Admin
// @ID AdminGetCategory
// @Accept json
// @Produce json
// @Success 200
// @Security AdminUserAuth
// @Router /admin/categories/{name} [get]
func (h AdminApisHandler) GetCategory(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	resData, err := h.app.Services.GetCategory(claims, name)
	if err != nil {
		log.Printf("Error on getting category with name - %s\n %s", name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal of category")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// UpdateCategory Updates a category with a specificed id.
// @Description  Updates a category with a specificed id.
// @Tags Admin
// @ID AdminUpdateCategory
// @Accept json
// @Produce json
// @Success 200 {object} model.Category
// @Security AdminUserAuth
// @Router /admin/categories [put]
func (h AdminApisHandler) UpdateCategory(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	var item model.Category
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("Error on unmarshal the update category request data - %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resData, err := h.app.Services.UpdateCategory(claims, &item)
	if err != nil {
		log.Printf("Error on updating category  - %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(resData)
	if err != nil {
		log.Println("Error on marshal the updated category")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// DeleteCategory Deletes a category with specified key
// @Description Deletes a category with specified key
// @Tags Admin
// @ID AdminDeleteCategory
// @Success 200
// @Security AdminUserAuth
// @Router /admin/categories/{name} [delete]
func (h AdminApisHandler) DeleteCategory(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	err := h.app.Services.DeleteCategory(claims, name)
	if err != nil {
		log.Printf("Error on deleting category with name - %s\n %s", name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// UploadFileContentItem Uploads a file to AWS S3
// @Description Uploads a file to AWS S3
// @Tags Admin
// @ID AdminUploadFileContentItem
// @Param fileName body string false "fileName - the uploaded file name"
// @Param category body string false "category - category of file content item"
// @Success 200
// @Security AdminUserAuth
// @Router /admin/files [post]
func (h AdminApisHandler) UploadFileContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {

	fileName := r.FormValue("fileName")
	if len(fileName) <= 0 {
		log.Print("Missing file name\n")
		http.Error(w, "missing 'fileName' form param", http.StatusBadRequest)
		return
	}

	category := r.FormValue("category")
	if len(category) <= 0 {
		log.Print("Missing category\n")
		http.Error(w, "missing 'category' form param", http.StatusBadRequest)
		return
	}

	// parse and validate file and post parameters
	file, _, err := r.FormFile("file")
	if err != nil {
		log.Print("Invalid file\n")
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// pass the file to be processed by the use case handler
	err = h.app.Services.UploadFileContentItem(file, claims, fileName, category)
	if err != nil {
		log.Printf("Error converting file: %s\n", err)
		http.Error(w, "Error converting file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetFileContentItem Get a file to AWS S3
// @Description Get a file to AWS S3
// @Tags Admin
// @ID AdminGetFileContentItem
// @Param fileName body string false "fileName - the uploaded file name"
// @Param category body string false "category - category of file content item"
// @Success 200
// @Security AdminUserAuth
// @Router /admin/files [get]
func (h AdminApisHandler) GetFileContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {

	fileName := r.URL.Query().Get("fileName")
	if len(fileName) <= 0 {
		log.Print("Missing file name\n")
		http.Error(w, "missing 'fileName' query param", http.StatusBadRequest)
		return
	}

	category := r.URL.Query().Get("category")
	if len(category) <= 0 {
		log.Print("Missing category\n")
		http.Error(w, "missing 'category' query param", http.StatusBadRequest)
		return
	}

	fileData, err := h.app.Services.GetFileContentItem(claims, fileName, category)
	if err != nil {
		log.Printf("Error getting file download stream: %s\n", err)
		http.Error(w, "Error getting file download stream", http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(w, fileData)
	if err != nil {
		log.Printf("Error copying file into response: %s\n", err.Error())
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.Header().Set("Cache-Control", "no-store")
}

// DeleteFileContentItem Deletes a file content item
// @Description Deletes a file content item
// @Tags Admin
// @ID AdminDeleteFileContentItem
// @Success 200
// @Security AdminUserAuth
// @Router /admin/files [delete]
func (h AdminApisHandler) DeleteFileContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("fileName")
	if len(fileName) <= 0 {
		log.Print("Missing file name\n")
		http.Error(w, "missing 'fileName' query param", http.StatusBadRequest)
		return
	}

	category := r.URL.Query().Get("category")
	if len(category) <= 0 {
		log.Print("Missing category\n")
		http.Error(w, "missing 'category' query param", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeleteFileContentItem(claims, fileName, category)
	if err != nil {
		log.Printf("error on delete AWS file: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
