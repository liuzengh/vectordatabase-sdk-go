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

package tcvectordb

import (
	"context"

	"git.woa.com/cloud_nosql/vectordb/vectordatabase-sdk-go/tcvectordb/api/index"
)

var _ IndexInterface = &implementerIndex{}

type IndexInterface interface {
	SdkClient
	RebuildIndex(ctx context.Context, options ...*RebuildIndexOption) (result *RebuildIndexResult, err error)
}

type implementerIndex struct {
	SdkClient
	database   Database
	collection Collection
}

func (i *implementerIndex) RebuildIndex(ctx context.Context, options ...*RebuildIndexOption) (*RebuildIndexResult, error) {
	if i.database.IsAIDatabase() {
		return nil, AIDbTypeError
	}
	req := new(index.RebuildReq)
	req.Database = i.database.DatabaseName
	req.Collection = i.collection.CollectionName

	if len(options) != 0 && options[0] != nil {
		option := options[0]
		req.DropBeforeRebuild = option.DropBeforeRebuild
		req.Throttle = int32(option.Throttle)
	}

	res := new(index.RebuildRes)
	err := i.Request(ctx, req, &res)
	if err != nil {
		return nil, err
	}
	result := new(RebuildIndexResult)
	result.TaskIds = res.TaskIds
	return result, nil
}
