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

package awsstorage

import (
	"bytes"
	"content/core/model"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
)

const (
	defaultUploadPresignExpirationMinutes   int = 5
	defaultDownloadPresignExpirationMinutes int = 60 * 24
)

// Adapter implements the Storage interface
type Adapter struct {
	config *model.AWSConfig

	uploadPresignExpirationMinutes   int
	downloadPresignExpirationMinutes int
}

// NewAWSStorageAdapter creates a new storage adapter instance
func NewAWSStorageAdapter(config *model.AWSConfig, uploadPresignExpirationMinutes int, downloadPresignExpirationMinutes int) *Adapter {
	//return &Adapter{S3Bucket: S3Bucket, S3Region: S3Region, AWSAccessKeyID: AWSAccessKeyID, AWSSecretAccessKey: AWSSecretAccessKey}
	if uploadPresignExpirationMinutes == 0 {
		uploadPresignExpirationMinutes = defaultUploadPresignExpirationMinutes
	}
	if downloadPresignExpirationMinutes == 0 {
		downloadPresignExpirationMinutes = defaultDownloadPresignExpirationMinutes
	}
	return &Adapter{config: config, uploadPresignExpirationMinutes: uploadPresignExpirationMinutes, downloadPresignExpirationMinutes: downloadPresignExpirationMinutes}
}

// LoadImage loads image at specific path
func (a *Adapter) LoadImage(path string) ([]byte, error) {
	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}

	buffer := aws.NewWriteAtBuffer([]byte{})

	downloader := s3manager.NewDownloader(s)
	_, err = downloader.Download(buffer,
		&s3.GetObjectInput{
			Bucket: aws.String(a.config.S3Bucket),
			Key:    aws.String(path),
		})
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// LoadProfileImage loads profile image at specific path
func (a *Adapter) LoadProfileImage(path string) ([]byte, error) {
	s, err := a.createS3Session(false)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}

	buffer := aws.NewWriteAtBuffer([]byte{})

	downloader := s3manager.NewDownloader(s)
	_, err = downloader.Download(buffer,
		&s3.GetObjectInput{
			Bucket: aws.String(a.config.S3ProfileImagesBucket),
			Key:    aws.String(path),
		})
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// CreateImage uploads an image instance from a file and image type
func (a *Adapter) CreateImage(body io.Reader, path string, preferredFileName *string) (*string, error) {
	log.Println("Create image")

	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}
	key := a.prepareKey(path, preferredFileName)
	objectLocation, err := a.uploadFileToS3(s, body, a.config.S3Bucket, key, "public-read")
	if err != nil {
		log.Printf("Could not upload file")
		return nil, err
	}

	return &objectLocation, nil
}

// CreateProfileImage uploads a profile image
func (a *Adapter) CreateProfileImage(body io.Reader, path string, preferredFileName *string) (*string, error) {
	log.Println("Create profile image")

	s, err := a.createS3Session(false)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}
	key := a.prepareKey(path, preferredFileName)
	objectLocation, err := a.uploadFileToS3(s, body, a.config.S3ProfileImagesBucket, key, "private")
	if err != nil {
		log.Printf("Could not upload file")
		return nil, err
	}

	return &objectLocation, nil
}

// DeleteProfileImage deletes profile image at specific path
func (a *Adapter) DeleteProfileImage(path string) error {
	s, err := a.createS3Session(false)
	if err != nil {
		log.Printf("Could not create S3 session")
		return err
	}

	session := s3.New(s)
	_, err = session.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &a.config.S3ProfileImagesBucket,
		Key:    &path,
	})
	if err != nil {
		return err
	}

	return nil
}

// CreateUserVoiceRecord uploads a voice record for the user
func (a *Adapter) CreateUserVoiceRecord(fileContent []byte, accountID string) (*string, error) {
	log.Println("Create user voice record")

	s, err := a.createS3Session(false)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}
	key := fmt.Sprintf("names-records/%s.m4a", accountID)
	objectLocation, err := a.uploadFileToS3(s, bytes.NewReader(fileContent), a.config.S3UsersAudiosBucket, key, "private")
	if err != nil {
		log.Printf("Could not upload file")
		return nil, err
	}

	return &objectLocation, nil
}

