package engine

import (
	"context"
	"strings"

	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/entity"
	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/internal/client"
	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/internal/engine/api/database"
)

var _ entity.DatabaseInterface = &implementerDatabase{}

type implementerDatabase struct {
	entity.SdkClient
}

// VectorDB new a vectordbClient interface
func VectorDB(sdkClient *client.Client) entity.VectorDBClient {
	databaseImpl := new(implementerDatabase)
	databaseImpl.SdkClient = sdkClient
	return databaseImpl
}

// CreateDatabase create database with database name. It returns error if name exist.
func (i *implementerDatabase) CreateDatabase(ctx context.Context, name string, option *entity.CreateDatabaseOption) (*entity.Database, error) {
	req := database.CreateReq{
		Database: name,
	}
	res := new(database.CreateRes)
	err := i.Request(ctx, req, res)
	if err != nil {
		return nil, err
	}
	return i.Database(name), err
}

// DropDatabase drop database with database name. If database not exist, it return nil.
func (i *implementerDatabase) DropDatabase(ctx context.Context, name string, option *entity.DropDatabaseOption) (result *entity.DatabaseResult, err error) {
	result = new(entity.DatabaseResult)

	req := database.DropReq{Database: name}
	res := new(database.DropRes)
	err = i.Request(ctx, req, res)
	if err != nil {
		if strings.Contains(err.Error(), "not exist") {
			return result, nil
		}
		return
	}
	result.AffectedCount = int(res.AffectedCount)
	return
}

// ListDatabase get database list. It returns the database list to operate the collection.
func (i *implementerDatabase) ListDatabase(ctx context.Context, option *entity.ListDatabaseOption) (databases []*entity.Database, err error) {
	req := database.ListReq{}
	res := new(database.ListRes)
	err = i.Request(ctx, req, res)
	if err != nil {
		return nil, err
	}

	for _, v := range res.Databases {
		databases = append(databases, i.Database(v))
	}
	return
}

// Database get a database interface to operate collection.  It could not send http request to vectordb.
func (i *implementerDatabase) Database(name string) *entity.Database {
	database := new(entity.Database)
	collImpl := new(implementerCollection)
	collImpl.SdkClient = i.SdkClient
	collImpl.databaseName = name
	database.CollectionInterface = collImpl

	aliasImpl := new(implementerAlias)
	aliasImpl.databaseName = name
	aliasImpl.SdkClient = i.SdkClient
	database.AliasInterface = aliasImpl

	indexImpl := new(implementerIndex)
	indexImpl.databaseName = name
	indexImpl.SdkClient = i.SdkClient
	database.IndexInterface = indexImpl

	database.DatabaseName = name
	return database
}
