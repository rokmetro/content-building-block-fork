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

package core

import (
	"bytes"
	"content/core/model"
	"errors"
	"fmt"
	"image"
	_ "image/gif"  // Allow image.Decode to detect GIFs
	_ "image/jpeg" // Allow image.Decode to detect JPEGs
	_ "image/png"  // Allow image.Decode to detect PNGs
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"github.com/rokwire/core-auth-library-go/v3/authutils"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

func (s *servicesImpl) GetVersion() string {
	return s.app.version
}

// Student guides

func (s *servicesImpl) GetStudentGuides(appID string, orgID string, ids []string) ([]bson.M, error) {
	items, err := s.app.storage.GetStudentGuides(appID, orgID, ids)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *servicesImpl) GetStudentGuide(appID string, orgID string, id string) (bson.M, error) {
	item, err := s.app.storage.GetStudentGuide(appID, orgID, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *servicesImpl) CreateStudentGuide(appID string, orgID string, item bson.M) (bson.M, error) {
	items, err := s.app.storage.CreateStudentGuide(appID, orgID, item)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *servicesImpl) UpdateStudentGuide(appID string, orgID string, id string, item bson.M) (bson.M, error) {
	items, err := s.app.storage.UpdateStudentGuide(appID, orgID, id, item)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *servicesImpl) DeleteStudentGuide(appID string, orgID string, id string) error {
	err := s.app.storage.DeleteStudentGuide(appID, orgID, id)
	return err
}

// Health Locations

func (s *servicesImpl) GetHealthLocations(appID string, orgID string, ids []string) ([]bson.M, error) {
	items, err := s.app.storage.GetHealthLocations(appID, orgID, ids)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *servicesImpl) GetHealthLocation(appID string, orgID string, id string) (bson.M, error) {
	item, err := s.app.storage.GetHealthLocation(appID, orgID, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *servicesImpl) CreateHealthLocation(appID string, orgID string, item bson.M) (bson.M, error) {
	items, err := s.app.storage.CreateHealthLocation(appID, orgID, item)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *servicesImpl) UpdateHealthLocation(appID string, orgID string, id string, item bson.M) (bson.M, error) {
	items, err := s.app.storage.UpdateHealthLocation(appID, orgID, id, item)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *servicesImpl) DeleteHealthLocation(appID string, orgID string, id string) error {
	err := s.app.storage.DeleteHealthLocation(appID, orgID, id)
	return err
}

// Content Items

func (s *servicesImpl) GetContentItemsCategories(allApps bool, appID string, orgID string) ([]string, error) {
	//logic
	var appIDParam *string
	if !allApps {
		appIDParam = &appID //associated with current app
	}
	return s.app.storage.GetContentItemsCategories(appIDParam, orgID)
}

func (s *servicesImpl) GetContentItems(allApps bool, appID string, orgID string, ids []string, categoryList []string, offset *int64, limit *int64, order *string) ([]model.ContentItemResponse, error) {
	//logic
	var appIDParam *string
	if !allApps {
		appIDParam = &appID //associated with current app
	}
	return s.app.storage.GetContentItems(appIDParam, orgID, ids, categoryList, offset, limit, order)
}

func (s *servicesImpl) GetContentItem(allApps bool, appID string, orgID string, id string) (*model.ContentItemResponse, error) {
	//logic
	var appIDParam *string
	if !allApps {
		appIDParam = &appID //associated with current app
	}
	return s.app.storage.GetContentItem(appIDParam, orgID, id)
}

func (s *servicesImpl) CreateContentItem(allApps bool, appID string, orgID string, category string, data interface{}) (*model.ContentItem, error) {
	//logic
	var appIDParam *string
	if !allApps {
		appIDParam = &appID //associated with current app
	}
	cItem := model.ContentItem{ID: uuid.NewString(), Category: category, DateCreated: time.Now().UTC(),
		Data: data, OrgID: orgID, AppID: appIDParam}
	return s.app.storage.CreateContentItem(cItem)
}

func (s *servicesImpl) UpdateContentItem(allApps bool, appID string, orgID string, id string, category string, data interface{}) (*model.ContentItem, error) {
	//logic
	var appIDParam *string
	if !allApps {
		appIDParam = &appID //associated with current app
	}

	//update
	item, err := s.app.storage.UpdateContentItem(appIDParam, orgID, id, category, data)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *servicesImpl) UpdateContentItemData(allApps bool, appID string, orgID string, id string, category string, data interface{}) (*model.ContentItem, error) {
	//logic
	var appIDParam *string
	if !allApps {
		appIDParam = &appID //associated with current app
	}

	//find the item
	items, err := s.app.storage.FindContentItems(appIDParam, orgID, []string{id}, []string{category}, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, errors.New("not found")
	}
	item := items[0]

	//update the data
	item.Data = data
	now := time.Now()
	item.DateUpdated = &now

	//save it
	err = s.app.storage.SaveContentItem(item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *servicesImpl) DeleteContentItem(allApps bool, appID string, orgID string, id string) error {
	//logic
	var appIDParam *string
	if !allApps {
		appIDParam = &appID //associated with current app
	}
	return s.app.storage.DeleteContentItem(appIDParam, orgID, id)
}

func (s *servicesImpl) DeleteContentItemByCategory(allApps bool, appID string, orgID string, id string, category string) error {
	//logic
	var appIDParam *string
	if !allApps {
		appIDParam = &appID //associated with current app
	}

	//find the item
	items, err := s.app.storage.FindContentItems(appIDParam, orgID, []string{id}, []string{category}, nil, nil, nil)
	if err != nil {
		return err
	}
	if len(items) != 1 {
		return errors.New("not found")
	}

	//delete it
	err = s.app.storage.DeleteContentItem(appIDParam, orgID, id)
	if err != nil {
		return err
	}

	return nil
}

// Misc

func (s *servicesImpl) UploadImage(imageBytes []byte, path string, spec model.ImageSpec) (*string, error) {
	image, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, fmt.Errorf("Error decoding image: %s", err)
	}

	if spec.Height > 0 || spec.Width > 0 {
		image = resize.Resize(uint(spec.Width), uint(spec.Height), image, resize.Lanczos3)
	}

	quality := 75
	if spec.Quality > 0 {
		quality = spec.Quality
	}

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, float32(quality))
	if err != nil {
		return nil, fmt.Errorf("Error creating webp encoder options: %s", err)
	}

	var output bytes.Buffer
	if err := webp.Encode(&output, image, options); err != nil {
		return nil, fmt.Errorf("Error encoding webp: %s", err)
	}

	url, err := s.app.awsAdapter.CreateImage(&output, path, nil)
	if err != nil {
		return nil, fmt.Errorf("Unable to upload to S3: %s", err)
	}

	if url != nil {
		return url, nil
	}

	return nil, nil
}

func (s *servicesImpl) GetProfileImage(userID string, imageType string) ([]byte, error) {
	return s.app.awsAdapter.LoadProfileImage(fmt.Sprintf("profile-images/%s-%s.webp", userID, imageType))
}

func (s *servicesImpl) UploadProfileImage(userID string, imageBytes []byte) error {
	var mediumImage image.Image
	var smallImage image.Image

	defaultFileNameWebp := fmt.Sprintf("%s-default", userID)
	mediumFileNameWebp := fmt.Sprintf("%s-medium", userID)
	smallFileNameWebp := fmt.Sprintf("%s-small", userID)

	defaultImage, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return fmt.Errorf("Error decoding image: %s", err)
	}

	bounds := defaultImage.Bounds()
	if bounds.Dx() > 512 || bounds.Dy() > 512 {
		image := resize.Resize(512, 0, defaultImage, resize.Lanczos3)
		mediumImage = image
	} else {
		mediumImage = defaultImage
	}

	if bounds.Dx() > 256 || bounds.Dy() > 256 {
		image := resize.Resize(256, 0, defaultImage, resize.Lanczos3)
		smallImage = image
	} else {
		smallImage = defaultImage
	}

	_, err = s.UploadProfileImageToAws(defaultImage, defaultFileNameWebp, "profile-images/", model.ImageSpec{})
	if err != nil {
		return fmt.Errorf("Unable to upload de file: %s. Error: %s", defaultFileNameWebp, err)
	}
	_, err = s.UploadProfileImageToAws(mediumImage, mediumFileNameWebp, "profile-images/", model.ImageSpec{})
	if err != nil {
		return fmt.Errorf("Unable to upload file: %s. Error: %s", mediumFileNameWebp, err)
	}
	_, err = s.UploadProfileImageToAws(smallImage, smallFileNameWebp, "profile-images/", model.ImageSpec{})
	if err != nil {
		return fmt.Errorf("Unable to upload file: %s. Error: %s", smallFileNameWebp, err)
	}

	return nil
}

func (s *servicesImpl) DeleteProfileImage(userID string) error {
	err := s.app.awsAdapter.DeleteProfileImage(fmt.Sprintf("profile-images/%s-default.webp", userID))
	if err != nil {
		return err
	}
	err = s.app.awsAdapter.DeleteProfileImage(fmt.Sprintf("profile-images/%s-medium.webp", userID))
	if err != nil {
		return err
	}
	err = s.app.awsAdapter.DeleteProfileImage(fmt.Sprintf("profile-images/%s-small.webp", userID))
	if err != nil {
		return err
	}
	return nil
}

func (s *servicesImpl) UploadProfileImageToAws(image image.Image, filename string, path string, spec model.ImageSpec) (*string, error) {
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
	if err != nil {
		return nil, fmt.Errorf("Error creating webp encoder options: %s", err)
	}

	var output bytes.Buffer
	if err := webp.Encode(&output, image, options); err != nil {
		return nil, fmt.Errorf("Error encoding webp: %s", err)
	}

	url, err := s.app.awsAdapter.CreateProfileImage(&output, path, &filename)
	if err != nil {
		return nil, fmt.Errorf("Unable to upload to S3: %s", err)
	}

	if url != nil {
		return url, nil
	}

	return nil, nil
}

func (s *servicesImpl) UploadVoiceRecord(userID string, bytes []byte) error {
	_, err := s.app.awsAdapter.CreateUserVoiceRecord(bytes, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *servicesImpl) GetVoiceRecord(userID string) ([]byte, error) {
	fileContent, err := s.app.awsAdapter.LoadUserVoiceRecord(userID)
	if err != nil {
		return nil, err
	}
	return fileContent, nil
}

func (s *servicesImpl) DeleteVoiceRecord(userID string) error {
	err := s.app.awsAdapter.DeleteUserVoiceRecord(userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *servicesImpl) GetTwitterPosts(userID string, twitterQueryParams string, force bool) (map[string]interface{}, error) {
	var err error
	posts := s.app.cacheAdapter.GetTwitterPosts(userID, twitterQueryParams)
	if posts == nil || force {
		s.app.cacheLock.Lock()
		posts = s.app.cacheAdapter.GetTwitterPosts(userID, twitterQueryParams)
		if posts == nil || force {
			if force {
				s.app.cacheAdapter.ClearTwitterCacheForUser(userID)
			}
			posts, err = s.app.twitterAdapter.GetTwitterPosts(userID, twitterQueryParams)
			if err == nil {
				s.app.cacheAdapter.SetTwitterPosts(userID, twitterQueryParams, posts)
			} else {
				fmt.Printf("error feeding twitter: %s", err)
			}
		}
		s.app.cacheLock.Unlock()
	}
	return posts, err
}

func (s *servicesImpl) GetDataContentItem(claims *tokenauth.Claims, key string) (*model.DataContentItem, error) {
	item, err := s.app.storage.FindDataContentItem(&claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *servicesImpl) GetDataContentItems(claims *tokenauth.Claims, category string) ([]*model.DataContentItem, error) {
	item, err := s.app.storage.FindDataContentItems(&claims.AppID, claims.OrgID, category)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *servicesImpl) CreateDataContentItem(claims *tokenauth.Claims, item *model.DataContentItem) (*model.DataContentItem, error) {

	category, err := s.app.storage.FindCategory(&claims.AppID, claims.OrgID, item.Category)
	if err != nil {
		return nil, err
	}

	if !checkPermissions(category.Permissions, claims.Permissions) {
		return nil, fmt.Errorf("unauthorized to create data content item: [%s]", strings.Join(category.Permissions, ", "))
	}

	item.ID = uuid.NewString()
	item.AppID = &claims.AppID
	item.OrgID = claims.OrgID
	item.DateCreated = time.Now().UTC()
	item, err = s.app.storage.CreateDataContentItem(item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *servicesImpl) UpdateDataContentItem(claims *tokenauth.Claims, item *model.DataContentItem) (*model.DataContentItem, error) {
	var dataItem *model.DataContentItem

	category, err := s.app.storage.FindCategory(&claims.AppID, claims.OrgID, item.Category)
	if err != nil {
		return nil, err
	}

	if !checkPermissions(category.Permissions, claims.Permissions) {
		return nil, fmt.Errorf("unauthorized to update data content item: [%s]", strings.Join(category.Permissions, ", "))
	}

	oldItem, err := s.app.storage.FindDataContentItem(&claims.AppID, claims.OrgID, item.Key)
	if err != nil {
		return nil, err
	}

	if item.Category != oldItem.Category {
		category, err = s.app.storage.FindCategory(&claims.AppID, claims.OrgID, oldItem.Category)
		if err != nil {
			return nil, err
		}

		if !checkPermissions(category.Permissions, claims.Permissions) {
			return nil, fmt.Errorf("unauthorized to update data content item: [%s]", strings.Join(category.Permissions, ", "))
		}
	}

	dataItem, err = s.app.storage.UpdateDataContentItem(&claims.AppID, claims.OrgID, item)
	if err != nil {
		return nil, err
	}

	return dataItem, err
}

func (s *servicesImpl) DeleteDataContentItem(claims *tokenauth.Claims, key string) error {

	item, err := s.app.storage.FindDataContentItem(&claims.AppID, claims.OrgID, key)
	if err != nil {
		return err
	}

	category, err := s.app.storage.FindCategory(&claims.AppID, claims.OrgID, item.Category)
	if err != nil {
		return err
	}

	if !checkPermissions(category.Permissions, claims.Permissions) {
		return fmt.Errorf("unauthorized to delete data content item: [%s]", strings.Join(category.Permissions, ", "))
	}

	err = s.app.storage.DeleteDataContentItem(&claims.AppID, claims.OrgID, key)
	if err != nil {
		return err
	}

	return nil
}

func (s *servicesImpl) CreateCategory(claims *tokenauth.Claims, item *model.Category) (*model.Category, error) {
	item.ID = uuid.NewString()
	item.AppID = &claims.AppID
	item.OrgID = claims.OrgID
	item.DateCreated = time.Now().UTC()
	item, err := s.app.storage.CreateCategory(item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *servicesImpl) GetCategory(claims *tokenauth.Claims, name string) (*model.Category, error) {
	item, err := s.app.storage.FindCategory(&claims.AppID, claims.OrgID, name)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *servicesImpl) UpdateCategory(claims *tokenauth.Claims, item *model.Category) (*model.Category, error) {
	item, err := s.app.storage.UpdateCategory(&claims.AppID, claims.OrgID, item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *servicesImpl) DeleteCategory(claims *tokenauth.Claims, name string) error {
	err := s.app.storage.DeleteCategory(&claims.AppID, claims.OrgID, name)
	if err != nil {
		return err
	}
	return nil
}

func (s *servicesImpl) UploadFileContentItem(file io.Reader, claims *tokenauth.Claims, fileName string, category string) error {

	path := claims.OrgID + "/" + claims.AppID + "/" + category + "/" + fileName

	categoryItem, err := s.app.storage.FindCategory(&claims.AppID, claims.OrgID, category)
	if err != nil {
		return err
	}

	if !checkPermissions(categoryItem.Permissions, claims.Permissions) {
		return fmt.Errorf("unauthorized to upload file content item: [%s]", strings.Join(categoryItem.Permissions, ", "))
	}

	_, err = s.app.awsAdapter.UploadFile(file, path)
	if err != nil {
		return fmt.Errorf("unable to upload to S3: %s", err)
	}

	return nil
}

func (s *servicesImpl) GetFileContentItem(claims *tokenauth.Claims, fileName string, category string) (io.ReadCloser, error) {

	path := claims.OrgID + "/" + claims.AppID + "/" + category + "/" + fileName

	fileData, err := s.app.awsAdapter.StreamDownloadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to get data for file download stream: %s", err.Error())
	}

	return fileData, nil
}

func (s *servicesImpl) GetFileContentUploadURLs(claims *tokenauth.Claims, count int, entityID string, category string) ([]model.FileContentItemRef, error) {
	paths := make([]string, count)
	fileIDs := make([]string, count)
	for i := 0; i < count; i++ {
		fileIDs[i] = uuid.NewString()

		paths[i] = claims.OrgID + "/" + claims.AppID + "/" + category
		if entityID != "" {
			paths[i] += "/" + entityID
		}
		paths[i] += "/" + fileIDs[i]
	}

	fileRefs, err := s.app.awsAdapter.GetPresignedURLsForUpload(fileIDs, paths)
	if err != nil {
		return nil, fmt.Errorf("unable to get file upload references: %s", err.Error())
	}

	return fileRefs, nil
}

func (s *servicesImpl) GetFileContentDownloadURLs(claims *tokenauth.Claims, fileIDs []string, entityID string, category string) ([]model.FileContentItemRef, error) {
	paths := make([]string, len(fileIDs))
	for i, id := range fileIDs {
		paths[i] = claims.OrgID + "/" + claims.AppID + "/" + category
		if entityID != "" {
			paths[i] += "/" + entityID
		}
		paths[i] += "/" + id
	}

	fileRefs, err := s.app.awsAdapter.GetPresignedURLsForDownload(fileIDs, paths)
	if err != nil {
		return nil, fmt.Errorf("unable to get file download references: %s", err.Error())
	}

	return fileRefs, nil
}

func (s *servicesImpl) DeleteFileContentItem(claims *tokenauth.Claims, fileName string, category string) error {
	categoryItem, err := s.app.storage.FindCategory(&claims.AppID, claims.OrgID, category)
	if err != nil {
		return err
	}

	if !checkPermissions(categoryItem.Permissions, claims.Permissions) {
		return fmt.Errorf("unauthorized to delete file content item: [%s]", strings.Join(categoryItem.Permissions, ", "))
	}

	path := claims.OrgID + "/" + claims.AppID + "/" + category + "/" + fileName

	err = s.app.awsAdapter.DeleteFile(path)
	if err != nil {
		return err
	}

	return nil
}

func checkPermissions(itemPermissions []string, claimsPermissions string) bool {
	permissions := strings.Split(claimsPermissions, ",")
	for _, element := range itemPermissions {
		if authutils.ContainsString(permissions, element) {
			return true
		}
	}

	return false
}

type servicesImpl struct {
	app *Application
}
