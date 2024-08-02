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

package main

import (
	"content/core"
	"content/core/model"
	"content/driven/awsstorage"
	cacheadapter "content/driven/cache"
	storage "content/driven/storage"
	"content/driven/twitter"
	driver "content/driver/web"
	"log"
	"strings"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/envloader"
	"github.com/rokwire/logging-library-go/v2/logs"
)

var (
	// Version : version of this executable
	Version string
	// Build : build date of this executable
	Build string
)

func main() {
	if len(Version) == 0 {
		Version = "dev"
	}

	serviceID := "content"

	loggerOpts := logs.LoggerOpts{SuppressRequests: logs.NewStandardHealthCheckHTTPRequestProperties(serviceID + "/version")}
	logger := logs.NewLogger(serviceID, &loggerOpts)
	envLoader := envloader.NewEnvLoader(Version, logger)

	envPrefix := strings.ToUpper(serviceID) + "_"

	port := envLoader.GetAndLogEnvVar(envPrefix+"PORT", true, false)

	//mongoDB adapter
	mongoDBAuth := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_AUTH", true, true)
	mongoDBName := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_DATABASE", true, false)
	mongoTimeout := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_TIMEOUT", false, false)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout)
	err := storageAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	// S3 Adapter
	s3Bucket := envLoader.GetAndLogEnvVar(envPrefix+"S3_BUCKET", true, true)
	s3ProfileImagesBucket := envLoader.GetAndLogEnvVar(envPrefix+"S3_PROFILE_IMAGES_BUCKET", true, true)
	s3Region := envLoader.GetAndLogEnvVar(envPrefix+"S3_REGION", true, true)
	awsAccessKeyID := envLoader.GetAndLogEnvVar(envPrefix+"AWS_ACCESS_KEY_ID", true, true)
	awsSecretAccessKey := envLoader.GetAndLogEnvVar(envPrefix+"AWS_SECRET_ACCESS_KEY", true, true)
	awsConfig := &model.AWSConfig{S3Bucket: s3Bucket, S3ProfileImagesBucket: s3ProfileImagesBucket, S3Region: s3Region, AWSAccessKeyID: awsAccessKeyID, AWSSecretAccessKey: awsSecretAccessKey}
	awsAdapter := awsstorage.NewAWSStorageAdapter(awsConfig)

	defaultCacheExpirationSeconds := envLoader.GetAndLogEnvVar(envPrefix+"DEFAULT_CACHE_EXPIRATION_SECONDS", false, false)
	cacheAdapter := cacheadapter.NewCacheAdapter(defaultCacheExpirationSeconds)

	twitterFeedURL := envLoader.GetAndLogEnvVar(envPrefix+"TWITTER_FEED_URL", true, false)
	twitterAccessToken := envLoader.GetAndLogEnvVar(envPrefix+"TWITTER_ACCESS_TOKEN", true, true)
	twitterAdapter := twitter.NewTwitterAdapter(twitterFeedURL, twitterAccessToken)

	mtAppID := envLoader.GetAndLogEnvVar(envPrefix+"MULTI_TENANCY_APP_ID", true, true)
	mtOrgID := envLoader.GetAndLogEnvVar(envPrefix+"MULTI_TENANCY_ORG_ID", true, true)

	// application
	application := core.NewApplication(Version, Build, storageAdapter, awsAdapter, twitterAdapter, cacheAdapter, mtAppID, mtOrgID)
	application.Start()

	// web adapter
	host := envLoader.GetAndLogEnvVar(envPrefix+"HOST", true, false)
	coreBBHost := envLoader.GetAndLogEnvVar(envPrefix+"CORE_BB_HOST", true, false)
	contentServiceURL := envLoader.GetAndLogEnvVar(envPrefix+"SERVICE_URL", true, false)

	authService := authservice.AuthService{
		ServiceID:   "content",
		ServiceHost: contentServiceURL,
		FirstParty:  true,
		AuthBaseURL: coreBBHost,
	}

	serviceRegLoader, err := authservice.NewRemoteServiceRegLoader(&authService, []string{"core"})
	if err != nil {
		log.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := authservice.NewServiceRegManager(&authService, serviceRegLoader, !strings.HasPrefix(contentServiceURL, "http://localhost"))
	if err != nil {
		log.Fatalf("Error initializing service registration manager: %v", err)
	}

	var corsAllowedHeaders []string
	var corsAllowedOrigins []string
	corsAllowedHeadersStr := envLoader.GetAndLogEnvVar("CORS_ALLOWED_HEADERS", false, true)
	if corsAllowedHeadersStr != "" {
		corsAllowedHeaders = strings.Split(corsAllowedHeadersStr, ",")
	}
	corsAllowedOriginsStr := envLoader.GetAndLogEnvVar("CORS_ALLOWED_ORIGINS", false, true)
	if corsAllowedOriginsStr != "" {
		corsAllowedOrigins = strings.Split(corsAllowedOriginsStr, ",")
	}

	webAdapter := driver.NewWebAdapter(host, port, application, serviceRegManager, corsAllowedOrigins, corsAllowedHeaders, logger)
	webAdapter.Start()
}
