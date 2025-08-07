package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"exchange/internal/pkg/config"
	appLogger "exchange/internal/pkg/logger"
)

// MongoDBService MongoDB文档数据库服务
type MongoDBService struct {
	client   *mongo.Client
	database *mongo.Database
	ctx      context.Context
}

// NewMongoDBService 创建MongoDB服务实例
func NewMongoDBService(cfg *config.Config) (*MongoDBService, error) {
	ctx := context.Background()

	// 设置连接选项
	clientOptions := options.Client().
		ApplyURI(cfg.MongoDB.URI).
		SetMaxPoolSize(20).
		SetMinPoolSize(5).
		SetMaxConnIdleTime(30 * time.Second).
		SetConnectTimeout(time.Duration(cfg.MongoDB.Timeout) * time.Second).
		SetSocketTimeout(time.Duration(cfg.MongoDB.Timeout) * time.Second).
		SetServerSelectionTimeout(time.Duration(cfg.MongoDB.Timeout) * time.Second)

	// 连接MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 测试连接
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// 获取数据库实例
	database := client.Database(cfg.MongoDB.Database)

	appLogger.Info("MongoDB connected successfully", map[string]interface{}{
		"uri":      cfg.MongoDB.URI,
		"database": cfg.MongoDB.Database,
	})

	return &MongoDBService{
		client:   client,
		database: database,
		ctx:      ctx,
	}, nil
}

// Client 获取MongoDB客户端
func (s *MongoDBService) Client() *mongo.Client {
	return s.client
}

// Database 获取数据库实例
func (s *MongoDBService) Database() *mongo.Database {
	return s.database
}

// Collection 获取集合
func (s *MongoDBService) Collection(name string) *mongo.Collection {
	return s.database.Collection(name)
}

// Close 关闭MongoDB连接
func (s *MongoDBService) Close() error {
	if s.client != nil {
		return s.client.Disconnect(s.ctx)
	}
	return nil
}

// HealthCheck MongoDB健康检查
func (s *MongoDBService) HealthCheck() error {
	if err := s.client.Ping(s.ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("MongoDB ping failed: %w", err)
	}
	return nil
}

// GetStats 获取MongoDB统计信息
func (s *MongoDBService) GetStats() (map[string]interface{}, error) {
	// 获取数据库统计信息
	var dbStats bson.M
	if err := s.database.RunCommand(s.ctx, bson.D{{Key: "dbStats", Value: 1}}).Decode(&dbStats); err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}

	// 获取服务器状态
	var serverStatus bson.M
	if err := s.database.RunCommand(s.ctx, bson.D{{Key: "serverStatus", Value: 1}}).Decode(&serverStatus); err != nil {
		return nil, fmt.Errorf("failed to get server status: %w", err)
	}

	return map[string]interface{}{
		"db_stats":      dbStats,
		"server_status": serverStatus,
	}, nil
}

// InsertOne 插入单个文档
func (s *MongoDBService) InsertOne(collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	collection := s.Collection(collectionName)
	result, err := collection.InsertOne(s.ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document into %s: %w", collectionName, err)
	}
	return result, nil
}

// InsertMany 插入多个文档
func (s *MongoDBService) InsertMany(collectionName string, documents []interface{}) (*mongo.InsertManyResult, error) {
	collection := s.Collection(collectionName)
	result, err := collection.InsertMany(s.ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("failed to insert documents into %s: %w", collectionName, err)
	}
	return result, nil
}

// FindOne 查找单个文档
func (s *MongoDBService) FindOne(collectionName string, filter bson.M, result interface{}) error {
	collection := s.Collection(collectionName)
	err := collection.FindOne(s.ctx, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("document not found in %s", collectionName)
		}
		return fmt.Errorf("failed to find document in %s: %w", collectionName, err)
	}
	return nil
}

// Find 查找多个文档
func (s *MongoDBService) Find(collectionName string, filter bson.M, results interface{}, opts ...*options.FindOptions) error {
	collection := s.Collection(collectionName)
	cursor, err := collection.Find(s.ctx, filter, opts...)
	if err != nil {
		return fmt.Errorf("failed to find documents in %s: %w", collectionName, err)
	}
	defer cursor.Close(s.ctx)

	if err := cursor.All(s.ctx, results); err != nil {
		return fmt.Errorf("failed to decode documents from %s: %w", collectionName, err)
	}
	return nil
}