// LoadUserVoiceRecord loads the voice record for the user
func (a *Adapter) LoadUserVoiceRecord(accountID string) ([]byte, error) {
	s, err := a.createS3Session(false)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}

	buffer := aws.NewWriteAtBuffer([]byte{})

	key := fmt.Sprintf("names-records/%s.m4a", accountID)

	downloader := s3manager.NewDownloader(s)
	_, err = downloader.Download(buffer,
		&s3.GetObjectInput{
			Bucket: aws.String(a.config.S3UsersAudiosBucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// DeleteUserVoiceRecord deletes the voice record for the user
func (a *Adapter) DeleteUserVoiceRecord(accountID string) error {
	s, err := a.createS3Session(false)
	if err != nil {
		log.Printf("Could not create S3 session")
		return err
	}

	key := fmt.Sprintf("names-records/%s.m4a", accountID)

	session := s3.New(s)
	_, err = session.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(a.config.S3UsersAudiosBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Adapter) prepareKey(path string, preferredFileName *string) string {
	var fileName string
	if preferredFileName == nil {
		uuid, _ := uuid.NewUUID() // add uuid for file name
		fileName = uuid.String()
	} else {
		fileName = *preferredFileName
	}

	if strings.HasSuffix(path, "/") {
		return path + fmt.Sprintf("%s", fileName) + ".webp"
	}
	return path + "/" + fmt.Sprintf("%s", fileName) + ".webp"
}

// UploadFile uploads an file content item to the s3 bucket
func (a *Adapter) UploadFile(body io.Reader, path string) (*string, error) {
	log.Println("Upload File")

	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}
	objectLocation, err := a.uploadFileToS3(s, body, a.config.S3Bucket, path, "private")
	if err != nil {
		log.Printf("Could not upload file")
		return nil, err
	}

	return &objectLocation, nil
}

// GetPresignedURLsForUpload gets a set of presigned URLs for file upload directly to S3 by a client application
func (a *Adapter) GetPresignedURLsForUpload(fileNames, paths []string) (map[string]string, error) {
	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}

	urls := make(map[string]string)
	for i, path := range paths {
		req, _ := s3.New(s).PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String(a.config.S3Bucket),
			Key:    aws.String(path),
		})
		url, err := req.Presign(time.Duration(a.uploadPresignExpirationMinutes) * time.Minute)
		if err != nil {
			return nil, err
		}
		urls[url] = fileNames[i]
	}
	return urls, nil
}

// DownloadFile loads a file at a specific path
func (a *Adapter) DownloadFile(path string) ([]byte, error) {
	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}

	// file, err := os.Create(path)
	// if err != nil {
	// 	log.Printf("Could not create S3 session")
	// 	return nil, err
	// }

	buffer := aws.NewWriteAtBuffer([]byte{})

	downloader := s3manager.NewDownloader(s)
	_, err = downloader.Download(buffer,
		&s3.GetObjectInput{
			Bucket: aws.String(a.config.S3Bucket),
			Key:    aws.String(path),
		})
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// GetPresignedURLsForDownload gets a set of presigned URLs for file download directly from S3 by a client application
func (a *Adapter) GetPresignedURLsForDownload(fileNames, paths []string) (map[string]string, error) {
	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}

	urls := make(map[string]string)
	for i, path := range paths {
		req, _ := s3.New(s).GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(a.config.S3Bucket),
			Key:    aws.String(path),
		})
		url, err := req.Presign(time.Duration(a.downloadPresignExpirationMinutes) * time.Minute)
		if err != nil {
			return nil, err
		}
		urls[url] = fileNames[i]
	}
	return urls, nil
}

// StreamDownloadFile streams a file downlod from S3
func (a *Adapter) StreamDownloadFile(path string) (io.ReadCloser, error) {
	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}

	file, _ := s3.New(s).GetObject(&s3.GetObjectInput{
		Bucket: aws.String(a.config.S3Bucket),
		Key:    aws.String(path),
	})

	return file.Body, nil
}

// DeleteFile deletes file at specific path
func (a *Adapter) DeleteFile(path string) error {
	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return err
	}

	session := s3.New(s)
	_, err = session.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &a.config.S3Bucket,
		Key:    &path,
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Adapter) createS3Session(accelerate bool) (*session.Session, error) {
	region := a.config.S3Region
	accessKeyID := a.config.AWSAccessKeyID
	secretAccessKey := a.config.AWSSecretAccessKey
	s, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			accessKeyID,
			secretAccessKey,
			""),
		S3UseAccelerate: &accelerate,
	})
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return s, nil
}

// UploadFileToS3 saves a file to aws bucket and returns the url to the file and an error if there's any
func (a *Adapter) uploadFileToS3(s *session.Session, body io.Reader, bucket string, key string, cannedACL string) (string, error) {
	uploader := s3manager.NewUploader(s)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String(cannedACL),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		log.Print(err)
		return "", err
	}
	return result.Location, err
}
