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

package storage

import (
	"content/core/interfaces"
	"content/core/model"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Adapter implements the Storage interface
type Adapter struct {
	db      *database
	context mongo.SessionContext
}

// Start starts the storage
func (sa *Adapter) Start() error {
	err := sa.db.start()
	return err
}

// PerformTransaction performs a transaction
func (sa *Adapter) PerformTransaction(transaction func(storage interfaces.Storage) error) error {
	// transaction
	callback := func(sessionContext mongo.SessionContext) (interface{}, error) {
		adapter := sa.withContext(sessionContext)

		err := transaction(adapter)
		if err != nil {
			if wrappedErr, ok := err.(interface {
				Internal() error
			}); ok && wrappedErr.Internal() != nil {
				return nil, wrappedErr.Internal()
			}
			return nil, err
		}

		return nil, nil
	}

	session, err := sa.db.dbClient.StartSession()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionStart, "mongo session", nil, err)
	}
	context := context.Background()
	defer session.EndSession(context)

	_, err = session.WithTransaction(context, callback)
	if err != nil {
		return errors.WrapErrorAction("performing", logutils.TypeTransaction, nil, err)
	}
	return nil
}

// GetStudentGuides retrieves all content items
func (sa *Adapter) GetStudentGuides(appID string, orgID string, ids []string) ([]bson.M, error) {
	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID}}
	if len(ids) > 0 {
		filter = bson.D{
			primitive.E{Key: "_id", Value: bson.M{"$in": ids}},
		}
	}

	var result []bson.M
	err := sa.db.studentGuides.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateStudentGuide creates a new student guide record
func (sa *Adapter) CreateStudentGuide(appID string, orgID string, item bson.M) (bson.M, error) {

	id := item["_id"]
	if id == nil {
		item["_id"] = uuid.NewString()
	}
	item["app_id"] = appID
	item["org_id"] = orgID

	_, err := sa.db.studentGuides.InsertOne(sa.context, &item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// GetStudentGuide retrieves a student guide record by id
func (sa *Adapter) GetStudentGuide(appID string, orgID string, id string) (bson.M, error) {

	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: id}}
	var result []bson.M
	err := sa.db.studentGuides.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if result == nil || len(result) == 0 {
		//not found
		return nil, fmt.Errorf("student guide with id: %s is not found", id)
	}
	return result[0], nil

}

// UpdateStudentGuide updates a student guide record
func (sa *Adapter) UpdateStudentGuide(appID string, orgID string, id string, item bson.M) (bson.M, error) {
	jsonID := item["_id"]
	if jsonID == nil && jsonID != id {
		return nil, fmt.Errorf("attempt to override another object")
	}

	item["app_id"] = appID
	item["org_id"] = orgID

	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: id}}
	err := sa.db.studentGuides.ReplaceOne(sa.context, filter, item, nil)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// DeleteStudentGuide deletes a student guide record with the desired id
func (sa *Adapter) DeleteStudentGuide(appID string, orgID string, id string) error {
	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: id}}
	result, err := sa.db.studentGuides.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("result is nil for resource item with id %s", id)
	}
	deletedCount := result.DeletedCount
	if deletedCount != 1 {
		return fmt.Errorf("error occured while deleting a resource item with id %s", id)
	}
	return nil

}

//// Health locations

// GetHealthLocations retrieves all content items
func (sa *Adapter) GetHealthLocations(appID string, orgID string, ids []string) ([]bson.M, error) {
	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID}}
	if len(ids) > 0 {
		filter = bson.D{
			primitive.E{Key: "_id", Value: bson.M{"$in": ids}},
		}
	}

	var result []bson.M
	err := sa.db.healthLocations.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateHealthLocation creates a new health location record
