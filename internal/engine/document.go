// Copyright (C) 2023 Tencent Cloud.
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the vectordb-sdk-java), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is furnished
// to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
// INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
// HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package engine

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/entity"
	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/entity/api/document"
)

var _ entity.DocumentInterface = &implementerDocument{}

type implementerDocument struct {
	entity.SdkClient
	database   entity.Database
	collection entity.Collection
}

// Upsert upsert documents into collection. Support for repeated insertion
func (i *implementerDocument) Upsert(ctx context.Context, documents []entity.Document, option *entity.UpsertDocumentOption) (result *entity.UpsertDocumentResult, err error) {
	req := new(document.UpsertReq)
	req.Database = i.database.DatabaseName
	req.Collection = i.collection.CollectionName
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

	if option != nil && option.BuildIndex != nil {
		req.BuildIndex = option.BuildIndex
	}

	res := new(document.UpsertRes)
	result = new(entity.UpsertDocumentResult)
	err = i.Request(ctx, req, res)
	if err != nil {
		return
	}
	result.AffectedCount = int(res.AffectedCount)
	return
}

// Query query the document by document ids.
// The parameters retrieveVector set true, will return the vector field, but will reduce the api speed.
func (i *implementerDocument) Query(ctx context.Context, documentIds []string, option *entity.QueryDocumentOption) (*entity.QueryDocumentResult, error) {
	req := new(document.QueryReq)
	req.Database = i.database.DatabaseName
	req.Collection = i.collection.CollectionName
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
	err := i.Request(ctx, req, res)
	if err != nil {
		return nil, err
	}

	result := new(entity.QueryDocumentResult)
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
	result.Documents = documents
	result.AffectedCount = len(documents)
	result.Total = res.Count
	return result, nil
}

// Search search document topK by vector. The optional parameters filter will add the filter condition to search.
// The optional parameters hnswParam only be set with the HNSW vector index type.
func (i *implementerDocument) Search(ctx context.Context, vectors [][]float32, option *entity.SearchDocumentOption) (*entity.SearchDocumentResult, error) {
	return i.search(ctx, nil, vectors, nil, option)
}

// Search search document topK by document ids. The optional parameters filter will add the filter condition to search.
// The optional parameters hnswParam only be set with the HNSW vector index type.
func (i *implementerDocument) SearchById(ctx context.Context, documentIds []string, option *entity.SearchDocumentOption) (*entity.SearchDocumentResult, error) {
	return i.search(ctx, documentIds, nil, nil, option)
}

func (i *implementerDocument) SearchByText(ctx context.Context, text map[string][]string, option *entity.SearchDocumentOption) (*entity.SearchDocumentResult, error) {
	return i.search(ctx, nil, nil, text, option)
}

func (i *implementerDocument) search(ctx context.Context, documentIds []string, vectors [][]float32, text map[string][]string, option *entity.SearchDocumentOption) (*entity.SearchDocumentResult, error) {
	req := new(document.SearchReq)
	req.Database = i.database.DatabaseName
	req.Collection = i.collection.CollectionName
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
	result := new(entity.SearchDocumentResult)
	result.Documents = documents
	return result, nil
}

// Delete delete document by document ids
func (i *implementerDocument) Delete(ctx context.Context, option *entity.DeleteDocumentOption) (result *entity.DeleteDocumentResult, err error) {
	req := new(document.DeleteReq)
	req.Database = i.database.DatabaseName
	req.Collection = i.collection.CollectionName
	if option != nil {
		req.Query = &document.QueryCond{
			DocumentIds: option.DocumentIds,
			Filter:      option.Filter.Cond(),
		}
	}

	res := new(document.DeleteRes)
	result = new(entity.DeleteDocumentResult)
	err = i.Request(ctx, req, res)
	if err != nil {
		return
	}
	result.AffectedCount = res.AffectedCount
	return
}

func (i *implementerDocument) Update(ctx context.Context, option *entity.UpdateDocumentOption) (*entity.UpdateDocumentResult, error) {
	req := new(document.UpdateReq)
	req.Database = i.database.DatabaseName
	req.Collection = i.collection.CollectionName
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
	result := new(entity.UpdateDocumentResult)
	err := i.Request(ctx, req, res)
	if err != nil {
		return result, err
	}
	result.AffectedCount = int(res.AffectedCount)
	return result, nil
}

func GetFieldInfo(field entity.Field) (string, entity.FieldType) {
	switch field.Val.(type) {
	case string:
		return field.String(), entity.String
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return strconv.FormatInt(field.Int(), 10), entity.Uint64
	}
	return "", entity.String
}

func getFileTypeFromFileName(fileName string) entity.FileType {
	extension := filepath.Ext(fileName)
	extension = strings.ToLower(extension)
	// 不带后缀的文件，默认为markdown文件
	if extension == "" {
		return entity.MarkdownFileType
	} else if extension == ".md" || extension == ".markdown" {
		return entity.MarkdownFileType
	} else {
		return entity.UnSupportFileType
	}
}

func isMarkdownFile(localFilePath string) bool {
	extension := filepath.Ext(localFilePath)
	extension = strings.ToLower(extension)
	return extension == ".md" || extension == ".markdown"
}

func checkFileSize(localFilePath string, maxContentLength int64) (bool, error) {
	fileInfo, err := os.Stat(localFilePath)
	if err != nil {
		return false, err
	}

	if fileInfo.Size() <= maxContentLength {
		return true, nil
	}
	return false, nil
}
