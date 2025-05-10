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
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/tokenauth"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/gabriel-vasile/mimetype"
)

const maxUploadSize = 15 * 1024 * 1024 // 15 mb

// ApisHandler handles the rest APIs implementation
type ApisHandler struct {
	app *core.Application
}

// Version gives the service version
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
		msg := fmt.Sprintf("Error parsing request form: max size is %d, err %v", maxUploadSize, err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// parse and validate file and post parameters
	file, _, err := r.FormFile("fileName")
	if err != nil {
		msg := fmt.Sprintf("Error reading file: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		msg := fmt.Sprintf("Error reading file: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
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

	err = h.app.Services.UploadProfileImage(claims.Subject, fileBytes)
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

// StoreVoiceRecord store the user voice record
func (h ApisHandler) StoreVoiceRecord(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	// validate file size
	maxUploadAudioFileSize := int64(5 * 1024 * 1024) // 5 mb
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadAudioFileSize)
	if err := r.ParseMultipartForm(maxUploadAudioFileSize); err != nil {
		msg := fmt.Sprintf("Error parsing request form: max audio file size is %d, err %v", maxUploadAudioFileSize, err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// parse and validate file and post parameters
	file, _, err := r.FormFile("voiceRecord")
	if err != nil {
		msg := fmt.Sprintf("Error reading file: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		msg := fmt.Sprintf("Error reading file: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// check file type
	mime := mimetype.Detect(fileBytes)
	if mime == nil {
		msg := fmt.Sprintf("Error checking file type: %v", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if mime.String() != "audio/mp4" && mime.String() != "audio/x-m4a" &&
		mime.String() != "audio/m4a" && mime.String() != "video/mp4" {
		log.Printf("Invalid file type - %s\n", mime.String())
		http.Error(w, "Invalid file type. Expected m4a!", http.StatusBadRequest)
		return
	}

	// upload voice record
	err = h.app.Services.UploadVoiceRecord(claims.Subject, fileBytes)
	if err != nil {
		log.Printf("Error uploading voice record: %s\n", err)
		http.Error(w, "Error uploading voice record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}

// GetVoiceRecord gets a voice record by user ID
func (h ApisHandler) GetVoiceRecord(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user-id"]

	fileBytes, err := h.app.Services.GetVoiceRecord(userID)
	if err != nil || len(fileBytes) == 0 {
		if err != nil {
			log.Printf("error on retrieve AWS audio file: %s", err)
		} else {
			log.Printf("voice record audio not found for user %s", userID)
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "audio/m4a")
	w.WriteHeader(http.StatusOK)
	w.Write(fileBytes)
}

// GetUserVoiceRecord gets the user voice record
func (h ApisHandler) GetUserVoiceRecord(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {

	fileBytes, err := h.app.Services.GetVoiceRecord(claims.Subject)
	if err != nil || len(fileBytes) == 0 {
		if err != nil {
			log.Printf("error on retrieve AWS audio file: %s", err)
		} else {
			log.Printf("voice record audio not found for user %s", claims.Subject)
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "audio/m4a")
	w.WriteHeader(http.StatusOK)
	w.Write(fileBytes)
}

// DeleteVoiceRecord deletes the user voice record
func (h ApisHandler) DeleteVoiceRecord(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	extension := r.URL.Query().Get("extension")
	if len(extension) <= 0 {
		log.Print("Missing extension query param\n")
		http.Error(w, "missing 'extension' query param", http.StatusBadRequest)
		return
	}

	err := h.app.Services.DeleteVoiceRecord(claims.Subject, extension)
	if err != nil {
		if err != nil {
			log.Printf("error on delete AWS voice audio file: %s", err)
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}

// GetStudentGuides retrieves  all student guides
// @Description Retrieves  all student guides
// @Tags Client
// @ID GetStudentGuides
// @Param ids query string false "Coma separated IDs of the desired records"
// @Accept json
// @Success 200
// @Security RokwireAuth
// @Deprecated true
// @Router /student_guides [get]
func (h ApisHandler) GetStudentGuides(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	IDskeys, ok := r.URL.Query()["ids"]
	if ok && len(IDskeys[0]) > 0 {
		extIDs := IDskeys[0]
		IDs = strings.Split(extIDs, ",")
	}

	resData, err := h.app.Services.GetStudentGuides(claims.AppID, claims.OrgID, IDs)
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
// @Deprecated true
// @Router /student_guides/{id} [get]
func (h ApisHandler) GetStudentGuide(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
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

// GetHealthLocations Retrieves  all health locations
// @Description Retrieves  all health locations
// @Tags Client
// @ID GetHealthLocations
// @Param ids query string false "Coma separated IDs of the desired records"
// @Accept json
// @Success 200
// @Security RokwireAuth
// @Deprecated true
// @Router /health_locations [get]
func (h ApisHandler) GetHealthLocations(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	IDskeys, ok := r.URL.Query()["ids"]
	if ok && len(IDskeys[0]) > 0 {
		extIDs := IDskeys[0]
		IDs = strings.Split(extIDs, ",")
	}

	resData, err := h.app.Services.GetHealthLocations(claims.AppID, claims.OrgID, IDs)
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
// @Deprecated true
// @Router /health_locations/{id} [get]
func (h ApisHandler) GetHealthLocation(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guideID := vars["id"]

	resData, err := h.app.Services.GetHealthLocation(claims.AppID, claims.OrgID, guideID)
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

// GetContentItems Retrieves  all content items. <b> The data element could be either a primitive or nested json or array.</b>
// @Description Retrieves  all content items. <b> The data element could be either a primitive or nested json or array.</b>
// @Tags Client
// @ID GetContentItems
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Param offset query string false "offset"
// @Param limit query string false "limit - limit the result"
// @Param order query string false "order - Possible values: asc, desc. Default: desc"
// @Param data body getContentItemsRequestBody false "Optional - body json of the all items ids that need to be filtered. NOTE: Bad/broken json will be interpreted as an empty filter and the request will be proceeded further."
// @Accept json
// @Success 200 {array} model.ContentItem
// @Security UserAuth
// @Router /content_items [get]
// @Router /content_items [post]
func (h ApisHandler) GetContentItems(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
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

	var body getContentItemsRequestBody
	bodyData, _ := ioutil.ReadAll(r.Body)
	if len(bodyData) > 0 {
		bodyErr := json.Unmarshal(bodyData, &body)
		if bodyErr != nil {
			log.Printf("Warning: bad getContentItemsRequestBody request: %s", bodyErr)
		}
	}

	resData, err := h.app.Services.GetContentItems(allApps, claims.AppID, claims.OrgID, body.IDs, body.Categories, offset, limit, order)
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
// @Tags Client
// @ID GetContentItem
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Accept json
// @Produce json
// @Success 200 {object} model.ContentItem
// @Security UserAuth
// @Router /content_items/{id} [get]
func (h ApisHandler) GetContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
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

// GetContentItemsCategories Retrieves  all content item categories that have in the database
// @Description Retrieves  all content item categories that have in the database
// @Tags Client
// @ID GetContentItemsCategories
// @Param all-apps query boolean false "It says if the data is associated with the current app or it is for all the apps within the organization. It is 'false' by default."
// @Success 200
// @Security UserAuth
// @Router /content_item/categories [get]
func (h ApisHandler) GetContentItemsCategories(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	//get all-apps param value
	allApps := false //false by defautl
	allAppsParam := r.URL.Query().Get("all-apps")
	if allAppsParam != "" {
		allApps, _ = strconv.ParseBool(allAppsParam)
	}

	resData, err := h.app.Services.GetContentItemsCategories(allApps, claims.AppID, claims.OrgID)
	if err != nil {
		log.Printf("Error on getting content items - %s\n", err)
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
func (h ApisHandler) UploadImage(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
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
	objectLocation, err := h.app.Services.UploadImage(fileBytes, path, imgSpec)
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
func (h ApisHandler) GetTweeterPosts(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
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

// GetDataContentItem Gets a data content type item
// @Description Gets a data content type item
// @Tags Client
// @ID GetDataContentItem
// @Accept json
// @Produce json
// @Success 200
// @Security UserAuth
// @Router /data/{key} [get]
func (h ApisHandler) GetDataContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
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

// GetFileContentItem Get a file to AWS S3
// @Description Get a file to AWS S3
// @Tags Client
// @ID GetFileContentItem
// @Param fileName body string false "fileName - the uploaded file name"
// @Param category body string false "category - category of file content item"
// @Success 200
// @Security UserAuth
// @Router /files [get]
func (h ApisHandler) GetFileContentItem(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {

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

// GetFileContentUploadURLs Get URLs to upload files to S3
// @Description Get URLs to upload files to S3
// @Tags Client
// @ID GetFileContentUploadURLs
// @Param fileNames body string false "fileNames - comma-separated list of file names"
// @Param category body string false "category - category of file content item"
// @Param entityID body string false "category - id of entity file content item belongs to"
// @Success 200
// @Security UserAuth
// @Router /files/upload [get]
func (h ApisHandler) GetFileContentUploadURLs(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {

	fileNamesStr := r.URL.Query().Get("fileNames")
	if len(fileNamesStr) <= 0 {
		log.Print("Missing file names query param\n")
		http.Error(w, "missing 'fileNames' query param", http.StatusBadRequest)
		return
	}

	fileNames := strings.Split(fileNamesStr, ",")
	if len(fileNames) <= 0 {
		log.Print("Missing file names\n")
		http.Error(w, "missing file names", http.StatusBadRequest)
		return
	}

	entityID := r.URL.Query().Get("entityID")
	category := r.URL.Query().Get("category")
	if len(category) <= 0 {
		log.Print("Missing category\n")
		http.Error(w, "missing 'category' query param", http.StatusBadRequest)
		return
	}

	handleDuplicateFileNames := true
	handleDuplicateFileNamesStr := r.URL.Query().Get("handle-duplicate-filenames")
	if handleDuplicateFileNamesStr != "" {
		handleDuplicateFileNamesVal, err := strconv.ParseBool(handleDuplicateFileNamesStr)
		if err == nil {
			handleDuplicateFileNames = handleDuplicateFileNamesVal
		}
	}

	addAppOrgIDToPath := true
	addAppOrgIDToPathStr := r.URL.Query().Get("add-path-apporg-id")
	if addAppOrgIDToPathStr != "" {
		addAppOrgIDToPathVal, err := strconv.ParseBool(addAppOrgIDToPathStr)
		if err == nil {
			addAppOrgIDToPath = addAppOrgIDToPathVal
		}
	}

	publicRead := true
	publicReadStr := r.URL.Query().Get("public-read")
	if publicReadStr != "" {
		publicReadVal, err := strconv.ParseBool(publicReadStr)
		if err == nil {
			publicRead = publicReadVal
		}
	}

	fileRefs, err := h.app.Services.GetFileContentUploadURLs(claims, fileNames, entityID, category, addAppOrgIDToPath, handleDuplicateFileNames, publicRead)
	if err != nil {
		log.Printf("Error getting file upload references: %s\n", err)
		http.Error(w, "Error getting file upload references", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fileRefs)
	if err != nil {
		log.Println("Error on marshal of file upload references")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetFileContentDownloadURLs Get URLs to download files from S3
// @Description Get URLs to download files from S3
// @Tags Client
// @ID GetFileContentDownloadURLs
// @Param fileNames body string false "fileNames - comma-separated list of file names"
// @Param category body string false "category - category of file content item"
// @Param entityID body string false "category - id of entity file content item belongs to"
// @Success 200
// @Security UserAuth
// @Router /files/download [get]
func (h ApisHandler) GetFileContentDownloadURLs(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {

	fileKeysStr := r.URL.Query().Get("fileKeys")
	if len(fileKeysStr) <= 0 {
		log.Print("Missing file keys query param\n")
		http.Error(w, "missing 'fileKeys' query param", http.StatusBadRequest)
		return
	}

	fileKeys := strings.Split(fileKeysStr, ",")
	if len(fileKeys) <= 0 {
		log.Print("Missing file keys\n")
		http.Error(w, "missing file keys", http.StatusBadRequest)
		return
	}

	entityID := r.URL.Query().Get("entityID")
	category := r.URL.Query().Get("category")
	if len(category) <= 0 {
		log.Print("Missing category\n")
		http.Error(w, "missing 'category' query param", http.StatusBadRequest)
		return
	}

	addAppOrgIDToPath := true
	addAppOrgIDToPathStr := r.URL.Query().Get("add-path-apporg-id")
	if addAppOrgIDToPathStr != "" {
		addAppOrgIDToPathVal, err := strconv.ParseBool(addAppOrgIDToPathStr)
		if err == nil {
			addAppOrgIDToPath = addAppOrgIDToPathVal
		}
	}

	fileRefs, err := h.app.Services.GetFileContentDownloadURLs(claims, fileKeys, entityID, category, addAppOrgIDToPath)
	if err != nil {
		log.Printf("Error getting file download references: %s\n", err)
		http.Error(w, "Error getting file download references", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fileRefs)
	if err != nil {
		log.Println("Error on marshal of file download references")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetDataContentItems Gets data content items
// @Descriptions Gets data content items
// @Tags Client
// @ID GetDataContentItems
// @Param category body string false "category - get all data content items based on category"
// @Accept json
// @Produce json
// @Success 200
// @Security UserAuth
// @Router /data [get]
func (h ApisHandler) GetDataContentItems(claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if len(category) <= 0 {
		log.Print("Missing category\n")
		http.Error(w, "missing 'category' query param", http.StatusBadRequest)
		return
	}

	resData, err := h.app.Services.GetDataContentItems(claims, category)
	if err != nil {
		log.Printf("Error on getting data content items with category - %s\n %s", category, err)
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

// NewBBSApisHandler creates new rest Handler instance
func NewBBSApisHandler(app *core.Application) BBsApisHandler {
	return BBsApisHandler{app: app}
}

// NewTPSApisHandler creates new rest Handler instance
func NewTPSApisHandler(app *core.Application) TPsApisHandler {
	return TPsApisHandler{app: app}
}
