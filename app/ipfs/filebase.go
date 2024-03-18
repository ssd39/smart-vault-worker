package ipfs

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/s3"
)

type FilebaseUploader struct {
	AccessKey string
	SecertKey string
	Bucket    string
	Name      string
}

func (uploader *FilebaseUploader) UploadBytes(data []byte) (string, error) {
	creds := credentials.NewStaticCredentials(uploader.AccessKey, uploader.SecertKey, "")
	cfg := aws.NewConfig().
		WithCredentials(creds).
		WithEndpoint("https://s3.filebase.com").
		WithRegion("us-east-1").
		WithS3ForcePathStyle(true)
	sess, err := session.NewSession(cfg)
	if err != nil {
		return "", err
	}

	// Create S3 service client
	svc := s3.New(sess)
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(uploader.Bucket),
		Key:    aws.String(uploader.Name),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return "", err
	}
	result, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(uploader.Bucket),
		Key:    aws.String(uploader.Name),
	})

	if err != nil {
		return "", err
	}

	return *result.Metadata["Cid"], nil
}
