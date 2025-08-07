package database

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"exchange/internal/pkg/config"
)

func TestNewMongoDBService(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		MongoDB: config.MongoConfig{
			URI:      "mongodb://localhost:27017",
			Database: "test_db",
			Timeout:  10,
		},
	}

	// 注意：这个测试需要实际的MongoDB连接，在CI/CD环境中可能需要跳过
	t.Skip("Skipping MongoDB integration test - requires actual MongoDB connection")

	service, err := NewMongoDBService(cfg)
	if err != nil {
		t.Fatalf("Failed to create MongoDB service: %v", err)
	}
	defer service.Close()

	// 测试健康检查
	if err := service.HealthCheck(); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}

	// 测试获取统计信息
	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats == nil {
		t.Fatal("Stats should not be nil")
	}
}

func TestMongoDBService_CRUD(t *testing.T) {
	t.Skip("Skipping MongoDB CRUD test - requires actual MongoDB connection")

	cfg := &config.Config{
		MongoDB: config.MongoConfig{
			URI:      "mongodb://localhost:27017",
			Database: "test_db",
			Timeout:  10,
		},
	}

	service, err := NewMongoDBService(cfg)
	if err != nil {
		t.Fatalf("Failed to create MongoDB service: %v", err)
	}
	defer service.Close()

	collectionName := "test_collection"
	
	// 测试数据
	testDoc := bson.M{
		"name":       "test_document",
		"value":      123,
		"created_at": time.Now(),
	}

	// 测试插入
	result, err := service.InsertOne(collectionName, testDoc)
	if err != nil {
		t.Fatalf("Failed to insert document: %v", err)
	}

	insertedID := result.InsertedID.(primitive.ObjectID)

	// 测试查找
	var foundDoc bson.M
	filter := bson.M{"_id": insertedID}
	if err := service.FindOne(collectionName, filter, &foundDoc); err != nil {
		t.Fatalf("Failed to find document: %v", err)
	}

	if foundDoc["name"] != testDoc["name"] {
		t.Errorf("Expected name %v, got %v", testDoc["name"], foundDoc["name"])
	}

	// 测试更新
	update := bson.M{"$set": bson.M{"value": 456}}
	updateResult, err := service.UpdateOne(collectionName, filter, update)
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	if updateResult.ModifiedCount != 1 {
		t.Errorf("Expected 1 modified document, got %d", updateResult.ModifiedCount)
	}

	// 测试删除
	deleteResult, err := service.DeleteOne(collectionName, filter)
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}

	if deleteResult.DeletedCount != 1 {
		t.Errorf("Expected 1 deleted document, got %d", deleteResult.DeletedCount)
	}
}

func TestMongoDBService_Aggregation(t *testing.T) {
	t.Skip("Skipping MongoDB aggregation test - requires actual MongoDB connection")

	cfg := &config.Config{
		MongoDB: config.MongoConfig{
			URI:      "mongodb://localhost:27017",
			Database: "test_db",
			Timeout:  10,
		},
	}

	service, err := NewMongoDBService(cfg)
	if err != nil {
		t.Fatalf("Failed to create MongoDB service: %v", err)
	}
	defer service.Close()

	collectionName := "test_aggregation"

	// 插入测试数据
	testDocs := []interface{}{
		bson.M{"category": "A", "value": 10},
		bson.M{"category": "A", "value": 20},
		bson.M{"category": "B", "value": 15},
		bson.M{"category": "B", "value": 25},
	}

	_, err = service.InsertMany(collectionName, testDocs)
	if err != nil {
		t.Fatalf("Failed to insert test documents: %v", err)
	}

	// 聚合查询：按category分组并计算总和
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":   "$category",
			"total": bson.M{"$sum": "$value"},
		}},
		{"$sort": bson.M{"_id": 1}},
	}

	var results []bson.M
	if err := service.Aggregate(collectionName, pipeline, &results); err != nil {
		t.Fatalf("Failed to perform aggregation: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 aggregation results, got %d", len(results))
	}

	// 清理测试数据
	service.DeleteMany(collectionName, bson.M{})
}

func TestMongoDBService_Index(t *testing.T) {
	t.Skip("Skipping MongoDB index test - requires actual MongoDB connection")

	cfg := &config.Config{
		MongoDB: config.MongoConfig{
			URI:      "mongodb://localhost:27017",
			Database: "test_db",
			Timeout:  10,
		},
	}

	service, err := NewMongoDBService(cfg)
	if err != nil {
		t.Fatalf("Failed to create MongoDB service: %v", err)
	}
	defer service.Close()

	collectionName := "test_index"

	// 创建索引
	indexName, err := service.CreateIndex(collectionName, bson.D{{Key: "name", Value: 1}})
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	if indexName == "" {
		t.Error("Index name should not be empty")
	}

	// 列出索引
	indexes, err := service.ListIndexes(collectionName)
	if err != nil {
		t.Fatalf("Failed to list indexes: %v", err)
	}

	if len(indexes) < 2 { // 至少应该有_id索引和我们创建的索引
		t.Errorf("Expected at least 2 indexes, got %d", len(indexes))
	}

	// 删除索引
	if err := service.DropIndex(collectionName, indexName); err != nil {
		t.Fatalf("Failed to drop index: %v", err)
	}
}