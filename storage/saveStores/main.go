package saveStores

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Store struct {
	ID          string   `json:"id"`
	CompanyID   string   `json:"companyID"`
	StoreID     string   `json:"storeID"`
	StoreName   string   `json:"storeName"`
	Address     *string  `json:"address,omitempty"`
	Coordinates *string  `json:"coordinates,omitempty"`
	Details     *string  `json:"details,omitempty"`
	Images      []string `json:"images,omitempty"`
	CrawledUrl  string   `json:"crawledUrl"`
}

// CreateDynamoDBClient は指定されたリージョン用の新しいDynamoDBクライアントを作成して返します。
func CreateDynamoDBClient(ctx context.Context, awsRegion string) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}
	return dynamodb.NewFromConfig(cfg), nil
}

// convertStoreToDynamoDBItem はStore構造体をDynamoDBのアイテムフォーマットに変換します。
func convertStoreToDynamoDBItem(store *Store) map[string]types.AttributeValue {
	item := map[string]types.AttributeValue{
		"id":         &types.AttributeValueMemberS{Value: store.ID},
		"companyID":  &types.AttributeValueMemberS{Value: store.CompanyID},
		"storeID":    &types.AttributeValueMemberS{Value: store.StoreID},
		"storeName":  &types.AttributeValueMemberS{Value: store.StoreName},
		"crawledUrl": &types.AttributeValueMemberS{Value: store.CrawledUrl},
		"createDate": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)}, // 作成日時
		"updateDate": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)}, // 更新日時
	}

	// オプショナルフィールド
	if store.Address != nil {
		item["address"] = &types.AttributeValueMemberS{Value: *store.Address}
	}
	if store.Coordinates != nil {
		item["coordinates"] = &types.AttributeValueMemberS{Value: *store.Coordinates}
	}
	if store.Details != nil {
		item["details"] = &types.AttributeValueMemberS{Value: *store.Details}
	}
	if len(store.Images) > 0 {
		imageValues := make([]string, len(store.Images))
		for i, img := range store.Images {
			imageValues[i] = img
		}
		item["images"] = &types.AttributeValueMemberSS{Value: imageValues}
	}

	return item
}

// SaveStore はDynamoDBに単一の店舗情報を保存します。
func SaveStore(ctx context.Context, client *dynamodb.Client, tableName string, store *Store) error {
	item := convertStoreToDynamoDBItem(store)

	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put item into DynamoDB: %w", err)
	}

	return nil
}
