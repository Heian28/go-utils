package gominio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Heian28/go-utils/gologger"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

// GoMinioConfig is a struct that represents the configuration for the GoMinio client.
type GoMinioConfig struct {
	Endpoint    string `mapstructure:"endpoint"`
	AccessKey   string `mapstructure:"accessKey"`
	SecretKey   string `mapstructure:"secretKey"`
	Location    string `mapstructure:"location"`
	ProjectName string `mapstructure:"projectName"`
	UseSSL      bool   `mapstructure:"useSSL"`
}

// gminio is a struct that represents the GoMinio client.
type gminio struct {
	client *minio.Client
	log    *logrus.Logger
	conf   GoMinioConfig
}

// GoMinioClient is a global variable that represents the GoMinio client.
var GoMinioClient *gminio

// InitGoMinio is a function that initializes the GoMinio client.
func InitGoMinio(config GoMinioConfig, log *logrus.Logger) {
	if log == nil {
		gologger.New(
			gologger.SetServiceName("GoMinio"),
		)
		log = gologger.Logger
	}
	c, er := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if er != nil {
		log.Fatalf("[GoMinio] Error initializing Minio client: %+v", er)
	}
	log.Infof("[GoMinio] Minio client initialized successfully...")
	GoMinioClient = &gminio{
		client: c,
		log:    log,
		conf:   config,
	}
}

// UploadFileResponse is a struct that represents the response from the UploadFile function.
type UploadFileResponse struct {
	Key      string
	Location string
}

// ExtractedFileInfo is a struct that represents the extracted file info.
type ExtractedFileInfo struct {
	ID       string
	Filename string
	Ext      string
	MimeType string
	Size     int64
	Buffer   []byte
}

// UploadFile is a function that uploads a file to Minio.
func (g *gminio) UploadFile(
	ctx context.Context,
	location string,
	key string,
	fileBuffer []byte,
	fileSize int64,
	contentType string,
) (*UploadFileResponse, error) {
	bucket := g.conf.ProjectName
	exists, err := g.checkBucket(ctx, bucket)
	if err != nil {
		g.log.Errorf("[GoMinio] Error checking bucket: %+v", err)
		return nil, err
	}
	if !exists {
		g.log.Errorf("[GoMinio] Bucket not found: %+v", bucket)
		return nil, errors.New("bucket not found")
	}
	dest := fmt.Sprintf("%s/%s", location, key)
	info, err := g.client.PutObject(
		ctx,
		bucket,
		dest,
		bytes.NewReader(fileBuffer),
		fileSize,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		g.log.Errorf("[GoMinio] Error uploading file: %+v", err)
		return nil, err
	}
	g.log.Infof("[GoMinio] File uploaded successfully: %+v", info)
	info.Location = fmt.Sprintf("/%s/%s", g.conf.ProjectName, dest)
	return &UploadFileResponse{
		Key:      info.Key,
		Location: info.Location,
	}, nil
}

// GetFile is a function that gets a file from Minio.
// It returns a pointer to the file and an error if the file is not found.
// If the file is not found, it returns nil, nil.
// If the file is found, it returns the file and nil.
func (g *gminio) GetFile(
	ctx context.Context,
	location string,
	key string,
) (*os.File, error) {
	bucket := g.conf.ProjectName
	dest := fmt.Sprintf("%s/%s", location, key)
	obj, err := g.client.GetObject(ctx, bucket, dest, minio.GetObjectOptions{})
	if err != nil {
		g.log.Errorf("[GoMinio] Error getting file: %+v", err)
		return nil, err
	}
	defer obj.Close()

	var buff bytes.Buffer
	_, err = buff.ReadFrom(obj)
	if err != nil {
		g.log.Errorf("[GoMinio] Error reading file: %+v", err)
		return nil, nil
	}

	var file *os.File
	file, err = os.CreateTemp("", key)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	_, err = file.Write(buff.Bytes())
	if err != nil {
		return nil, err
	}

	return file, nil
}

// ExtractFileInfo is a function that extracts the file info from a file.
func (g *gminio) ExtractFileInfo(f *os.File) (*ExtractedFileInfo, error) {
	if f == nil {
		return nil, errors.New("file is nil")
	}
	file, err := f.Stat()
	if err != nil {
		g.log.Errorf("[GoMinio] Error getting file info: %+v", err)
		return nil, err
	}
	size := file.Size()
	buffer := make([]byte, size)
	_, err = f.Read(buffer)
	if err != nil {
		g.log.Errorf("[GoMinio] Error reading file: %+v", err)
		return nil, err
	}
	fn := f.Name()
	ext := filepath.Ext(fn)
	ct := http.DetectContentType(buffer)
	return &ExtractedFileInfo{
		ID:       uuid.New().String(),
		Filename: fn,
		Ext:      ext,
		MimeType: ct,
		Size:     size,
		Buffer:   buffer,
	}, nil
}

// MExtractFileInfo is a function that extracts the file info from a multipart file.
func (g *gminio) MExtractFileInfo(f *multipart.FileHeader) (*ExtractedFileInfo, error) {
	fn := f.Filename
	fs := f.Size
	ext := filepath.Ext(fn)
	ct := f.Header.Get("Content-Type")
	buffer := make([]byte, fs)
	file, err := f.Open()
	if err != nil {
		g.log.Errorf("[GoMinio] Error opening file: %+v", err)
		return nil, err
	}
	defer file.Close()
	_, err = file.Read(buffer)
	if err != nil {
		g.log.Errorf("[GoMinio] Error reading file: %+v", err)
		return nil, err
	}
	return &ExtractedFileInfo{
		ID:       uuid.New().String(),
		Filename: fn,
		Ext:      ext,
		MimeType: ct,
		Size:     fs,
		Buffer:   buffer,
	}, nil
}

// checkBucket is a function that checks if a bucket exists.
func (g *gminio) checkBucket(
	ctx context.Context,
	bucketName string,
) (bool, error) {
	exists, err := g.client.BucketExists(ctx, bucketName)
	if err != nil {
		g.log.Errorf("[GoMinio] Error checking bucket: %+v", err)
		return false, err
	}
	return exists, nil
}
