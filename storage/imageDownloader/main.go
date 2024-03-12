package imageDownloader

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ImageDownloader struct {
	S3Client *s3.Client
	Bucket   string
}

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

func (d *ImageDownloader) DownloadAndUploadImage(ctx context.Context, imageUrl, fileName string) (string, error) {
	// 画像をダウンロードして一時ファイルに保存
	tempFilePath, err := downloadImageToTempFile(imageUrl)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFilePath) // Clean up the temp file later

	// 一時ファイルを開く
	file, err := os.Open(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("unable to open temp file: %w", err)
	}
	defer file.Close()

	// 一時ファイルをS3にアップロード
	return uploadImageToS3(ctx, d.S3Client, d.Bucket, fileName, file)
}

func downloadImageToTempFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-OK HTTP status code: %d", resp.StatusCode)
	}

	// 一時ファイルを作成
	tempFile, err := os.CreateTemp("", "image-*.jpg")
	if err != nil {
		return "", fmt.Errorf("unable to create temp file: %w", err)
	}
	defer tempFile.Close()

	// レスポンスボディをファイルに書き込む
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to write to temp file: %w", err)
	}

	return tempFile.Name(), nil
}

func uploadImageToS3(ctx context.Context, client *s3.Client, bucket, key string, body io.Reader) (string, error) {
	objectKey := fmt.Sprintf("images/%s", key)
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
		Body:   body,
	})
	if err != nil {
		return "", fmt.Errorf("upload error: %w", err)
	}

	log.Printf("Image uploaded successfully to %s/%s", bucket, key)
	return objectKey, nil
}
