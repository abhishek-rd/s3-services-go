package s3service

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
)

type S3Service struct {
	sess *session.Session
}

func New() (*S3Service, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)

	if err != nil {
		return nil, err
	}

	return &S3Service{sess: sess}, nil
}

func (s *S3Service) GetObject(bucket string, key string) (interface{}, error) {
	svc := s3.New(s.sess)

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := svc.GetObject(input)
	if err != nil {
		return nil, err
	}

	defer result.Body.Close()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	var data interface{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *S3Service) UploadObject(bucket string, key string, file *multipart.FileHeader) error {
	uploader := s3manager.NewUploader(s.sess)

	openedFile, err := file.Open()
	if err != nil {
		return err
	}
	defer openedFile.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   openedFile,
	})
	return err
}
