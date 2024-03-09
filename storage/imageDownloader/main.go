package imageDownloader

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// NOTE: HTTPレスポンスのボディを直接アップロードしようとしたところエラーが発生したので、一時ファイルを作成してそれをアップロードすることで解決
// {
//   "errorMessage": "operation error S3: PutObject, https response error StatusCode: 501, RequestID: XXX, HostID: YYY, api error NotImplemented: A header you provided implies functionality that is not implemented",
//   "errorType": "OperationError"
// }

// ImageDownloader is a struct to hold necessary AWS clients and configurations.
type ImageDownloader struct {
	S3Client *s3.Client
	Bucket   string
}

// NewImageDownloader creates a new ImageDownloader instance.
func NewImageDownloader(ctx context.Context, bucketName string, awsRegion string) (*ImageDownloader, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name is not set")
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &ImageDownloader{
		S3Client: s3.NewFromConfig(cfg),
		Bucket:   bucketName,
	}, nil
}

// DownloadAndUploadImage downloads an image from a URL and uploads it to the configured S3 bucket.
func (d *ImageDownloader) DownloadAndUploadImage(ctx context.Context, imageUrl, fileName string) error {
	resp, err := downloadImage(imageUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return uploadImageToS3(ctx, d.S3Client, d.Bucket, fileName, resp.Body)
}

// downloadImage performs an HTTP GET request for the given URL and returns the response.
func downloadImage(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download error: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close() // Close the body to avoid resource leaks
		return nil, fmt.Errorf("non-OK HTTP status code: %d", resp.StatusCode)
	}

	return resp, nil
}

// uploadImageToS3 uploads the given image (as an io.Reader) to the specified S3 bucket and key.
func uploadImageToS3(ctx context.Context, client *s3.Client, bucket, key string, body io.Reader) error {
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fmt.Sprintf("images/%s", key)),
		Body:   body,
	})
	if err != nil {
		return fmt.Errorf("upload error: %w", err)
	}

	log.Printf("Image uploaded successfully to %s/%s", bucket, key)
	return nil
}
