package engine

import (
	"context"

	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/internal/engine/api/document"
	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/model"
)

type implementerDocument struct {
	model.SdkClient
	databaseName   string
	collectionName string
}

// Upsert upsert documents into collection. Support for repeated insertion
func (i *implementerDocument) Upsert(ctx context.Context, documents []model.Document, buidIndex bool) (err error) {
	req := new(document.UpsertReq)
	req.Database = i.databaseName
	req.Collection = i.collectionName
	req.BuildIndex = buidIndex
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

	res := new(document.UpsertRes)
	err = i.Request(ctx, req, res)
	return
}

// Query query the document by document ids. The parameters retrieveVector set true, will return the vector field, but will reduce the api speed.
func (i *implementerDocument) Query(ctx context.Context, documentIds []string, filter *model.Filter, readConsistency string, retrieveVector bool, outputFields []string, offset, limit int64) ([]model.Document, uint64, error) {
	req := new(document.QueryReq)
	req.Database = i.databaseName
	req.Collection = i.collectionName
	req.Query = &document.QueryCond{
		DocumentIds:    documentIds,
		RetrieveVector: retrieveVector,
		Offset:         offset,
		Limit:          limit,
		OutputFields:   outputFields,
	}
	if filter != nil {
		req.Query.Filter = filter.Cond()
	}
	res := new(document.QueryRes)
	err := i.Request(ctx, req, res)
	if err != nil {
		return nil, 0, err
	}
	var documents []model.Document
	for _, doc := range res.Documents {
		var d model.Document
		d.Id = doc.Id
		d.Vector = doc.Vector
		d.Fields = make(map[string]model.Field)

		for n, v := range doc.Fields {
			d.Fields[n] = model.Field{Val: v}
		}
		documents = append(documents, d)
	}
	return documents, res.Count, nil
}

// Search search document topK by vector. The optional parameters filter will add the filter condition to search.
// The optional parameters hnswParam only be set with the HNSW vector index type.
func (i *implementerDocument) Search(ctx context.Context, vectors [][]float32, filter *model.Filter, readConsistency model.ReadConsistency, searchParam *model.SearchParams, retrieveVector bool, outputFields []string, limit int) ([][]model.Document, error) {
	return i.search(ctx, nil, vectors, nil, filter, readConsistency, searchParam, retrieveVector, outputFields, limit)
}

// Search search document topK by document ids. The optional parameters filter will add the filter condition to search.
// The optional parameters hnswParam only be set with the HNSW vector index type.
func (i *implementerDocument) SearchById(ctx context.Context, documentIds []string, filter *model.Filter, readConsistency model.ReadConsistency, searchParam *model.SearchParams, retrieveVector bool, outputFields []string, limit int) ([][]model.Document, error) {
	return i.search(ctx, documentIds, nil, nil, filter, readConsistency, searchParam, retrieveVector, outputFields, limit)
}

func (i *implementerDocument) SearchByText(ctx context.Context, text map[string][]string, filter *model.Filter, readConsistency model.ReadConsistency, searchParam *model.SearchParams, retrieveVector bool, outputFields []string, limit int) ([][]model.Document, error) {
	return i.search(ctx, nil, nil, text, filter, readConsistency, searchParam, retrieveVector, outputFields, limit)
}

func (i *implementerDocument) search(ctx context.Context, documentIds []string, vectors [][]float32, text map[string][]string, filter *model.Filter, readConsistency model.ReadConsistency, searchParam *model.SearchParams, retrieveVector bool, outputFields []string, limit int) ([][]model.Document, error) {
	req := new(document.SearchReq)
	req.Database = i.databaseName
	req.Collection = i.collectionName
	req.Search = new(document.SearchCond)
	req.Search.DocumentIds = documentIds
	req.Search.Vectors = vectors
	if filter != nil {
		req.Search.Filter = filter.Cond()
	}
	req.ReadConsistency = string(readConsistency)
	req.Search.RetrieveVector = retrieveVector
	req.Search.Limit = uint32(limit)
	if searchParam != nil {
		req.Search.Params = &document.SearchParams{
			Ef:     searchParam.Ef,
			Nprobe: searchParam.Nprobe,
			Radius: searchParam.Radius,
		}
	}
	req.Search.Outputfields = outputFields
	for _, v := range text {
		req.Search.EmbeddingItems = v
	}

	res := new(document.SearchRes)
	err := i.Request(ctx, req, res)
	if err != nil {
		return nil, err
	}
	var documents [][]model.Document
	for _, result := range res.Documents {
		var vecDoc []model.Document
		for _, doc := range result {
			d := model.Document{
				Id:     doc.Id,
				Vector: doc.Vector,
				Score:  doc.Score,
				Fields: make(map[string]model.Field),
			}
			for n, v := range doc.Fields {
				d.Fields[n] = model.Field{Val: v}
			}
			vecDoc = append(vecDoc, d)
		}
		documents = append(documents, vecDoc)
	}

	return documents, nil
}

// Delete delete document by document ids
func (i *implementerDocument) Delete(ctx context.Context, documentIds []string, filter *model.Filter) (err error) {
	req := new(document.DeleteReq)
	req.Database = i.databaseName
	req.Collection = i.collectionName
	req.Query = &document.QueryCond{
		DocumentIds: documentIds,
		Filter:      filter.Cond(),
	}

	res := new(document.DeleteRes)
	err = i.Request(ctx, req, res)
	return
}

func (i *implementerDocument) Update(ctx context.Context, queryIds []string, queryFilter *model.Filter, updateVector []float32, updateFields map[string]model.Field) (uint64, error) {
	req := new(document.UpdateReq)
	req.Database = i.databaseName
	req.Collection = i.collectionName
	req.Query = &document.QueryCond{
		DocumentIds: queryIds,
	}
	if queryFilter != nil {
		req.Query.Filter = queryFilter.Cond()
	}
	req.Update.Vector = updateVector
	req.Update.Fields = make(map[string]interface{})
	for k, v := range updateFields {
		req.Update.Fields[k] = v.Val
	}

	res := new(document.UpdateRes)
	err := i.Request(ctx, req, res)
	if err != nil {
		return 0, err
	}
	return res.AffectedCount, nil
}
