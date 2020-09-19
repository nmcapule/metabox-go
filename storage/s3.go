package storage

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/nmcapule/metabox-go/config"
)

type S3 struct {
	config  *config.S3StorageConfig
	session *session.Session
}

func NewS3(config *config.S3StorageConfig) (*S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      &config.Region,
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		Endpoint:    &config.Endpoint,
	})
	if err != nil {
		return nil, err
	}
	return &S3{
		config:  config,
		session: sess,
	}, nil
}

func (s *S3) Exists(key string) (bool, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(s.config.PrefixPath + key),
	}
	_, err := s3.New(s.session).GetObject(input)
	if err != nil {
		return false, fmt.Errorf("GetObject(%+v): %v", input, err)
	}
	return true, nil
}

func (s *S3) Upload(key string, source io.Reader) error {
	uploader := s3manager.NewUploader(s.session)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(s.config.PrefixPath + key),
		Body:   source,
	})
	if err != nil {
		return fmt.Errorf("upload %q: %v", key, err)
	}
	return nil
}

func (s *S3) Download(key string, destination io.Writer) error {
	downloader := s3manager.NewDownloader(s.session)
	buf := aws.NewWriteAtBuffer([]byte{})
	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(s.config.PrefixPath + key),
	})
	if err != nil {
		return fmt.Errorf("download %q: %v", key, err)
	}
	if _, err := io.Copy(destination, bytes.NewBuffer(buf.Bytes())); err != nil {
		return fmt.Errorf("copy to destination: %v", err)
	}
	return nil
}