// UpdateOne 更新单个文档
func (s *MongoDBService) UpdateOne(collectionName string, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	collection := s.Collection(collectionName)
	result, err := collection.UpdateOne(s.ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update document in %s: %w", collectionName, err)
	}
	return result, nil
}

// UpdateMany 更新多个文档
func (s *MongoDBService) UpdateMany(collectionName string, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	collection := s.Collection(collectionName)
	result, err := collection.UpdateMany(s.ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update documents in %s: %w", collectionName, err)
	}
	return result, nil
}

// DeleteOne 删除单个文档
func (s *MongoDBService) DeleteOne(collectionName string, filter bson.M) (*mongo.DeleteResult, error) {
	collection := s.Collection(collectionName)
	result, err := collection.DeleteOne(s.ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete document from %s: %w", collectionName, err)
	}
	return result, nil
}

// DeleteMany 删除多个文档
func (s *MongoDBService) DeleteMany(collectionName string, filter bson.M) (*mongo.DeleteResult, error) {
	collection := s.Collection(collectionName)
	result, err := collection.DeleteMany(s.ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete documents from %s: %w", collectionName, err)
	}
	return result, nil
}

// CountDocuments 统计文档数量
func (s *MongoDBService) CountDocuments(collectionName string, filter bson.M) (int64, error) {
	collection := s.Collection(collectionName)
	count, err := collection.CountDocuments(s.ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents in %s: %w", collectionName, err)
	}
	return count, nil
}

// Aggregate 聚合查询
func (s *MongoDBService) Aggregate(collectionName string, pipeline []bson.M, results interface{}) error {
	collection := s.Collection(collectionName)
	cursor, err := collection.Aggregate(s.ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to aggregate documents in %s: %w", collectionName, err)
	}
	defer cursor.Close(s.ctx)

	if err := cursor.All(s.ctx, results); err != nil {
		return fmt.Errorf("failed to decode aggregation results from %s: %w", collectionName, err)
	}
	return nil
}

// CreateIndex 创建索引
func (s *MongoDBService) CreateIndex(collectionName string, keys bson.D, opts ...*options.IndexOptions) (string, error) {
	collection := s.Collection(collectionName)
	indexModel := mongo.IndexModel{
		Keys: keys,
	}
	if len(opts) > 0 {
		indexModel.Options = opts[0]
	}

	indexName, err := collection.Indexes().CreateOne(s.ctx, indexModel)
	if err != nil {
		return "", fmt.Errorf("failed to create index on %s: %w", collectionName, err)
	}
	return indexName, nil
}

// DropIndex 删除索引
func (s *MongoDBService) DropIndex(collectionName string, indexName string) error {
	collection := s.Collection(collectionName)
	_, err := collection.Indexes().DropOne(s.ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to drop index %s on %s: %w", indexName, collectionName, err)
	}
	return nil
}

// ListIndexes 列出索引
func (s *MongoDBService) ListIndexes(collectionName string) ([]bson.M, error) {
	collection := s.Collection(collectionName)
	cursor, err := collection.Indexes().List(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes on %s: %w", collectionName, err)
	}
	defer cursor.Close(s.ctx)

	var indexes []bson.M
	if err := cursor.All(s.ctx, &indexes); err != nil {
		return nil, fmt.Errorf("failed to decode indexes from %s: %w", collectionName, err)
	}
	return indexes, nil
}

// Transaction 执行事务
func (s *MongoDBService) Transaction(fn func(sessCtx mongo.SessionContext) error) error {
	session, err := s.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(s.ctx)

	// 包装函数以匹配WithTransaction的签名
	wrappedFn := func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	}

	_, err = session.WithTransaction(s.ctx, wrappedFn)
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}
	return nil
}

// BulkWrite 批量写操作
func (s *MongoDBService) BulkWrite(collectionName string, operations []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	collection := s.Collection(collectionName)
	result, err := collection.BulkWrite(s.ctx, operations, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to perform bulk write on %s: %w", collectionName, err)
	}
	return result, nil
}