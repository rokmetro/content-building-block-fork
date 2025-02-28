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
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	defaultUploadPresignExpirationMinutes          int = 5
	defaultMultipartUploadPresignExpirationMinutes int = 15
	defaultDownloadPresignExpirationMinutes        int = 60 * 24

	mib         int = 1024 * 1024
	gib         int = 1024 * mib
	tib         int = 1024 * gib
	maxFileSize int = 5 * tib
	minPartSize int = 5 * mib
	maxPartSize int = 5 * gib
	maxParts    int = 10000
)

// Adapter implements the Storage interface
type Adapter struct {
	config *model.AWSConfig

	uploadPresignExpirationMinutes          int
	multipartUploadPresignExpirationMinutes int
	downloadPresignExpirationMinutes        int

	logger *logs.Logger
}

// NewAWSStorageAdapter creates a new storage adapter instance
func NewAWSStorageAdapter(config *model.AWSConfig, uploadPresignExpirationMinutes int, multipartUploadPresignExpirationMinutes int, downloadPresignExpirationMinutes int, logger *logs.Logger) *Adapter {
	//return &Adapter{S3Bucket: S3Bucket, S3Region: S3Region, AWSAccessKeyID: AWSAccessKeyID, AWSSecretAccessKey: AWSSecretAccessKey}
	if uploadPresignExpirationMinutes == 0 {
		uploadPresignExpirationMinutes = defaultUploadPresignExpirationMinutes
	}
	if multipartUploadPresignExpirationMinutes == 0 {
		multipartUploadPresignExpirationMinutes = defaultMultipartUploadPresignExpirationMinutes
	}
	if downloadPresignExpirationMinutes == 0 {
		downloadPresignExpirationMinutes = defaultDownloadPresignExpirationMinutes
	}
	return &Adapter{
		config:                                  config,
		uploadPresignExpirationMinutes:          uploadPresignExpirationMinutes,
		multipartUploadPresignExpirationMinutes: multipartUploadPresignExpirationMinutes,
		downloadPresignExpirationMinutes:        downloadPresignExpirationMinutes,
		logger:                                  logger,
	}
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
func (a *Adapter) GetPresignedURLsForUpload(fileKeys, paths []string) ([]model.FileContentItemRef, error) {
	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}

	refs := make([]model.FileContentItemRef, len(paths))
	for i, path := range paths {
		req, _ := s3.New(s).PutObjectRequest(&s3.PutObjectInput{
			Bucket: aws.String(a.config.S3Bucket),
			Key:    aws.String(path),
		})
		url, err := req.Presign(time.Duration(a.uploadPresignExpirationMinutes) * time.Minute)
		if err != nil {
			return nil, err
		}
		refs[i] = model.FileContentItemRef{Key: fileKeys[i], URL: url}
	}
	return refs, nil
}

