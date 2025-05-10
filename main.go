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
	corebb "content/driven/core"
	storage "content/driven/storage"
	"content/driven/twitter"
	driver "content/driver/web"
	"log"
	"strconv"
	"strings"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/keys"
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/sigauth"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/envloader"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
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

	//common
	host := envLoader.GetAndLogEnvVar(envPrefix+"HOST", true, false)
	coreBBHost := envLoader.GetAndLogEnvVar(envPrefix+"CORE_BB_HOST", true, false)
	contentServiceURL := envLoader.GetAndLogEnvVar(envPrefix+"SERVICE_URL", true, false)

	authService := auth.Service{
		ServiceID:   serviceID,
		ServiceHost: contentServiceURL,
		FirstParty:  true,
		AuthBaseURL: coreBBHost,
	}

	serviceRegLoader, err := auth.NewRemoteServiceRegLoader(&authService, []string{"auth"})
	if err != nil {
		log.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := auth.NewServiceRegManager(&authService, serviceRegLoader, !strings.HasPrefix(contentServiceURL, "http://localhost"))
	if err != nil {
		log.Fatalf("Error initializing service registration manager: %v", err)
	}
	//end common

	//mongoDB adapter
	mongoDBAuth := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_AUTH", true, true)
	mongoDBName := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_DATABASE", true, false)
	mongoTimeout := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_TIMEOUT", false, false)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout, logger)
	err = storageAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	// S3 Adapter
	s3Bucket := envLoader.GetAndLogEnvVar(envPrefix+"S3_BUCKET", true, true)
	s3BucketAccelerateStr := envLoader.GetAndLogEnvVar(envPrefix+"S3_BUCKET_ACCELERATE", false, false)
	s3BucketAccelerate := false
	s3BucketAccelerate, err = strconv.ParseBool(s3BucketAccelerateStr)
	if err != nil {
		logger.Warnf("error parsing S3 bucket accelerate: %s - applying default", err.Error())
	}
	// only allow S3 transfer accleration on the bucket used for all file types for now

	s3ProfileImagesBucket := envLoader.GetAndLogEnvVar(envPrefix+"S3_PROFILE_IMAGES_BUCKET", true, true)
	s3UsersAudiosBucket := envLoader.GetAndLogEnvVar(envPrefix+"S3_USERS_AUDIOS_BUCKET", true, true)
	s3Region := envLoader.GetAndLogEnvVar(envPrefix+"S3_REGION", true, true)
	awsAccessKeyID := envLoader.GetAndLogEnvVar(envPrefix+"AWS_ACCESS_KEY_ID", true, true)
	awsSecretAccessKey := envLoader.GetAndLogEnvVar(envPrefix+"AWS_SECRET_ACCESS_KEY", true, true)

	awsConfig := &model.AWSConfig{S3Bucket: s3Bucket,
		S3BucketAccelerate:    s3BucketAccelerate,
		S3ProfileImagesBucket: s3ProfileImagesBucket,
		S3UsersAudiosBucket:   s3UsersAudiosBucket,
		S3Region:              s3Region, AWSAccessKeyID: awsAccessKeyID, AWSSecretAccessKey: awsSecretAccessKey}

	uploadPresignExpirationMinutesVal := envLoader.GetAndLogEnvVar(envPrefix+"S3_UPLOAD_PRESIGN_EXPIRATION_MINUTES", false, false)
	uploadPresignExpirationMinutes, err := strconv.Atoi(uploadPresignExpirationMinutesVal)
	if err != nil {
		logger.Warnf("error parsing S3 upload presign expiration minutes: %s - applying default", err.Error())
	}
	downloadPresignExpirationMinutesVal := envLoader.GetAndLogEnvVar(envPrefix+"S3_DOWNLOAD_PRESIGN_EXPIRATION_MINUTES", false, false)
	downloadPresignExpirationMinutes, err := strconv.Atoi(downloadPresignExpirationMinutesVal)
	if err != nil {
		logger.Warnf("error parsing S3 download presign expiration minutes: %s - applying default", err.Error())
	}
	awsAdapter := awsstorage.NewAWSStorageAdapter(awsConfig, uploadPresignExpirationMinutes, downloadPresignExpirationMinutes)

	defaultCacheExpirationSeconds := envLoader.GetAndLogEnvVar(envPrefix+"DEFAULT_CACHE_EXPIRATION_SECONDS", false, false)
	cacheAdapter := cacheadapter.NewCacheAdapter(defaultCacheExpirationSeconds)

	twitterFeedURL := envLoader.GetAndLogEnvVar(envPrefix+"TWITTER_FEED_URL", true, false)
	twitterAccessToken := envLoader.GetAndLogEnvVar(envPrefix+"TWITTER_ACCESS_TOKEN", true, true)
	twitterAdapter := twitter.NewTwitterAdapter(twitterFeedURL, twitterAccessToken)

	mtAppID := envLoader.GetAndLogEnvVar(envPrefix+"MULTI_TENANCY_APP_ID", true, true)
	mtOrgID := envLoader.GetAndLogEnvVar(envPrefix+"MULTI_TENANCY_ORG_ID", true, true)

	//core adapter
	var serviceAccountManager *auth.ServiceAccountManager

	serviceAccountID := envLoader.GetAndLogEnvVar(envPrefix+"SERVICE_ACCOUNT_ID", false, false)
	privKeyRaw := envLoader.GetAndLogEnvVar(envPrefix+"PRIV_KEY", true, true)
	privKeyRaw = strings.ReplaceAll(privKeyRaw, "\\n", "\n")
	privKey, err := keys.NewPrivKey(keys.PS256, privKeyRaw)
	if err != nil {
		logger.Errorf("Error parsing priv key: %v", err)
	} else if serviceAccountID == "" {
		logger.Errorf("Missing service account id")
	} else {
		signatureAuth, err := sigauth.NewSignatureAuth(privKey, serviceRegManager, false, false)
		if err != nil {
			logger.Fatalf("Error initializing signature auth: %v", err)
		}

		serviceAccountLoader, err := auth.NewRemoteServiceAccountLoader(&authService, serviceAccountID, signatureAuth)
		if err != nil {
			logger.Fatalf("Error initializing remote service account loader: %v", err)
		}

		serviceAccountManager, err = auth.NewServiceAccountManager(&authService, serviceAccountLoader)
		if err != nil {
			logger.Fatalf("Error initializing service account manager: %v", err)
		}
	}
	coreAdapter := corebb.NewCoreAdapter(coreBBHost, serviceAccountManager)

	// application
	application := core.NewApplication(Version, Build, storageAdapter, awsAdapter, twitterAdapter, cacheAdapter, mtAppID, mtOrgID, serviceID, coreAdapter, logger)
	application.Start()

	// web adapter
	var corsAllowedHeaders []string
	var corsAllowedOrigins []string
	corsAllowedHeadersStr := envLoader.GetAndLogEnvVar(envPrefix+"CORS_ALLOWED_HEADERS", false, true)
	if corsAllowedHeadersStr != "" {
		corsAllowedHeaders = strings.Split(corsAllowedHeadersStr, ",")
	}
	corsAllowedOriginsStr := envLoader.GetAndLogEnvVar(envPrefix+"CORS_ALLOWED_ORIGINS", false, true)
	if corsAllowedOriginsStr != "" {
		corsAllowedOrigins = strings.Split(corsAllowedOriginsStr, ",")
	}

	webAdapter := driver.NewWebAdapter(host, port, application, serviceRegManager, corsAllowedOrigins, corsAllowedHeaders, logger)
	webAdapter.Start()
}
