package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/count"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"go.uber.org/zap"
	"io/ioutil"
	"megin/library/logger"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var once sync.Once
var elasticClient *ElasticClient

type ElasticClient struct {
	//老版api client
	Client *elasticsearch.Client

	//	新版本api,多使用下面这个client
	//	参考如下: https://www.elastic.co/guide/en/elasticsearch/client/go-api/8.5/examples.html#search
	TypedClient *elasticsearch.TypedClient
}

func GetElasticClient() *ElasticClient {
	return elasticClient
}

func InitElasticClient(addrs []string, username, password string) (*ElasticClient, error) {
	cfg := elasticsearch.Config{
		Addresses: addrs,
		Username:  username,
		Password:  password,
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	elasticClient = &ElasticClient{
		Client: client,
	}

	elasticClient.TypedClient, err = elasticsearch.NewTypedClient(cfg)
	if err != nil {
		return nil, err
	}
	return elasticClient, nil
}

//创建索引模版
func (this *ElasticClient) CreateIndexTemplate(indexName string, templateBody string) error {
	req := esapi.IndicesPutTemplateRequest{
		Name: indexName,
		Body: strings.NewReader(templateBody),
	}
	res, err := req.Do(context.Background(), this.Client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

//添加记录
func (this *ElasticClient) Add(indexName string, doc IDoc) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: doc.GetID(),
		Body:       strings.NewReader(string(data)),
		Refresh:    "true",
	}
	res, err := req.Do(context.Background(), this.Client)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

//高性能写入,批量添加多条记录bulk操作
func (this *ElasticClient) BulkAdd(indexName string, docList map[string]IDoc) error {
	start := time.Now()
	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:  indexName,
		Client: this.Client,
		OnFlushStart: func(ctx context.Context) context.Context {
			logger.Info("OnFlushStart", zap.String("indexName", indexName))
			return ctx
		},
		OnFlushEnd: func(ctx context.Context) {
			logger.Info("OnFlushEnd", zap.String("indexName", indexName))
		},
	})

	if err != nil {
		return err
	}
	var successfulCount int32
	var resErr error
	for id, doc := range docList {
		data, err := json.Marshal(doc)
		if err != nil {
			return err
		}

		err = indexer.Add(context.Background(),
			esutil.BulkIndexerItem{
				Action:     "index",
				DocumentID: id,
				Routing:    doc.GetRouting(),
				Body:       bytes.NewReader(data),
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddInt32(&successfulCount, 1)
				},
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						resErr = err
					}
					logger.Error("BulkAdd Error1", zap.Error(err))
				},
			},
		)
		if err != nil {
			logger.Error("BulkAdd Error2", zap.Error(err))
			return err
		}

		if resErr != nil {
			logger.Error("BulkAdd Error3", zap.Error(resErr))
			return resErr
		}
	}

	if err = indexer.Close(context.Background()); err != nil {
		return err
	}

	dur := time.Since(start)
	logger.Info("BulkAdd Result", zap.String("index", indexName), zap.Int64("Milliseconds Cost", dur.Milliseconds()), zap.Int32("successfulCount", successfulCount))
	return nil
}

//查询
func (this *ElasticClient) SearchList(indexName string, routing string, sourceIncludes string, searchReq *search.Request) (string, string, error) {
	es := this.TypedClient
	reqJsonDsl, err := json.Marshal(searchReq)
	logger.Info("MustQuery Request", zap.String("reqJsonDsl", string(reqJsonDsl)))
	search := es.Search().Index(indexName)
	if len(sourceIncludes) > 0 {
		search.SourceIncludes_(sourceIncludes)
	}
	//在路由明确的情况下使用路由可以提升性能
	if len(routing) > 0 {
		search.Routing(routing)
	}

	resp, err := search.Request(searchReq).Do(context.Background())
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	resp.Body.Close()
	return string(response), string(reqJsonDsl), err
}

func (this *ElasticClient) SearchCount(indexName string, routing string, searchReq *count.Request) (string, error) {
	es := this.TypedClient
	reqJsonDsl, err := json.Marshal(searchReq)
	logger.Info("SearchCount Request", zap.String("reqJsonDsl", string(reqJsonDsl)))

	search := es.Count().Index(indexName)
	//在路由明确的情况下使用路由可以提升性能
	if len(routing) > 0 {
		search.Routing(routing)
	}

	resp, err := search.Request(searchReq).Do(context.Background())
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	return string(response), err
}
