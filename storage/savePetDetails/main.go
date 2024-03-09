package savePetDetails

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type PetDetail struct {
	ID          string   `json:"id"`
	CompanyID   string   `json:"companyID"`
	StoreID     string   `json:"storeID"`
	PetID       string   `json:"petID"`
	PetType     string   `json:"petType"`
	Type        string   `json:"type"`
	PriceExTax  float64  `json:"priceExTax"`
	PriceIncTax float64  `json:"priceIncTax"`
	Father      *string  `json:"father"`
	Mother      *string  `json:"mother"`
	Color       *string  `json:"color"`
	Origin      *string  `json:"origin"`
	Sex         *string  `json:"sex"`
	Birthdate   *string  `json:"birthdate"`
	Images      []string `json:"images"`
	CrawledUrl  string   `json:"crawledUrl"`
	CreateDate  string   `json:"createDate"`
	UpdateDate  string   `json:"updateDate"`
}

func CreateDynamoDBClient(ctx context.Context, awsRegion string) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}
	return dynamodb.NewFromConfig(cfg), nil
}

func convertPetDetailToDynamoDBItem(petDetail *PetDetail) map[string]types.AttributeValue {
	item := map[string]types.AttributeValue{
		"id":          &types.AttributeValueMemberS{Value: petDetail.ID},
		"companyID":   &types.AttributeValueMemberS{Value: petDetail.CompanyID},
		"storeID":     &types.AttributeValueMemberS{Value: petDetail.StoreID},
		"petID":       &types.AttributeValueMemberS{Value: petDetail.PetID},
		"petType":     &types.AttributeValueMemberS{Value: petDetail.PetType},
		"type":        &types.AttributeValueMemberS{Value: petDetail.Type},
		"priceExTax":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", petDetail.PriceExTax)},
		"priceIncTax": &types.AttributeValueMemberN{Value: fmt.Sprintf("%.2f", petDetail.PriceIncTax)},
		"crawledUrl":  &types.AttributeValueMemberS{Value: petDetail.CrawledUrl},
		"createDate":  &types.AttributeValueMemberS{Value: petDetail.CreateDate},
		"updateDate":  &types.AttributeValueMemberS{Value: petDetail.UpdateDate},
	}

	if petDetail.Father != nil {
		item["father"] = &types.AttributeValueMemberS{Value: *petDetail.Father}
	}

	if petDetail.Mother != nil {
		item["mother"] = &types.AttributeValueMemberS{Value: *petDetail.Mother}
	}

	if petDetail.Color != nil {
		item["color"] = &types.AttributeValueMemberS{Value: *petDetail.Color}
	}

	if petDetail.Origin != nil {
		item["origin"] = &types.AttributeValueMemberS{Value: *petDetail.Origin}
	}

	if petDetail.Sex != nil {
		item["sex"] = &types.AttributeValueMemberS{Value: *petDetail.Sex}
	}

	if petDetail.Birthdate != nil {
		item["birthdate"] = &types.AttributeValueMemberS{Value: *petDetail.Birthdate}
	}

	if len(petDetail.Images) > 0 {
		ssValues := make([]string, len(petDetail.Images))
		for i, v := range petDetail.Images {
			ssValues[i] = v
		}
		item["images"] = &types.AttributeValueMemberSS{Value: ssValues}
	}

	return item
}

func BatchSavePetDetails(ctx context.Context, client *dynamodb.Client, tableName string, petDetails []*PetDetail) error {
	// DynamoDBにバッチで書き込みます。1回のバッチにつき最大25項目まで。
	for i := 0; i < len(petDetails); i += 25 {
		end := i + 25
		if end > len(petDetails) {
			end = len(petDetails)
		}

		writeRequests := make([]types.WriteRequest, 0, len(petDetails[i:end]))
		for _, petDetail := range petDetails[i:end] {
			item := convertPetDetailToDynamoDBItem(petDetail)
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