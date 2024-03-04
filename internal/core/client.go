package core

import (
	"context"
	"fmt"

	cloud "github.com/G-Core/gcore-cloud-sdk-go"
)

func CloudClient(ctx context.Context) (*cloud.ClientWithResponses, error) {
	var client *cloud.ClientWithResponses

	profile, err := GetClientProfile(ctx)
	if err != nil {
		return nil, err
	}
	baseUrl := *profile.ApiUrl
	authFunc := ExtractAuthFunc(ctx)

	if !profile.IsLocal() {
		baseUrl += "/cloud"
	}

	client, err = cloud.NewClientWithResponses(baseUrl, cloud.WithRequestEditorFn(authFunc))
	if err != nil {
		return nil, fmt.Errorf("cannot init SDK: %w", err)
	}

	return client, nil
}