// GetPresignedURLsForMultipartUpload creates a multipart upload and generates a list of signed URLs for the client to use to upload file parts
func (a *Adapter) GetPresignedURLsForMultipartUpload(fileKey string, path string, fileSize int) (*model.FileContentItemMultipartUpload, error) {
	if fileSize > maxFileSize {
		return nil, errors.ErrorData(logutils.StatusInvalid, "file size", &logutils.FieldArgs{"size": fileSize, "max": maxFileSize})
	}
	//TODO: add check for minimum file size (AWS recommends using PutObject for files <100MB in size)

	//TODO: evaluate parts calculation for performance, usability (consider client application memory capacity)
	partSize := minPartSize
	if fileSize > tib/2 {
		// 512 GiB (512) - 5 TiB (5120)
		partSize = 1 * gib
	} else if fileSize > 256*gib {
		// 256 GiB (128) - 512 GiB (1024)
		partSize = 512 * mib
	} else if fileSize > 64*gib {
		// 64 GiB (256) - 256 GiB (1024)
		partSize = 256 * mib
	} else if fileSize > 8*gib {
		// 8 GiB (128) - 64 GiB (1024)
		partSize = 64 * mib
	} else if fileSize > gib {
		// 1 GiB (32) - 8 GiB (256)
		partSize = 32 * mib
	} else if fileSize > 512*mib {
		// 512 MiB (32) - 1 GiB (64)
		partSize = 16 * mib
	} else if fileSize > minPartSize {
		// 5 MiB (1) - 512 MiB (64)
		partSize = 8 * mib
	}

	parts := fileSize / partSize
	if fileSize%partSize > 0 {
		parts++
	}

	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}

	result, err := s3.New(s).CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(a.config.S3Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "S3 multipart upload", &logutils.FieldArgs{"bucket": a.config.S3Bucket, "key": path}, err)
	}

	uploadID := "nil"
	if result.UploadId != nil {
		uploadID = *result.UploadId
	}
	signedURLs := make([]string, parts)
	for i := range parts {
		partReq, _ := s3.New(s).UploadPartRequest(&s3.UploadPartInput{
			Bucket:     aws.String(a.config.S3Bucket),
			Key:        aws.String(path),
			PartNumber: aws.Int64(int64(i + 1)),
			UploadId:   result.UploadId,
		})

		url, err := partReq.Presign(time.Duration(a.multipartUploadPresignExpirationMinutes) * time.Minute)
		if err != nil {
			a.logger.Warnf("error signing S3 upload part request for bucket %s, key %s, part number %d, upload_id %s: %s", a.config.S3Bucket, path, i+1, uploadID, err.Error())
			err = a.AbortMultipartUpload(path, uploadID, s)
			if err != nil {
				return nil, err
			}

			return nil, errors.WrapErrorAction("signing", "S3 upload part request", &logutils.FieldArgs{"bucket": a.config.S3Bucket, "key": path, "part": i + 1, "upload_id": uploadID}, err)
		}
		signedURLs[i] = url
	}

	upload := model.FileContentItemMultipartUpload{Key: fileKey, URLs: signedURLs, UploadID: uploadID}
	return &upload, nil
}

// CompleteMultipartUpload completes a multipart upload
func (a *Adapter) CompleteMultipartUpload(path string, uploadID string, eTags []string) error {
	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return err
	}

	parts := make([]*s3.CompletedPart, len(eTags))
	for i := range len(eTags) {
		parts[i] = &s3.CompletedPart{
			PartNumber: aws.Int64(int64(i + 1)),
			ETag:       aws.String(eTags[i]),
		}
	}
	_, err = s3.New(s).CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(a.config.S3Bucket),
		Key:             aws.String(path),
		UploadId:        aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: parts},
	})
	if err != nil {
		return errors.WrapErrorAction("completing", "S3 multipart upload", &logutils.FieldArgs{"bucket": a.config.S3Bucket, "key": path, "uploadID": uploadID}, err)
	}
	return nil
}

// AbortMultipartUpload aborts a multipart upload
func (a *Adapter) AbortMultipartUpload(path string, uploadID string, s *session.Session) error {
	var err error
	if s == nil {
		s, err = a.createS3Session(a.config.S3BucketAccelerate)
		if err != nil {
			log.Printf("Could not create S3 session")
			return err
		}
	}

	_, err = s3.New(s).AbortMultipartUpload(&s3.AbortMultipartUploadInput{
		Bucket:   aws.String(a.config.S3Bucket),
		Key:      aws.String(path),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		return errors.WrapErrorAction("aborting", "S3 multipart upload", &logutils.FieldArgs{"bucket": a.config.S3Bucket, "key": path, "uploadID": uploadID}, err)
	}
	return nil
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
func (a *Adapter) GetPresignedURLsForDownload(fileKeys, paths []string) ([]model.FileContentItemRef, error) {
	s, err := a.createS3Session(a.config.S3BucketAccelerate)
	if err != nil {
		log.Printf("Could not create S3 session")
		return nil, err
	}

	refs := make([]model.FileContentItemRef, len(paths))
	for i, path := range paths {
		req, _ := s3.New(s).GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(a.config.S3Bucket),
			Key:    aws.String(path),
		})
		url, err := req.Presign(time.Duration(a.downloadPresignExpirationMinutes) * time.Minute)
		if err != nil {
			return nil, err
		}
		refs[i] = model.FileContentItemRef{Key: fileKeys[i], URL: url}
	}
	return refs, nil
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
