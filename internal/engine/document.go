package engine

import (
	"context"

	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/entity"
	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/internal/engine/api/document"
)

var _ entity.DocumentInterface = &implementerDocument{}

type implementerDocument struct {
	entity.SdkClient
	databaseName   string
	collectionName string
}

// Upsert upsert documents into collection. Support for repeated insertion
func (i *implementerDocument) Upsert(ctx context.Context, documents []entity.Document, option *entity.UpsertDocumentOption) (result *entity.DocumentResult, err error) {
	req := new(document.UpsertReq)
	req.Database = i.databaseName
	req.Collection = i.collectionName
	for _, doc := range documents {
		d := &document.Document{}
		d.Id = doc.Id
		d.Vector = doc.Vector
		d.Fields = make(map[string]interface{})
		for k, v := range doc.Fields {
			d.Fields[k] = v.Val
		}
		req.Documents = append(req.Documents, d)
	}

	if option != nil {
		req.BuildIndex = option.BuildIndex
	}

	res := new(document.UpsertRes)
	result = new(entity.DocumentResult)
	err = i.Request(ctx, req, res)
	if err != nil {
		return
	}
	result.AffectedCount = int(res.AffectedCount)
	return
}

// Query query the document by document ids. The parameters retrieveVector set true, will return the vector field, but will reduce the api speed.
func (i *implementerDocument) Query(ctx context.Context, documentIds []string, option *entity.QueryDocumentOption) ([]entity.Document, *entity.DocumentResult, error) {
	req := new(document.QueryReq)
	req.Database = i.databaseName
	req.Collection = i.collectionName
	req.Query = &document.QueryCond{
		DocumentIds: documentIds,
	}
	req.ReadConsistency = string(i.SdkClient.Options().ReadConsistency)
	if option != nil {
		req.Query.Filter = option.Filter.Cond()
		req.Query.RetrieveVector = option.RetrieveVector
		req.Query.OutputFields = option.OutputFields
		req.Query.Offset = option.Offset
		req.Query.Limit = option.Limit
	}

	res := new(document.QueryRes)
	result := new(entity.DocumentResult)
	err := i.Request(ctx, req, res)
	if err != nil {
		return nil, result, err
	}
	var documents []entity.Document
	for _, doc := range res.Documents {
		var d entity.Document
		d.Id = doc.Id
		d.Vector = doc.Vector
		d.Fields = make(map[string]entity.Field)

		for n, v := range doc.Fields {
			d.Fields[n] = entity.Field{Val: v}
		}
		documents = append(documents, d)
	}
	result.AffectedCount = len(documents)
	result.Total = int(res.Count)
	return documents, result, nil
}

// Search search document topK by vector. The optional parameters filter will add the filter condition to search.
// The optional parameters hnswParam only be set with the HNSW vector index type.
func (i *implementerDocument) Search(ctx context.Context, vectors [][]float32, option *entity.SearchDocumentOption) ([][]entity.Document, error) {
	return i.search(ctx, nil, vectors, nil, option)
}

// Search search document topK by document ids. The optional parameters filter will add the filter condition to search.
// The optional parameters hnswParam only be set with the HNSW vector index type.
func (i *implementerDocument) SearchById(ctx context.Context, documentIds []string, option *entity.SearchDocumentOption) ([][]entity.Document, error) {
	return i.search(ctx, documentIds, nil, nil, option)
}

func (i *implementerDocument) SearchByText(ctx context.Context, text map[string][]string, option *entity.SearchDocumentOption) ([][]entity.Document, error) {
	return i.search(ctx, nil, nil, text, option)
}

func (i *implementerDocument) search(ctx context.Context, documentIds []string, vectors [][]float32, text map[string][]string, option *entity.SearchDocumentOption) ([][]entity.Document, error) {
	req := new(document.SearchReq)
	req.Database = i.databaseName
	req.Collection = i.collectionName
	req.ReadConsistency = string(i.SdkClient.Options().ReadConsistency)
	req.Search = new(document.SearchCond)
	req.Search.DocumentIds = documentIds
	req.Search.Vectors = vectors
	for _, v := range text {
		req.Search.EmbeddingItems = v
	}

	if option != nil {
		req.Search.Filter = option.Filter.Cond()
		req.Search.RetrieveVector = option.RetrieveVector
		req.Search.OutputFields = option.OutputFields
		req.Search.Limit = option.Limit

		if option.Params != nil {
			req.Search.Params = new(document.SearchParams)
			req.Search.Params.Nprobe = option.Params.Nprobe
			req.Search.Params.Ef = option.Params.Ef
			req.Search.Params.Radius = option.Params.Radius
		}
	}

	res := new(document.SearchRes)
	err := i.Request(ctx, req, res)
	if err != nil {
		return nil, err
	}
	var documents [][]entity.Document
	for _, result := range res.Documents {
		var vecDoc []entity.Document
		for _, doc := range result {
			d := entity.Document{
				Id:     doc.Id,
				Vector: doc.Vector,
				Score:  doc.Score,
				Fields: make(map[string]entity.Field),
			}
			for n, v := range doc.Fields {
				d.Fields[n] = entity.Field{Val: v}
			}
			vecDoc = append(vecDoc, d)
		}
		documents = append(documents, vecDoc)
	}

	return documents, nil
}

// Delete delete document by document ids
func (i *implementerDocument) Delete(ctx context.Context, option *entity.DeleteDocumentOption) (result *entity.DocumentResult, err error) {
	req := new(document.DeleteReq)
	req.Database = i.databaseName
	req.Collection = i.collectionName
	if option != nil {
		req.Query = &document.QueryCond{
			DocumentIds: option.DocumentIds,
			Filter:      option.Filter.Cond(),
		}
	}

	res := new(document.DeleteRes)
	result = new(entity.DocumentResult)
	err = i.Request(ctx, req, res)
	if err != nil {
		return
	}
	result.AffectedCount = int(res.AffectedCount)
	return
}

func (i *implementerDocument) Update(ctx context.Context, option *entity.UpdateDocumentOption) (*entity.DocumentResult, error) {
	req := new(document.UpdateReq)
	req.Database = i.databaseName
	req.Collection = i.collectionName
	req.Query = new(document.QueryCond)

	if option != nil {
		req.Query.DocumentIds = option.QueryIds
		req.Query.Filter = option.QueryFilter.Cond()
		req.Update.Vector = option.UpdateVector
		req.Update.Fields = make(map[string]interface{})
		if len(option.UpdateFields) != 0 {
			for k, v := range option.UpdateFields {
				req.Update.Fields[k] = v.Val
			}
		}
	}

	res := new(document.UpdateRes)
	result := new(entity.DocumentResult)
	err := i.Request(ctx, req, res)
	if err != nil {
		return result, err
	}
	result.AffectedCount = int(res.AffectedCount)
	return result, nil
}