func (sa *Adapter) CreateHealthLocation(appID string, orgID string, item bson.M) (bson.M, error) {

	id := item["_id"]
	if id == nil {
		item["_id"] = uuid.NewString()
	}
	item["app_id"] = appID
	item["org_id"] = orgID

	_, err := sa.db.healthLocations.InsertOne(sa.context, &item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// GetHealthLocation retrieves a health location record by id
func (sa *Adapter) GetHealthLocation(appID string, orgID string, id string) (bson.M, error) {

	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: id}}
	var result []bson.M
	err := sa.db.healthLocations.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if result == nil || len(result) == 0 {
		//not found
		return nil, fmt.Errorf("student guide with id: %s is not found", id)
	}
	return result[0], nil

}

// UpdateHealthLocation updates a health location record
func (sa *Adapter) UpdateHealthLocation(appID string, orgID string, id string, item bson.M) (bson.M, error) {
	jsonID := item["_id"]
	if jsonID == nil && jsonID != id {
		return nil, fmt.Errorf("attempt to override another object")
	}
	item["app_id"] = appID
	item["org_id"] = orgID

	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: id}}
	err := sa.db.healthLocations.ReplaceOne(sa.context, filter, item, nil)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// DeleteHealthLocation deletes a health location record with the desired id
func (sa *Adapter) DeleteHealthLocation(appID string, orgID string, id string) error {
	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: id}}
	result, err := sa.db.healthLocations.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("result is nil for resource item with id %s", id)
	}
	deletedCount := result.DeletedCount
	if deletedCount != 1 {
		return fmt.Errorf("error occured while deleting a resource item with id %s", id)
	}
	return nil

}

// Content Items

type getContentItemsCategoriesData struct {
	CategoryName string `json:"_id" bson:"_id"`
}

// GetContentItemsCategories  retrieve all content item categories
func (sa *Adapter) GetContentItemsCategories(appID *string, orgID string) ([]string, error) {
	pipeline := primitive.A{
		bson.M{"$match": bson.M{"app_id": appID, "org_id": orgID}},
		bson.M{"$group": bson.M{"_id": "$category"}},
	}
	var data []getContentItemsCategoriesData
	categories := []string{}

	err := sa.db.contentItems.Aggregate(sa.context, pipeline, &data, &options.AggregateOptions{})
	if err != nil {
		return nil, err
	}
	if data != nil && len(data) > 0 {
		for _, dataEntry := range data {
			categories = append(categories, dataEntry.CategoryName)
		}
	}

	return categories, nil
}

// FindContentItems finds content items
func (sa *Adapter) FindContentItems(appID *string, orgID string, ids []string, categoryList []string, offset *int64, limit *int64, order *string) ([]model.ContentItem, error) {
	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID}}
	if len(ids) > 0 {
		filter = append(filter, primitive.E{Key: "_id", Value: bson.M{"$in": ids}})
	}
	if categoryList != nil && len(categoryList) > 0 {
		filter = append(filter, primitive.E{Key: "category", Value: bson.M{"$in": categoryList}})
	}

	findOptions := options.Find()
	if order != nil && "desc" == *order {
		findOptions.SetSort(bson.M{"date_created": -1})
	} else {
		findOptions.SetSort(bson.M{"date_created": 1})
	}
	if limit != nil {
		findOptions.SetLimit(*limit)
	}
	if offset != nil {
		findOptions.SetSkip(*offset)
	}

	var result []model.ContentItem
	err := sa.db.contentItems.Find(sa.context, filter, &result, findOptions)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetContentItems retrieves all content items
func (sa *Adapter) GetContentItems(appID *string, orgID string, ids []string, categoryList []string, offset *int64, limit *int64, order *string) ([]model.ContentItemResponse, error) {

	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID}}
	if len(ids) > 0 {
		filter = append(filter, primitive.E{Key: "_id", Value: bson.M{"$in": ids}})
	}
	if categoryList != nil && len(categoryList) > 0 {
		filter = append(filter, primitive.E{Key: "category", Value: bson.M{"$in": categoryList}})
	}

	findOptions := options.Find()
	if order != nil && "desc" == *order {
		findOptions.SetSort(bson.M{"date_created": -1})
	} else {
		findOptions.SetSort(bson.M{"date_created": 1})
	}
	if limit != nil {
		findOptions.SetLimit(*limit)
	}
	if offset != nil {
		findOptions.SetSkip(*offset)
	}

	var result []model.ContentItemResponse
	err := sa.db.contentItems.Find(sa.context, filter, &result, findOptions)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateContentItem creates a new content item record
