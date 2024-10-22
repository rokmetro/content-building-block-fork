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
	"context"
	"log"
	"time"

	"github.com/rokwire/logging-library-go/v2/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type database struct {
	mongoDBAuth  string
	mongoDBName  string
	mongoTimeout time.Duration

	db       *mongo.Database
	dbClient *mongo.Client

	studentGuides    *collectionWrapper
	healthLocations  *collectionWrapper
	contentItems     *collectionWrapper
	dataContentItems *collectionWrapper
	categories       *collectionWrapper

	logger *logs.Logger
}

func (m *database) start() error {

	log.Println("database -> start")

	//connect to the database
	clientOptions := options.Client().ApplyURI(m.mongoDBAuth)
	connectContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	client, err := mongo.Connect(connectContext, clientOptions)
	cancel()
	if err != nil {
		return err
	}

	//ping the database
	pingContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	err = client.Ping(pingContext, nil)
	cancel()
	if err != nil {
		return err
	}

	//apply checks
	db := client.Database(m.mongoDBName)

	studentGuides := &collectionWrapper{database: m, coll: db.Collection("student_guides")}
	err = m.applyStudentGuidesChecks(studentGuides)
	if err != nil {
		return err
	}

	healthLocations := &collectionWrapper{database: m, coll: db.Collection("health_locations")}
	err = m.applyHealthLocationsChecks(healthLocations)
	if err != nil {
		return err
	}

	contentItems := &collectionWrapper{database: m, coll: db.Collection("content_items")}
	err = m.applyContentItemsChecks(contentItems)
	if err != nil {
		return err
	}

	dataContentItems := &collectionWrapper{database: m, coll: db.Collection("data_content_items")}
	err = m.applyDataContentItemsChecks(dataContentItems)
	if err != nil {
		return err
	}

	categories := &collectionWrapper{database: m, coll: db.Collection("categories")}
	err = m.applyCategoriesChecks(categories)
	if err != nil {
		return err
	}

	//asign the db, db client and the collections
	m.db = db
	m.dbClient = client

	m.studentGuides = studentGuides
	m.healthLocations = healthLocations
	m.contentItems = contentItems
	m.dataContentItems = dataContentItems
	m.categories = categories

	return nil
}

func (m *database) applyStudentGuidesChecks(studentGuides *collectionWrapper) error {
	log.Println("apply student guides checks.....")

	//Add org_id + app_id index
	err := studentGuides.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1},
		primitive.E{Key: "app_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("student guides checks passed")
	return nil
}

func (m *database) applyHealthLocationsChecks(healthLocations *collectionWrapper) error {
	log.Println("health locations guides checks.....")

	//Add org_id + app_id index
	err := healthLocations.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1},
		primitive.E{Key: "app_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("health locations checks passed")
	return nil
}

func (m *database) applyContentItemsChecks(contentItems *collectionWrapper) error {
	log.Println("apply content_items checks.....")

	//Add org_id + app_id index
	err := contentItems.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1},
		primitive.E{Key: "app_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	// Add category index
	err = contentItems.AddIndex(bson.D{primitive.E{Key: "category", Value: 1}}, false)
	if err != nil {
		return err
	}

	// Add date_created index
	err = contentItems.AddIndex(bson.D{primitive.E{Key: "date_created", Value: 1}}, false)
	if err != nil {
		return err
	}

	log.Println("content_items checks passed")
	return nil
}

func (m *database) applyDataContentItemsChecks(dataContentItems *collectionWrapper) error {
	log.Println("apply data_content_items checks.....")

	//Add org_id + app_id index
	err := dataContentItems.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "key", Value: 1}}, true)
	if err != nil {
		return err
	}

	log.Println("data_content_items checks passed")
	return nil
}

func (m *database) applyCategoriesChecks(categories *collectionWrapper) error {
	log.Println("apply categories checks.....")

	//Add org_id + app_id index
	err := categories.AddIndex(bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "name", Value: 1}}, true)
	if err != nil {
		return err
	}

	log.Println("categories checks passed")
	return nil
}

// Event

func (m *database) onDataChanged(changeDoc map[string]interface{}) {
	if changeDoc == nil {
		return
	}
	log.Printf("onDataChanged: %+v\n", changeDoc)
	ns := changeDoc["ns"]
	if ns == nil {
		return
	}
	nsMap := ns.(map[string]interface{})
	coll := nsMap["coll"]

	if "configs" == coll {
		log.Println("configs collection changed")
	} else {
		log.Println("other collection changed")
	}
}
