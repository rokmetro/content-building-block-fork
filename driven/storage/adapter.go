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
	"content/core/model"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Adapter implements the Storage interface
type Adapter struct {
	db *database
}

// Start starts the storage
func (sa *Adapter) Start() error {
	err := sa.db.start()
	return err
}

// PerformTransaction performs a transaction
func (sa *Adapter) PerformTransaction(transaction func(context TransactionContext) error) error {
	// transaction
	err := sa.db.dbClient.UseSession(context.Background(), func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			sa.abortTransaction(sessionContext)
			return errors.WrapErrorAction(logutils.ActionStart, logutils.TypeTransaction, nil, err)
		}

		err = transaction(sessionContext)
		if err != nil {
			sa.abortTransaction(sessionContext)
			return errors.WrapErrorAction("performing", logutils.TypeTransaction, nil, err)
		}

		err = sessionContext.CommitTransaction(sessionContext)
		if err != nil {
			sa.abortTransaction(sessionContext)
			return errors.WrapErrorAction(logutils.ActionCommit, logutils.TypeTransaction, nil, err)
		}
		return nil
	})

	return err
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
	err := sa.db.studentGuides.Find(filter, &result, nil)
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

	_, err := sa.db.studentGuides.InsertOne(&item)
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
	err := sa.db.studentGuides.Find(filter, &result, nil)
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

	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: id}}
	err := sa.db.studentGuides.ReplaceOne(filter, item, nil)
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
	result, err := sa.db.studentGuides.DeleteOne(filter, nil)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("result is nil for resource item with id " + id)
	}
	deletedCount := result.DeletedCount
	if deletedCount != 1 {
		return fmt.Errorf("error occured while deleting a resource item with id " + id)
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
	err := sa.db.healthLocations.Find(filter, &result, nil)
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

	_, err := sa.db.healthLocations.InsertOne(&item)
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
	err := sa.db.healthLocations.Find(filter, &result, nil)
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

	filter := bson.D{primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "_id", Value: id}}
	err := sa.db.healthLocations.ReplaceOne(filter, item, nil)
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
	result, err := sa.db.healthLocations.DeleteOne(filter, nil)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("result is nil for resource item with id " + id)
	}
	deletedCount := result.DeletedCount
	if deletedCount != 1 {
		return fmt.Errorf("error occured while deleting a resource item with id " + id)
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

	err := sa.db.contentItems.Aggregate(pipeline, &data, &options.AggregateOptions{})
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
	err := sa.db.contentItems.Find(filter, &result, findOptions)
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
	err := sa.db.contentItems.Find(filter, &result, findOptions)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateContentItem creates a new content item record
func (sa *Adapter) CreateContentItem(item model.ContentItem) (*model.ContentItem, error) {
	_, err := sa.db.contentItems.InsertOne(&item)
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
	err := sa.db.contentItems.Find(filter, &result, nil)
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
	_, err := sa.db.contentItems.UpdateOne(filter, update, nil)
	if err != nil {
		log.Printf("error updating content item: %s", err)
		return nil, err
	}

	//get it to return the updated object
	var result []model.ContentItem
	err = sa.db.contentItems.Find(filter, &result, nil)
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
	result, err := sa.db.contentItems.DeleteOne(filter, nil)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("result is nil for resource item with id " + id)
	}
	deletedCount := result.DeletedCount
	if deletedCount != 1 {
		return fmt.Errorf("error occured while deleting a resource item with id " + id)
	}
	return nil
}

// SaveContentItem saves content item
func (sa *Adapter) SaveContentItem(item model.ContentItem) error {
	filter := bson.M{"_id": item.ID}
	opts := options.Replace().SetUpsert(true)
	err := sa.db.contentItems.ReplaceOne(filter, item, opts)
	if err != nil {
		return err
	}
	return nil
}

// FindAllContentItems finds all content items
func (sa *Adapter) FindAllContentItems(context TransactionContext) ([]model.ContentItemResponse, error) {
	filter := bson.D{}
	var result []model.ContentItemResponse
	err := sa.db.contentItems.FindWithContext(context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// StoreMultiTenancyData stores multi-tenancy to already exisiting data in the collections
func (sa *Adapter) StoreMultiTenancyData(context TransactionContext, appID string, orgID string) error {

	filter := bson.D{}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "app_id", Value: appID},
			primitive.E{Key: "org_id", Value: orgID},
		}},
	}
	//content items
	_, err := sa.db.contentItems.UpdateManyWithContext(context, filter, update, nil)
	if err != nil {
		return err
	}
	//health locations
	_, err = sa.db.healthLocations.UpdateManyWithContext(context, filter, update, nil)
	if err != nil {
		return err
	}
	//student guides
	_, err = sa.db.studentGuides.UpdateManyWithContext(context, filter, update, nil)
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
func NewStorageAdapter(mongoDBAuth string, mongoDBName string, mongoTimeout string) *Adapter {
	timeout, err := strconv.Atoi(mongoTimeout)
	if err != nil {
		log.Println("Set default timeout - 500")
		timeout = 500
	}
	timeoutMS := time.Millisecond * time.Duration(timeout)

	db := &database{mongoDBAuth: mongoDBAuth, mongoDBName: mongoDBName, mongoTimeout: timeoutMS}
	return &Adapter{db: db}
}

// TransactionContext wraps mongo.SessionContext for use by external packages
type TransactionContext interface {
	mongo.SessionContext
}
