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

// DownloadAndUploadImage downloads an image from a URL and uploads it to an S3 bucket.
func DownloadAndUploadImage(ctx context.Context, imageUrl string, fileName string, bucketName string) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("ap-northeast-1"))
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %w", err)
	}

	if bucketName == "" {
		return fmt.Errorf("FILE_BUCKET_NAME environment variable not set")
	}

	resp, err := downloadImage(imageUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return uploadImageToS3(ctx, s3.NewFromConfig(cfg), bucketName, fileName, resp.Body)
}

// downloadImage performs an HTTP GET request for the given URL and returns the response.
func downloadImage(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
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