func (sa *Adapter) CreateContentItem(item model.ContentItem) (*model.ContentItem, error) {
	_, err := sa.db.contentItems.InsertOne(sa.context, &item)
	if err != nil {
		log.Printf("error create content item: %s", err)
		return nil, err
	}
	return &item, nil
}

// GetContentItem retrieves a content item record by id
func (sa *Adapter) GetContentItem(appID *string, orgID string, id string) (*model.ContentItemResponse, error) {

	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: id}}
	var result []model.ContentItemResponse
	err := sa.db.contentItems.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if result == nil || len(result) == 0 {
		//not found
		log.Printf("content item with id: %s is not found", id)
		return nil, fmt.Errorf("content item with id: %s is not found", id)
	}
	return &result[0], nil

}

// UpdateContentItem updates a content item record
func (sa *Adapter) UpdateContentItem(appID *string, orgID string, id string,
	category string, data interface{}) (*model.ContentItem, error) {
	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: id}}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "category", Value: category},
			primitive.E{Key: "data", Value: data},
			primitive.E{Key: "date_updated", Value: time.Now().UTC()},
		}},
	}
	_, err := sa.db.contentItems.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		log.Printf("error updating content item: %s", err)
		return nil, err
	}

	//get it to return the updated object
	var result []model.ContentItem
	err = sa.db.contentItems.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if result == nil || len(result) == 0 {
		//not found
		log.Printf("content item with id: %s is not found", id)
		return nil, fmt.Errorf("content item with id: %s is not found", id)
	}
	return &result[0], nil
}

// DeleteContentItem deletes a content item record with the desired id
func (sa *Adapter) DeleteContentItem(appID *string, orgID string, id string) error {
	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: id}}
	result, err := sa.db.contentItems.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("result is nil for resource item with id %s", id)
	}
	deletedCount := result.DeletedCount
	if deletedCount != 1 {
		return fmt.Errorf("error occured while deleting a resource item with id %s", id)
	}
	return nil
}

// SaveContentItem saves content item
func (sa *Adapter) SaveContentItem(item model.ContentItem) error {
	filter := bson.D{primitive.E{Key: "org_id", Value: item.OrgID},
		primitive.E{Key: "_id", Value: item.ID}}
	if item.AppID != nil {
		filter = append(filter, primitive.E{Key: "app_id", Value: item.AppID})
	}

	opts := options.Replace().SetUpsert(true)
	err := sa.db.contentItems.ReplaceOne(sa.context, filter, item, opts)
	if err != nil {
		return err
	}
	return nil
}

// FindAllContentItems finds all content items
func (sa *Adapter) FindAllContentItems() ([]model.ContentItemResponse, error) {
	filter := bson.D{}
	var result []model.ContentItemResponse
	err := sa.db.contentItems.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateDataContentItem creates a data content item
func (sa *Adapter) CreateDataContentItem(item *model.DataContentItem) (*model.DataContentItem, error) {

	_, err := sa.db.dataContentItems.InsertOne(sa.context, &item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// FindDataContentItem gets a data content item
func (sa *Adapter) FindDataContentItem(appID *string, orgID string, key string) (*model.DataContentItem, error) {

	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "key", Value: key}}

	var result *model.DataContentItem
	err := sa.db.dataContentItems.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FindDataContentItems gets multiple data content items
func (sa *Adapter) FindDataContentItems(appID *string, orgID string, category string) ([]*model.DataContentItem, error) {
	var filter bson.D
	if len(category) > 0 {
		filter = bson.D{primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "org_id", Value: orgID},
			primitive.E{Key: "category", Value: category}}
	} else {
		filter = bson.D{primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "org_id", Value: orgID}}
	}

	var result []*model.DataContentItem
	err := sa.db.dataContentItems.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateDataContentItem updates a data content item
func (sa *Adapter) UpdateDataContentItem(appID *string, orgID string, item *model.DataContentItem) (*model.DataContentItem, error) {

	filter := bson.D{
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "key", Value: item.Key}}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "category", Value: item.Category},
			primitive.E{Key: "data", Value: item.Data},
			primitive.E{Key: "date_updated", Value: time.Now().UTC()},
		}},
	}
	_, err := sa.db.dataContentItems.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		log.Printf("error updating data content item: %s", err)
		return nil, err
	}

	return item, nil
}

