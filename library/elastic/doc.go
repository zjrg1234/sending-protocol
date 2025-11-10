package elastic

import (
	"encoding/json"
	"errors"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/mitchellh/mapstructure"
)

//所有doc都需要有一个id,所以要继承此类
type Doc struct {
	ID string `json:"id"`
}
type IDoc interface {
	GetID() string
	GetRouting() string
}

type BatchDoc[T IDoc] struct {
	DocList []IDoc
}

type Shards struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

//{"_index":"game_order_data_202212","_id":"PH3535726","_score":null,"_source":{"id":"PH3535726","or
type HitsDoc[T IDoc] struct {
	Index  string             `json:"_index"`
	ID     string             `json:"_id"`
	Source T                  `json:"_source"`
	Sort   []types.FieldValue `json:"sort"`
}

type Hits[T IDoc] struct {
	Hits []HitsDoc[T] `json:"hits"`
}

type ListResponse[T IDoc] struct {
	Took     int     `json:"took"`
	TimedOut bool    `json:"timed_out"`
	Shards   Shards  `json:"_shards"`
	Hits     Hits[T] `json:"hits"`
}

type AggregationsResponse[T any] struct {
	Took         int    `json:"took"`
	TimedOut     bool   `json:"timed_out"`
	Shards       Shards `json:"_shards"`
	Aggregations T      `json:"aggregations"`
	Error        any    `json:"error"`
}

type CountResponse struct {
	Count  int    `json:"count"`
	Shards Shards `json:"_shards"`
}

type SortValue struct {
	FirstSort []types.FieldValue `json:"first_sort"` //第一行的排序字段值,searchAfter值
	LastSort  []types.FieldValue `json:"last_sort"`  //最后一行的排序字段值,searchAfter值
}

// []types.FieldValue{}
type Result[T IDoc] struct {
	Took         int       `json:"took"`
	SerializedID string    `json:"serialized_id"`
	TimedOut     bool      `json:"timed_out"`
	Shards       Shards    `json:"_shards"`
	SortValue    SortValue `json:"sort_value"`
	PageSize     int       `json:"page_size"`
	PageNo       int       `json:"page_no"`
	PageSummary  any       `json:"page_summary"`
	List         *[]T      `json:"list"`
}

func ParseQueryResult[T IDoc](jsonStr string) (*Result[T], error) {
	var resp ListResponse[T]
	result := &Result[T]{}
	err := json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		return result, err
	}

	var list []T
	len := len(resp.Hits.Hits)
	if len > 0 {
		result.SortValue.FirstSort = resp.Hits.Hits[0].Sort
		result.SortValue.LastSort = resp.Hits.Hits[len-1].Sort
	}

	for _, item := range resp.Hits.Hits {
		list = append(list, item.Source)
	}

	result.Took = resp.Took
	result.Shards = resp.Shards
	result.TimedOut = resp.TimedOut
	result.List = &list
	return result, nil
}

func ParseAggregations[T any](jsonStr string) (T, error) {
	var t T
	var resp AggregationsResponse[map[string]any]
	err := json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		return t, err
	}

	if resp.Error != nil {
		return t, errors.New(jsonStr)
	}

	if len(resp.Aggregations) > 0 {
		aggrMap := make(map[string]any, 0)
		for key, aggr := range resp.Aggregations {
			value, ok := aggr.(map[string]any)
			if ok {
				aggrMap[key] = value["value"]
			}
		}
		mapstructure.Decode(aggrMap, &t)
	}
	return t, nil
}

func ParseAggregationResponse(jsonStr string) (AggregationsResponse[map[string]any], error) {
	var resp AggregationsResponse[map[string]any]
	err := json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		return resp, err
	}
	if resp.Error != nil {
		return resp, errors.New(jsonStr)
	}
	return resp, nil
}

func ParseQueryCount(jsonStr string) (int, error) {
	var resp CountResponse
	err := json.Unmarshal([]byte(jsonStr), &resp)
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}
