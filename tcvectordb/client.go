package tcvectordb

import (
	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/entry"
	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/internal/client"
	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/internal/engine"
	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/model"
)

func NewClient(url, username, key string, option *model.ClientOption) (entry.VectorDBClient, error) {
	sdkCli, err := client.NewClient(url, username, key, option)
	if err != nil {
		return nil, err
	}
	return engine.VectorDB(sdkCli), nil
}
