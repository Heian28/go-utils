package gos3

import (
	"context"
	"mime/multipart"

	"github.com/Heian28/go-utils/gologger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
)

type GoS3Config struct {
	Region          string
	BaseEndpoint    string
	AccessKeyID     string
	SecretAccessKey string
	ProjectName     string
}

type gos3 struct {
	client *s3.Client
	log    *logrus.Logger
	conf   GoS3Config
}

var GoS3Client *gos3

func InitGoS3(
	ctx context.Context,
	conf GoS3Config,
	log *logrus.Logger,
) {
	if log == nil {
		gologger.New(
			gologger.SetServiceName("GoS3"),
		)
		log = gologger.Logger
	}

	cnf, er := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(conf.Region),
		config.WithBaseEndpoint(conf.BaseEndpoint),
	)
	if er != nil {
		log.Fatalf("[GoS3] Error loading default config: %+v", er)
	}

	client := s3.NewFromConfig(cnf, func(o *s3.Options) {
		o.UsePathStyle = true
		o.Credentials = credentials.NewStaticCredentialsProvider(
			conf.AccessKeyID,
			conf.SecretAccessKey,
			"",
		)
	})

	GoS3Client = &gos3{
		client: client,
		log:    log,
		conf:   conf,
	}
	log.Info("[GoS3] Initialized successfully...")
}

func (g *gos3) Client() *s3.Client {
	return g.client
}

type GoS3UploadFileResult struct {
	ID       string
	Location string
}

func (g *gos3) UploadFile(
	ctx context.Context,
	bucket string,
	key string,
	file multipart.File,
) (*GoS3UploadFileResult, error) {
	u := manager.NewUploader(g.client)
	uploadDest := g.conf.ProjectName
	if len(bucket) > 0 {
		uploadDest += bucket
	}
	res, err := u.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(uploadDest),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		g.log.Errorf("[GoS3] Error uploading file: %+v", err)
		return nil, err
	}

	g.log.Info("[GoS3] File uploaded successfully...")

	return &GoS3UploadFileResult{
		ID:       res.UploadID,
		Location: res.Location,
	}, nil
}
