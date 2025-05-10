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
	"content/core/interfaces"
	"content/driven/awsstorage"
	cacheadapter "content/driven/cache"
	"content/driven/twitter"
	"log"
	"sync"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
)

// Application represents the core application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	cacheLock *sync.Mutex

	Services interfaces.Services //expose to the drivers adapters

	storage        interfaces.Storage
	awsAdapter     *awsstorage.Adapter
	twitterAdapter *twitter.Adapter
	cacheAdapter   *cacheadapter.CacheAdapter

	//TODO - remove this when applied to all environemnts
	multiTenancyAppID string
	multiTenancyOrgID string

	logger *logs.Logger

	//delete data logic
	deleteDataLogic deleteDataLogic
}

// Start starts the core part of the application
func (app *Application) Start() {

	//TODO - remove this when applied to all environemnts
	err := app.storeMultiTenancyData()
	if err != nil {
		log.Fatalf("error initializing multi-tenancy data: %s", err.Error())
	}

	app.deleteDataLogic.start()
}

// as the service starts supporting multi-tenancy we need to add the needed multi-tenancy fields for the existing data,
func (app *Application) storeMultiTenancyData() error {
	log.Println("storeMultiTenancyData...")

	//in transaction
	transaction := func(storage interfaces.Storage) error {
		//check if we need to apply multi-tenancy data
		var applyData bool
		items, err := storage.FindAllContentItems()
		if err != nil {
			return err
		}
		for _, current := range items {
			if val, ok := current["app_id"]; ok {
				log.Printf("\thas already app_id:%s", val)
				applyData = false
				break
			} else {
				log.Print("\tno app_id")
				applyData = true
				break
			}
		}

		//apply data if necessary
		if applyData {
			log.Print("\tapplying multi-tenancy data..")

			err := storage.StoreMultiTenancyData(app.multiTenancyAppID, app.multiTenancyOrgID)
			if err != nil {
				return err
			}
		} else {
			log.Print("\tno need to apply multi-tenancy data, so do nothing")
		}

		return nil
	}

	err := app.storage.PerformTransaction(transaction)
	if err != nil {
		log.Printf("error performing transaction for multi tenancy")
		return err
	}

	return nil
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage interfaces.Storage, awsAdapter *awsstorage.Adapter,
	twitterAdapter *twitter.Adapter, cacheadapter *cacheadapter.CacheAdapter, mtAppID string, mtOrgID string,
	serviceID string, coreBB interfaces.Core, logger *logs.Logger) *Application {
	cacheLock := &sync.Mutex{}
	deleteDataLogic := deleteLogic(*logger, coreBB, serviceID, storage, awsAdapter)

	application := Application{version: version, build: build, cacheLock: cacheLock, storage: storage,
		awsAdapter: awsAdapter, twitterAdapter: twitterAdapter, cacheAdapter: cacheadapter,
		multiTenancyAppID: mtAppID, multiTenancyOrgID: mtOrgID, deleteDataLogic: deleteDataLogic, logger: logger}

	// add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}

	return &application
}
