package saveStores

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Store struct {
	CompanyID   string     `json:"companyID"`
	StoreID     string     `json:"storeID"`
	StoreName   string     `json:"storeName"`
	Address     *string    `json:"address,omitempty"`
	Coordinates *string    `json:"coordinates,omitempty"`
	Details     *string    `json:"details,omitempty"`
	Images      []string   `json:"images,omitempty"`
	CrawledUrl  string     `json:"crawledUrl"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
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
		"companyID":  &types.AttributeValueMemberS{Value: store.CompanyID},
		"storeID":    &types.AttributeValueMemberS{Value: store.StoreID},
		"storeName":  &types.AttributeValueMemberS{Value: store.StoreName},
		"crawledUrl": &types.AttributeValueMemberS{Value: store.CrawledUrl},
		"createdAt":  &types.AttributeValueMemberS{Value: store.CreatedAt.Format(time.RFC3339)},
		"updatedAt":  &types.AttributeValueMemberS{Value: store.UpdatedAt.Format(time.RFC3339)},
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
	if store.DeletedAt != nil {
		item["deletedAt"] = &types.AttributeValueMemberS{Value: store.DeletedAt.Format(time.RFC3339)}
	}

	return item
}

func BatchSaveStores(ctx context.Context, client *dynamodb.Client, tableName string, stores []*Store) error {
	// DynamoDBにバッチで書き込みます。1回のバッチにつき最大25項目まで。
	for i := 0; i < len(stores); i += 25 {
		end := i + 25
		if end > len(stores) {
			end = len(stores)
		}

		writeRequests := make([]types.WriteRequest, 0, len(stores[i:end]))
		for _, petDetail := range stores[i:end] {
			item := convertStoreToDynamoDBItem(petDetail)
			writeRequests = append(writeRequests, types.WriteRequest{
				PutRequest: &types.PutRequest{Item: item},
			})
		}

		_, err := client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{tableName: writeRequests},
		})
		if err != nil {
			return fmt.Errorf("failed to batch write items: %v", err)
		}
	}

	return nil
}