// DeleteDataContentItem deletes a data content item
func (sa *Adapter) DeleteDataContentItem(appID *string, orgID string, key string) error {

	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "key", Value: key}}

	result, err := sa.db.dataContentItems.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("result is nil for data content item with key %s", key)
	}
	deletedCount := result.DeletedCount
	if deletedCount != 1 {
		return fmt.Errorf("error occured while deleting a data content item with key %s", key)
	}
	return nil
}

// CreateCategory created a new category
func (sa *Adapter) CreateCategory(item *model.Category) (*model.Category, error) {

	_, err := sa.db.categories.InsertOne(sa.context, &item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// FindCategory fins a category
func (sa *Adapter) FindCategory(appID *string, orgID string, name string) (*model.Category, error) {
	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "name", Value: name}}

	var result *model.Category
	err := sa.db.categories.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateCategory updates a  category
func (sa *Adapter) UpdateCategory(appID *string, orgID string, item *model.Category) (*model.Category, error) {
	filter := bson.D{
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: item.ID}}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "name", Value: item.Name},
			primitive.E{Key: "permissions", Value: item.Permissions},
			primitive.E{Key: "date_updated", Value: time.Now().UTC()},
		}},
	}
	_, err := sa.db.categories.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		log.Printf("error updating category: %s", err)
		return nil, err
	}

	return item, nil
}

// DeleteCategory deletes a category
func (sa *Adapter) DeleteCategory(appID *string, orgID string, name string) error {
	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "name", Value: name}}

	result, err := sa.db.categories.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("result is nil for cateogry with id %s", name)
	}
	deletedCount := result.DeletedCount
	if deletedCount != 1 {
		return fmt.Errorf("error occured while deleting a category with id %s", name)
	}
	return nil
}

// StoreMultiTenancyData stores multi-tenancy to already exisiting data in the collections
func (sa *Adapter) StoreMultiTenancyData(appID string, orgID string) error {

	filter := bson.D{}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "org_id", Value: orgID},
		}},
	}
	//content items
	_, err := sa.db.contentItems.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		return err
	}
	//health locations
	_, err = sa.db.healthLocations.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		return err
	}
	//student guides
	_, err = sa.db.studentGuides.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		return err
	}

	return nil
}

func (sa *Adapter) abortTransaction(sessionContext mongo.SessionContext) {
	err := sessionContext.AbortTransaction(sessionContext)
	if err != nil {
		log.Printf("error aborting a transaction - %s", err)
	}
}

// NewStorageAdapter creates a new storage adapter instance
func NewStorageAdapter(mongoDBAuth string, mongoDBName string, mongoTimeout string, logger *logs.Logger) *Adapter {
	timeout, err := strconv.Atoi(mongoTimeout)
	if err != nil {
		log.Println("Set default timeout - 500")
		timeout = 500
	}
	timeoutMS := time.Millisecond * time.Duration(timeout)

	db := &database{mongoDBAuth: mongoDBAuth, mongoDBName: mongoDBName, mongoTimeout: timeoutMS, logger: logger}
	return &Adapter{db: db}
}

// Creates a new Adapter with provided context
func (sa *Adapter) withContext(context mongo.SessionContext) *Adapter {
	return &Adapter{db: sa.db, context: context}
}
