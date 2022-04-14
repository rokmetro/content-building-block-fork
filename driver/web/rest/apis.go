/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package rest

import (
	"content/core"
	"content/core/model"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/rokmetro/auth-library/tokenauth"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const maxUploadSize = 15 * 1024 * 1024 // 15 mb

//ApisHandler handles the rest APIs implementation
type ApisHandler struct {
	app *core.Application
}

//Version gives the service version
// @Description Gives the service version.
// @Tags Client
// @ID Version
// @Produce plain
// @Success 200
// @Router /version [get]
func (h ApisHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.app.Services.GetVersion()))
}

// GetProfilePhoto Retrieves the profile photo
// @Description Retrieves the profile photo
// @Tags Client
// @ID GetProfilePhoto
// @Param size query string false "Possible values: default, medium, small"
// @Success 200
// @Security RokwireAuth
// @Router /profile_photo/{user-id} [get]
func (h ApisHandler) GetProfilePhoto(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user-id"]
	size := getStringQueryParam(r, "size")
	var sizeType string
	if size != nil {
		if *size == "small" || *size == "medium" || *size == "default" {
			sizeType = *size
		}
	} else {
		sizeType = "default"
	}

	imageBytes, err := h.app.Services.GetProfileImage(userID, sizeType)
	if err != nil || len(imageBytes) == 0 {
		if err != nil {
			log.Printf("error on retrieve AWS image: %s", err)
		} else {
			log.Printf("profile photo not found for user %s", userID)
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/webp")
	w.WriteHeader(http.StatusOK)
	w.Write(imageBytes)
}

// GetUserProfilePhoto Retrieves the profile photo of the requested user
// @Description Retrieves the profile photo of the requested user
// @Tags Client
// @ID GetUserProfilePhoto
// @Param size query string false "Possible values: default, medium, small"
// @Success 200
// @Security RokwireAuth
// @Router /profile_photo [get]
func (h ApisHandler) GetUserProfilePhoto(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	size := getStringQueryParam(r, "size")
	var sizeType string
	if size != nil {
		if *size == "small" || *size == "medium" || *size == "default" {
			sizeType = *size
		}
	} else {
		sizeType = "default"
	}

	imageBytes, err := h.app.Services.GetProfileImage(claims.Subject, sizeType)
	if err != nil || len(imageBytes) == 0 {
		if err != nil {
			log.Printf("error on retrieve AWS image: %s", err)
		} else {
			log.Printf("profile photo not found for user %s", claims.Subject)
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/webp")
	w.WriteHeader(http.StatusOK)
	w.Write(imageBytes)
}

// StoreProfilePhoto Stores profile photo
// @Description Stores profile photo
// @Tags Client
// @ID StoreProfilePhoto
// @Accept json
// @Success 200
// @Security RokwireAuth
// @Router /profile_photo [post]
func (h ApisHandler) StoreProfilePhoto(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {

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
	default:
		log.Print("Invalid file type\n")
		http.Error(w, "Invalid file type. Expected jpeg, png or gif!", http.StatusBadRequest)
		return
	}

	err = h.app.Services.UploadProfileImage(claims.Subject, filetype, fileBytes)
	if err != nil {
		log.Printf("Error converting image: %s\n", err)
		http.Error(w, "Error converting image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// DeleteProfilePhoto Deletes the profile photo of the user who request
// @Description Deletes the profile photo of the user who request
// @Tags Client
// @ID DeleteProfilePhoto
// @Success 200
// @Security RokwireAuth
// @Router /profile_photo [get]
func (h ApisHandler) DeleteProfilePhoto(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {

	err := h.app.Services.DeleteProfileImage(claims.Subject)
	if err != nil {
		if err != nil {
			log.Printf("error on delete AWS profile image: %s", err)
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetStudentGuides retrieves  all student guides
// @Description Retrieves  all student guides
// @Tags Client
// @ID GetStudentGuides
// @Param ids query string false "Coma separated IDs of the desired records"
// @Accept json
// @Success 200
// @Security RokwireAuth
// @Router /student_guides [get]
func (h ApisHandler) GetStudentGuides(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	IDskeys, ok := r.URL.Query()["ids"]
	if ok && len(IDskeys[0]) > 0 {
		extIDs := IDskeys[0]
		IDs = strings.Split(extIDs, ",")
	}

	resData, err := h.app.Services.GetStudentGuides(IDs)
	if err != nil {
		log.Printf("Error on getting student guides by ids - %s\n", err)
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
// @Description Retrieves a student guide by id
// @Tags Client
// @ID GetStudentGuide
// @Accept json
// @Produce json
// @Success 200
// @Security RokwireAuth
// @Router /student_guides/{id} [get]
func (h ApisHandler) GetStudentGuide(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guideID := vars["id"]

	resData, err := h.app.Services.GetStudentGuide(guideID)
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

// GetHealthLocations Retrieves  all health locations
// @Description Retrieves  all health locations
// @Tags Client
// @ID GetHealthLocations
// @Param ids query string false "Coma separated IDs of the desired records"
// @Accept json
// @Success 200
// @Security RokwireAuth
// @Router /health_locations [get]
func (h ApisHandler) GetHealthLocations(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	IDskeys, ok := r.URL.Query()["ids"]
	if ok && len(IDskeys[0]) > 0 {
		extIDs := IDskeys[0]
		IDs = strings.Split(extIDs, ",")
	}

	resData, err := h.app.Services.GetHealthLocations(IDs)
	if err != nil {
		log.Printf("Error on getting health locations by ids - %s\n", err)
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
// @Tags Client
// @ID GetHealthLocation
// @Accept json
// @Produce json
// @Success 200
// @Security RokwireAuth
// @Router /health_locations/{id} [get]
func (h ApisHandler) GetHealthLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guideID := vars["id"]

	resData, err := h.app.Services.GetHealthLocation(guideID)
	if err != nil {
		log.Printf("Error on getting health location id - %s\n %s", guideID, err)
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

// GetContentItems Retrieves  all content items
// @Description Retrieves  all content items
// @Param ids query string false "Coma separated IDs of the desired records"
// @Tags Client
// @ID GetContentItems
// @Param category query string false "limit - Filters by category. Supports query array. Warning: Consider to use as getContentItemsRequestBody json body if you plan to use long list of categories!"
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Param start_date query string false "start_date - Start date filter in milliseconds as an integer epoch value"
// @Param end_date query string false "end_date - End date filter in milliseconds as an integer epoch value"
// @Param data body getContentItemsRequestBody false "body json of the all items ids that need to be filtered"
// @Accept json
// @Success 200
// @Security UserAuth
// @Router /admin/content_items [get]
func (h ApisHandler) GetContentItems(w http.ResponseWriter, r *http.Request) {
	var categoryList []string
	getCategories, ok := r.URL.Query()["category"]
	if ok && len(getCategories) > 0 {
		categoryList = getCategories
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

	var itemIDs []string
	bodyData, _ := ioutil.ReadAll(r.Body)
	if bodyData != nil {
		var body getContentItemsRequestBody
		bodyErr := json.Unmarshal(bodyData, &body)
		if bodyErr == nil {
			itemIDs = body.IDs
			if len(body.Categories) > 0 {
				if categoryList == nil {
					categoryList = body.Categories
				} else {
					categoryList = append(categoryList, body.Categories...)
				}
			}
		}
	}

	resData, err := h.app.Services.GetContentItems(itemIDs, categoryList, offset, limit, order)
	if err != nil {
		log.Printf("Error on cgetting content items - %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resData == nil {
		resData = []model.ContentItem{}
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

// GetContentItem Retrieves a content item by id
// @Description Retrieves a content item by id
// @Tags Client
// @ID GetContentItem
// @Accept json
// @Produce json
// @Success 200
// @Security UserAuth
// @Router /admin/content_items/{id} [get]
func (h ApisHandler) GetContentItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	resData, err := h.app.Services.GetContentItem(id)
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

// GetContentItemsCategories Retrieves  all content item categories that have in the database
// @Description Retrieves  all content item categories that have in the database
// @Tags Client
// @ID GetContentItemsCategories
// @Success 200
// @Security UserAuth
// @Router /content_item/categories [get]
func (h ApisHandler) GetContentItemsCategories(w http.ResponseWriter, r *http.Request) {
	resData, err := h.app.Services.GetContentItemsCategories()
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

// UploadImage Uploads an image to AWS S3
// @Description Uploads an image to AWS S3
// @Tags Client
// @ID UploadImage
// @Param path body string true "path - path within the S3 bucket"
// @Param width body string false "width - width of the image to resize. If width and height are missing - then the new image will use the original size"
// @Param height body string false "height - height of the image to resize. If width and height are missing - then the new image will use the original size"
// @Param quality body string false "quality - quality of the image. Default: 100"
// @Param fileName body string false "fileName - the uploaded file name"
// @Accept multipart/form-data
// @Produce json
// @Success 200
// @Security UserAuth
// @Router /image [post]
func (h ApisHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// validate the image type
	path := r.PostFormValue("path")
	if len(path) == 0 {
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
	file, fileHeader, err := r.FormFile("fileName")
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
	fileName := fileHeader.Filename
	objectLocation, err := h.app.Services.UploadImage(fileName, filetype, fileBytes, path, imgSpec)
	if err != nil {
		log.Printf("Error converting image: %s\n", err)
		http.Error(w, "Error converting image", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(map[string]interface{}{
		"url": objectLocation,
	})
	if err != nil {
		log.Println("Error on marshal s3 location data")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// GetTweeterPosts Retrieves Twitter tweets for the specified user id. This API is intended to be invoked with the original Twitter query params to https://api.twitter.com/2/users/%s/tweets
// @Description Retrieves Twitter tweets for the specified user id. This API is intended to be invoked with the original Twitter query params to https://api.twitter.com/2/users/%s/tweets
// @Tags Client
// @ID GetTweeterPosts
// @Param id path string true "id"
// @Produce json
// @Success 200
// @Security RokwireAuth
// @Router /twitter/users/{user_id}/tweets [get]
func (h ApisHandler) GetTweeterPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	if userID == "" {
		log.Printf("user_id is required query param")
		http.Error(w, "user_id is required query param", http.StatusBadRequest)
		return
	}

	twitterQueryParams := r.URL.RawQuery
	if twitterQueryParams == "" {
		log.Printf("Missing raw query params for Twitter")
		http.Error(w, "Missing raw query params for Twitter", http.StatusBadRequest)
		return
	}

	cacheControl := r.Header.Get("Cache-Control")
	force := cacheControl == "no-cache"

	resData, err := h.app.Services.GetTwitterPosts(userID, twitterQueryParams, force)
	if err != nil {
		log.Printf("Error on getting Twitter Posts: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(resData)
	if err != nil {
		log.Printf("Error on marshal the Twitter Posts: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func intPostValueFromString(stringValue string) int {
	var value int
	if len(stringValue) > 0 {
		val, err := strconv.Atoi(stringValue)
		if err == nil {
			value = val
		}
	}
	return value
}

// NewApisHandler creates new rest Handler instance
func NewApisHandler(app *core.Application) ApisHandler {
	return ApisHandler{app: app}
}

// NewAdminApisHandler creates new rest Handler instance
func NewAdminApisHandler(app *core.Application) AdminApisHandler {
	return AdminApisHandler{app: app}
}
