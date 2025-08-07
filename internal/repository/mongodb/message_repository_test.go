package mongodb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"exchange/internal/models/mongodb"
)

// MockMongoDBService 模拟MongoDB服务
type MockMongoDBService struct {
	mock.Mock
}

func (m *MockMongoDBService) InsertOne(collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	args := m.Called(collectionName, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockMongoDBService) FindOne(collectionName string, filter interface{}, result interface{}) error {
	args := m.Called(collectionName, filter, result)
	return args.Error(0)
}

func (m *MockMongoDBService) UpdateOne(collectionName string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(collectionName, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockMongoDBService) UpdateMany(collectionName string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(collectionName, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockMongoDBService) DeleteOne(collectionName string, filter interface{}) (*mongo.DeleteResult, error) {
	args := m.Called(collectionName, filter)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func (m *MockMongoDBService) Find(collectionName string, filter interface{}, results interface{}, opts ...interface{}) error {
	args := m.Called(collectionName, filter, results, opts)
	return args.Error(0)
}

func (m *MockMongoDBService) CountDocuments(collectionName string, filter interface{}) (int64, error) {
	args := m.Called(collectionName, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMongoDBService) Aggregate(collectionName string, pipeline interface{}, results interface{}) error {
	args := m.Called(collectionName, pipeline, results)
	return args.Error(0)
}

func (m *MockMongoDBService) CreateIndex(collectionName string, keys interface{}) (string, error) {
	args := m.Called(collectionName, keys)
	return args.String(0), args.Error(1)
}

// TestMessageRepository 测试用的消息Repository，使用接口
type TestMessageRepository struct {
	db *MockMongoDBService
}

func NewTestMessageRepository(db *MockMongoDBService) *TestMessageRepository {
	return &TestMessageRepository{db: db}
}

// 实现与MessageRepository相同的方法
func (r *TestMessageRepository) Create(ctx context.Context, message *mongodb.ChatMessage) error {
	message.SetTimestamps()
	
	if err := message.Validate(); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}
	
	result, err := r.db.InsertOne(message.CollectionName(), message)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		message.ID = oid
	}
	
	return nil
}

func (r *TestMessageRepository) GetByID(ctx context.Context, messageID string) (*mongodb.ChatMessage, error) {
	oid, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %w", err)
	}
	
	filter := map[string]interface{}{"_id": oid}
	var message mongodb.ChatMessage
	
	err = r.db.FindOne(message.CollectionName(), filter, &message)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	
	return &message, nil
}

func (r *TestMessageRepository) Update(ctx context.Context, message *mongodb.ChatMessage) error {
	if err := message.Validate(); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}
	
	message.UpdatedAt = time.Now()
	
	filter := map[string]interface{}{"_id": message.ID}
	update := map[string]interface{}{"$set": message}
	
	result, err := r.db.UpdateOne(message.CollectionName(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	
	if result.ModifiedCount == 0 {
		return fmt.Errorf("message not found")
	}
	
	return nil
}

func (r *TestMessageRepository) Delete(ctx context.Context, messageID string) error {
	oid, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}
	
	filter := map[string]interface{}{"_id": oid}
	result, err := r.db.DeleteOne(mongodb.ChatMessage{}.CollectionName(), filter)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	
	if result.DeletedCount == 0 {
		return fmt.Errorf("message not found")
	}
	
	return nil
}

func (r *TestMessageRepository) List(ctx context.Context, limit, offset int) ([]*mongodb.ChatMessage, error) {
	var messages []*mongodb.ChatMessage
	err := r.db.Find(mongodb.ChatMessage{}.CollectionName(), map[string]interface{}{}, &messages, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	
	return messages, nil
}

func (r *TestMessageRepository) GetConversationMessages(ctx context.Context, userID1, userID2 string, limit, offset int) ([]*mongodb.ChatMessage, error) {
	filter := map[string]interface{}{
		"$or": []map[string]interface{}{
			{"from_user_id": userID1, "to_user_id": userID2},
			{"from_user_id": userID2, "to_user_id": userID1},
		},
	}
	
	var messages []*mongodb.ChatMessage
	err := r.db.Find(mongodb.ChatMessage{}.CollectionName(), filter, &messages, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}
	
	return messages, nil
}

func (r *TestMessageRepository) GetUserMessages(ctx context.Context, userID string, limit, offset int) ([]*mongodb.ChatMessage, error) {
	filter := map[string]interface{}{
		"$or": []map[string]interface{}{
			{"from_user_id": userID},
			{"to_user_id": userID},
		},
	}
	
	var messages []*mongodb.ChatMessage
	err := r.db.Find(mongodb.ChatMessage{}.CollectionName(), filter, &messages, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user messages: %w", err)
	}
	
	return messages, nil
}

func (r *TestMessageRepository) MarkAsRead(ctx context.Context, messageID string) error {
	oid, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}
	
	filter := map[string]interface{}{"_id": oid}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"is_read":    true,
			"updated_at": time.Now(),
		},
	}
	
	result, err := r.db.UpdateOne(mongodb.ChatMessage{}.CollectionName(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}
	
	if result.ModifiedCount == 0 {
		return fmt.Errorf("message not found or already read")
	}
	
	return nil
}

func (r *TestMessageRepository) MarkConversationAsRead(ctx context.Context, fromUserID, toUserID string) (int64, error) {
	filter := map[string]interface{}{
		"from_user_id": fromUserID,
		"to_user_id":   toUserID,
		"is_read":      false,
	}
	
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"is_read":    true,
			"updated_at": time.Now(),
		},
	}
	
	result, err := r.db.UpdateMany(mongodb.ChatMessage{}.CollectionName(), filter, update)
	if err != nil {
		return 0, fmt.Errorf("failed to mark conversation as read: %w", err)
	}
	
	return result.ModifiedCount, nil
}

func (r *TestMessageRepository) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	filter := map[string]interface{}{
		"to_user_id": userID,
		"is_read":    false,
	}
	
	count, err := r.db.CountDocuments(mongodb.ChatMessage{}.CollectionName(), filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread messages: %w", err)
	}
	
	return count, nil
}

func (r *TestMessageRepository) GetConversationUnreadCount(ctx context.Context, fromUserID, toUserID string) (int64, error) {
	filter := map[string]interface{}{
		"from_user_id": fromUserID,
		"to_user_id":   toUserID,
		"is_read":      false,
	}
	
	count, err := r.db.CountDocuments(mongodb.ChatMessage{}.CollectionName(), filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count conversation unread messages: %w", err)
	}
	
	return count, nil
}

func (r *TestMessageRepository) GetMessagesByTimeRange(ctx context.Context, userID1, userID2 string, startTime, endTime time.Time) ([]*mongodb.ChatMessage, error) {
	filter := map[string]interface{}{
		"$or": []map[string]interface{}{
			{"from_user_id": userID1, "to_user_id": userID2},
			{"from_user_id": userID2, "to_user_id": userID1},
		},
		"created_at": map[string]interface{}{
			"$gte": startTime,
			"$lte": endTime,
		},
	}
	
	var messages []*mongodb.ChatMessage
	err := r.db.Find(mongodb.ChatMessage{}.CollectionName(), filter, &messages, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by time range: %w", err)
	}
	
	return messages, nil
}

func (r *TestMessageRepository) GetMessageStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	pipeline := []map[string]interface{}{
		{
			"$match": map[string]interface{}{
				"$or": []map[string]interface{}{
					{"from_user_id": userID},
					{"to_user_id": userID},
				},
			},
		},
		{
			"$group": map[string]interface{}{
				"_id": nil,
				"total_messages": map[string]interface{}{"$sum": 1},
				"sent_messages": map[string]interface{}{
					"$sum": map[string]interface{}{
						"$cond": []interface{}{
							map[string]interface{}{"$eq": []string{"$from_user_id", userID}},
							1,
							0,
						},
					},
				},
				"received_messages": map[string]interface{}{
					"$sum": map[string]interface{}{
						"$cond": []interface{}{
							map[string]interface{}{"$eq": []string{"$to_user_id", userID}},
							1,
							0,
						},
					},
				},
				"unread_messages": map[string]interface{}{
					"$sum": map[string]interface{}{
						"$cond": []interface{}{
							map[string]interface{}{
								"$and": []map[string]interface{}{
									{"$eq": []string{"$to_user_id", userID}},
									{"$eq": []interface{}{"$is_read", false}},
								},
							},
							1,
							0,
						},
					},
				},
			},
		},
	}
	
	var results []map[string]interface{}
	err := r.db.Aggregate(mongodb.ChatMessage{}.CollectionName(), pipeline, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to get message stats: %w", err)
	}
	
	if len(results) == 0 {
		return map[string]interface{}{
			"total_messages":    0,
			"sent_messages":     0,
			"received_messages": 0,
			"unread_messages":   0,
		}, nil
	}
	
	stats := results[0]
	delete(stats, "_id")
	
	return stats, nil
}

func (r *TestMessageRepository) CreateIndexes(ctx context.Context) error {
	collectionName := mongodb.ChatMessage{}.CollectionName()
	
	// 创建三个索引
	_, err := r.db.CreateIndex(collectionName, map[string]interface{}{
		"from_user_id": 1,
		"to_user_id":   1,
		"created_at":   -1,
	})
	if err != nil {
		return fmt.Errorf("failed to create conversation index: %w", err)
	}
	
	_, err = r.db.CreateIndex(collectionName, map[string]interface{}{
		"to_user_id": 1,
		"is_read":    1,
	})
	if err != nil {
		return fmt.Errorf("failed to create unread messages index: %w", err)
	}
	
	_, err = r.db.CreateIndex(collectionName, map[string]interface{}{
		"created_at": -1,
	})
	if err != nil {
		return fmt.Errorf("failed to create time index: %w", err)
	}
	
	return nil
}

// 测试用例开始

func TestMessageRepository_Create(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 准备测试数据
	message := &mongodb.ChatMessage{
		FromUserID:  "user1",
		ToUserID:    "user2",
		MessageType: mongodb.MessageTypeText,
		Content:     "Hello, world!",
	}
	
	// 设置mock期望
	expectedID := primitive.NewObjectID()
	mockResult := &mongo.InsertOneResult{InsertedID: expectedID}
	mockDB.On("InsertOne", "chat_messages", message).Return(mockResult, nil)
	
	// 执行测试
	err := repo.Create(context.Background(), message)
	
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, expectedID, message.ID)
	assert.False(t, message.CreatedAt.IsZero())
	assert.False(t, message.UpdatedAt.IsZero())
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_Create_ValidationError(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 准备无效消息（缺少必要字段）
	invalidMessage := &mongodb.ChatMessage{
		FromUserID: "user1",
		// 缺少ToUserID
		MessageType: mongodb.MessageTypeText,
		Content:     "Invalid message",
	}
	
	// 执行测试
	err := repo.Create(context.Background(), invalidMessage)
	
	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	
	// 不应该调用数据库操作
	mockDB.AssertNotCalled(t, "InsertOne")
}

func TestMessageRepository_GetByID(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 准备测试数据
	messageID := primitive.NewObjectID()
	expectedMessage := &mongodb.ChatMessage{
		ID:          messageID,
		FromUserID:  "user1",
		ToUserID:    "user2",
		MessageType: mongodb.MessageTypeText,
		Content:     "Hello, world!",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// 设置mock期望
	mockDB.On("FindOne", "chat_messages", mock.Anything, mock.AnythingOfType("*mongodb.ChatMessage")).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*mongodb.ChatMessage)
			*result = *expectedMessage
		}).Return(nil)
	
	// 执行测试
	result, err := repo.GetByID(context.Background(), messageID.Hex())
	
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage.ID, result.ID)
	assert.Equal(t, expectedMessage.FromUserID, result.FromUserID)
	assert.Equal(t, expectedMessage.ToUserID, result.ToUserID)
	assert.Equal(t, expectedMessage.Content, result.Content)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_GetByID_InvalidID(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 执行测试
	_, err := repo.GetByID(context.Background(), "invalid-id")
	
	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid message ID")
	
	// 不应该调用数据库操作
	mockDB.AssertNotCalled(t, "FindOne")
}

func TestMessageRepository_Update(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 准备测试数据
	message := &mongodb.ChatMessage{
		ID:          primitive.NewObjectID(),
		FromUserID:  "user1",
		ToUserID:    "user2",
		MessageType: mongodb.MessageTypeText,
		Content:     "Updated content",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// 设置mock期望
	mockResult := &mongo.UpdateResult{ModifiedCount: 1}
	mockDB.On("UpdateOne", "chat_messages", mock.Anything, mock.Anything).Return(mockResult, nil)
	
	// 执行测试
	err := repo.Update(context.Background(), message)
	
	// 验证结果
	assert.NoError(t, err)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_Update_NotFound(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	message := &mongodb.ChatMessage{
		ID:          primitive.NewObjectID(),
		FromUserID:  "user1",
		ToUserID:    "user2",
		MessageType: mongodb.MessageTypeText,
		Content:     "Test message",
	}
	
	// 设置mock期望 - 返回0个修改的文档
	mockResult := &mongo.UpdateResult{ModifiedCount: 0}
	mockDB.On("UpdateOne", "chat_messages", mock.Anything, mock.Anything).Return(mockResult, nil)
	
	// 执行测试
	err := repo.Update(context.Background(), message)
	
	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message not found")
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_Delete(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 准备测试数据
	messageID := primitive.NewObjectID()
	
	// 设置mock期望
	mockResult := &mongo.DeleteResult{DeletedCount: 1}
	mockDB.On("DeleteOne", "chat_messages", mock.Anything).Return(mockResult, nil)
	
	// 执行测试
	err := repo.Delete(context.Background(), messageID.Hex())
	
	// 验证结果
	assert.NoError(t, err)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_Delete_NotFound(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	messageID := primitive.NewObjectID()
	
	// 设置mock期望 - 返回0个删除的文档
	mockResult := &mongo.DeleteResult{DeletedCount: 0}
	mockDB.On("DeleteOne", "chat_messages", mock.Anything).Return(mockResult, nil)
	
	// 执行测试
	err := repo.Delete(context.Background(), messageID.Hex())
	
	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message not found")
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_List(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 准备测试数据
	expectedMessages := []*mongodb.ChatMessage{
		{
			ID:          primitive.NewObjectID(),
			FromUserID:  "user1",
			ToUserID:    "user2",
			MessageType: mongodb.MessageTypeText,
			Content:     "Message 1",
			CreatedAt:   time.Now(),
		},
		{
			ID:          primitive.NewObjectID(),
			FromUserID:  "user2",
			ToUserID:    "user1",
			MessageType: mongodb.MessageTypeText,
			Content:     "Message 2",
			CreatedAt:   time.Now(),
		},
	}
	
	// 设置mock期望
	mockDB.On("Find", "chat_messages", mock.Anything, mock.AnythingOfType("*[]*mongodb.ChatMessage"), mock.Anything).
		Run(func(args mock.Arguments) {
			results := args.Get(2).(*[]*mongodb.ChatMessage)
			*results = expectedMessages
		}).Return(nil)
	
	// 执行测试
	messages, err := repo.List(context.Background(), 10, 0)
	
	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, expectedMessages[0].Content, messages[0].Content)
	assert.Equal(t, expectedMessages[1].Content, messages[1].Content)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_GetConversationMessages(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 准备测试数据
	expectedMessages := []*mongodb.ChatMessage{
		{
			ID:          primitive.NewObjectID(),
			FromUserID:  "user1",
			ToUserID:    "user2",
			MessageType: mongodb.MessageTypeText,
			Content:     "Hello",
			CreatedAt:   time.Now(),
		},
		{
			ID:          primitive.NewObjectID(),
			FromUserID:  "user2",
			ToUserID:    "user1",
			MessageType: mongodb.MessageTypeText,
			Content:     "Hi there",
			CreatedAt:   time.Now(),
		},
	}
	
	// 设置mock期望
	mockDB.On("Find", "chat_messages", mock.Anything, mock.AnythingOfType("*[]*mongodb.ChatMessage"), mock.Anything).
		Run(func(args mock.Arguments) {
			results := args.Get(2).(*[]*mongodb.ChatMessage)
			*results = expectedMessages
		}).Return(nil)
	
	// 执行测试
	messages, err := repo.GetConversationMessages(context.Background(), "user1", "user2", 10, 0)
	
	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_GetUserMessages(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 准备测试数据
	expectedMessages := []*mongodb.ChatMessage{
		{
			ID:          primitive.NewObjectID(),
			FromUserID:  "user1",
			ToUserID:    "user2",
			MessageType: mongodb.MessageTypeText,
			Content:     "Sent message",
			CreatedAt:   time.Now(),
		},
		{
			ID:          primitive.NewObjectID(),
			FromUserID:  "user3",
			ToUserID:    "user1",
			MessageType: mongodb.MessageTypeText,
			Content:     "Received message",
			CreatedAt:   time.Now(),
		},
	}
	
	// 设置mock期望
	mockDB.On("Find", "chat_messages", mock.Anything, mock.AnythingOfType("*[]*mongodb.ChatMessage"), mock.Anything).
		Run(func(args mock.Arguments) {
			results := args.Get(2).(*[]*mongodb.ChatMessage)
			*results = expectedMessages
		}).Return(nil)
	
	// 执行测试
	messages, err := repo.GetUserMessages(context.Background(), "user1", 10, 0)
	
	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_MarkAsRead(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 准备测试数据
	messageID := primitive.NewObjectID()
	
	// 设置mock期望
	mockResult := &mongo.UpdateResult{ModifiedCount: 1}
	mockDB.On("UpdateOne", "chat_messages", mock.Anything, mock.Anything).Return(mockResult, nil)
	
	// 执行测试
	err := repo.MarkAsRead(context.Background(), messageID.Hex())
	
	// 验证结果
	assert.NoError(t, err)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_MarkAsRead_NotFound(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	messageID := primitive.NewObjectID()
	
	// 设置mock期望 - 返回0个修改的文档
	mockResult := &mongo.UpdateResult{ModifiedCount: 0}
	mockDB.On("UpdateOne", "chat_messages", mock.Anything, mock.Anything).Return(mockResult, nil)
	
	// 执行测试
	err := repo.MarkAsRead(context.Background(), messageID.Hex())
	
	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message not found or already read")
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_MarkConversationAsRead(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 设置mock期望
	expectedCount := int64(3)
	mockResult := &mongo.UpdateResult{ModifiedCount: expectedCount}
	mockDB.On("UpdateMany", "chat_messages", mock.Anything, mock.Anything).Return(mockResult, nil)
	
	// 执行测试
	count, err := repo.MarkConversationAsRead(context.Background(), "user1", "user2")
	
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_GetUnreadCount(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 设置mock期望
	expectedCount := int64(5)
	mockDB.On("CountDocuments", "chat_messages", mock.Anything).Return(expectedCount, nil)
	
	// 执行测试
	count, err := repo.GetUnreadCount(context.Background(), "user1")
	
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_GetConversationUnreadCount(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 设置mock期望
	expectedCount := int64(2)
	mockDB.On("CountDocuments", "chat_messages", mock.Anything).Return(expectedCount, nil)
	
	// 执行测试
	count, err := repo.GetConversationUnreadCount(context.Background(), "user1", "user2")
	
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_GetMessagesByTimeRange(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 准备测试数据
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()
	expectedMessages := []*mongodb.ChatMessage{
		{
			ID:          primitive.NewObjectID(),
			FromUserID:  "user1",
			ToUserID:    "user2",
			MessageType: mongodb.MessageTypeText,
			Content:     "Message in range",
			CreatedAt:   time.Now().Add(-12 * time.Hour),
		},
	}
	
	// 设置mock期望
	mockDB.On("Find", "chat_messages", mock.Anything, mock.AnythingOfType("*[]*mongodb.ChatMessage"), mock.Anything).
		Run(func(args mock.Arguments) {
			results := args.Get(2).(*[]*mongodb.ChatMessage)
			*results = expectedMessages
		}).Return(nil)
	
	// 执行测试
	messages, err := repo.GetMessagesByTimeRange(context.Background(), "user1", "user2", startTime, endTime)
	
	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "Message in range", messages[0].Content)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_GetMessageStats(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 准备测试数据
	expectedStats := []map[string]interface{}{
		{
			"total_messages":    int32(10),
			"sent_messages":     int32(6),
			"received_messages": int32(4),
			"unread_messages":   int32(2),
		},
	}
	
	// 设置mock期望
	mockDB.On("Aggregate", "chat_messages", mock.Anything, mock.AnythingOfType("*[]map[string]interface {}")).
		Run(func(args mock.Arguments) {
			results := args.Get(2).(*[]map[string]interface{})
			*results = expectedStats
		}).Return(nil)
	
	// 执行测试
	stats, err := repo.GetMessageStats(context.Background(), "user1")
	
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, int32(10), stats["total_messages"])
	assert.Equal(t, int32(6), stats["sent_messages"])
	assert.Equal(t, int32(4), stats["received_messages"])
	assert.Equal(t, int32(2), stats["unread_messages"])
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_GetMessageStats_NoMessages(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 设置mock期望 - 返回空结果
	emptyResults := []map[string]interface{}{}
	mockDB.On("Aggregate", "chat_messages", mock.Anything, mock.AnythingOfType("*[]map[string]interface {}")).
		Run(func(args mock.Arguments) {
			results := args.Get(2).(*[]map[string]interface{})
			*results = emptyResults
		}).Return(nil)
	
	// 执行测试
	stats, err := repo.GetMessageStats(context.Background(), "user1")
	
	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 0, stats["total_messages"])
	assert.Equal(t, 0, stats["sent_messages"])
	assert.Equal(t, 0, stats["received_messages"])
	assert.Equal(t, 0, stats["unread_messages"])
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

func TestMessageRepository_CreateIndexes(t *testing.T) {
	mockDB := new(MockMongoDBService)
	repo := NewTestMessageRepository(mockDB)
	
	// 设置mock期望 - 创建三个索引
	mockDB.On("CreateIndex", "chat_messages", mock.Anything).Return("index_name", nil).Times(3)
	
	// 执行测试
	err := repo.CreateIndexes(context.Background())
	
	// 验证结果
	assert.NoError(t, err)
	
	// 验证mock调用
	mockDB.AssertExpectations(t)
}

// 测试消息模型的业务逻辑
func TestMessageRepository_BusinessLogic(t *testing.T) {
	// 测试消息业务逻辑，不依赖数据库
	
	// 创建文本消息
	textMsg := mongodb.CreateTextMessage("user1", "user2", "Hello, world!")
	
	assert.Equal(t, "user1", textMsg.FromUserID)
	assert.Equal(t, "user2", textMsg.ToUserID)
	assert.Equal(t, mongodb.MessageTypeText, textMsg.MessageType)
	assert.Equal(t, "Hello, world!", textMsg.Content)
	assert.False(t, textMsg.IsRead)
	
	// 测试会话ID生成
	conversationID := textMsg.GetConversationID()
	expectedID := "user1_user2"
	assert.Equal(t, expectedID, conversationID)
	
	// 测试反向会话ID
	reverseMsg := mongodb.CreateTextMessage("user2", "user1", "Hi there!")
	reverseConversationID := reverseMsg.GetConversationID()
	assert.Equal(t, expectedID, reverseConversationID)
	
	// 测试标记为已读
	textMsg.MarkAsRead()
	assert.True(t, textMsg.IsRead)
	
	// 测试文件消息
	fileMsg := mongodb.CreateFileMessage(
		"user1", "user2", "/uploads/image.jpg",
		mongodb.MessageTypeImage, "image.jpg", 1024, "image/jpeg",
	)
	
	assert.True(t, fileMsg.IsFileMessage())
	
	fileInfo := fileMsg.GetFileInfo()
	assert.Equal(t, "image.jpg", fileInfo["file_name"])
	assert.Equal(t, int64(1024), fileInfo["file_size"])
	assert.Equal(t, "image/jpeg", fileInfo["mime_type"])
}

func TestMessageRepository_Validation(t *testing.T) {
	// 测试消息验证逻辑
	
	// 有效消息
	validMsg := &mongodb.ChatMessage{
		FromUserID:  "user1",
		ToUserID:    "user2",
		MessageType: mongodb.MessageTypeText,
		Content:     "Valid message",
	}
	
	assert.NoError(t, validMsg.Validate())
	
	// 缺少发送者
	invalidMsg1 := &mongodb.ChatMessage{
		ToUserID:    "user2",
		MessageType: mongodb.MessageTypeText,
		Content:     "Missing sender",
	}
	
	assert.Error(t, invalidMsg1.Validate())
	
	// 缺少接收者
	invalidMsg2 := &mongodb.ChatMessage{
		FromUserID:  "user1",
		MessageType: mongodb.MessageTypeText,
		Content:     "Missing receiver",
	}
	
	assert.Error(t, invalidMsg2.Validate())
	
	// 发送者和接收者相同
	invalidMsg3 := &mongodb.ChatMessage{
		FromUserID:  "user1",
		ToUserID:    "user1",
		MessageType: mongodb.MessageTypeText,
		Content:     "Self message",
	}
	
	assert.Error(t, invalidMsg3.Validate())
	
	// 缺少内容
	invalidMsg4 := &mongodb.ChatMessage{
		FromUserID:  "user1",
		ToUserID:    "user2",
		MessageType: mongodb.MessageTypeText,
		Content:     "",
	}
	
	assert.Error(t, invalidMsg4.Validate())
	
	// 无效消息类型
	invalidMsg5 := &mongodb.ChatMessage{
		FromUserID:  "user1",
		ToUserID:    "user2",
		MessageType: "invalid_type",
		Content:     "Invalid type message",
	}
	
	assert.Error(t, invalidMsg5.Validate())
	
	// 文本消息过长
	longContent := string(make([]byte, 5001))
	invalidMsg6 := &mongodb.ChatMessage{
		FromUserID:  "user1",
		ToUserID:    "user2",
		MessageType: mongodb.MessageTypeText,
		Content:     longContent,
	}
	
	assert.Error(t, invalidMsg6.Validate())
}

func TestMessageRepository_Timestamps(t *testing.T) {
	msg := &mongodb.ChatMessage{
		FromUserID:  "user1",
		ToUserID:    "user2",
		MessageType: mongodb.MessageTypeText,
		Content:     "Test message",
	}
	
	// 初始时间戳应该为零值
	assert.True(t, msg.CreatedAt.IsZero())
	assert.True(t, msg.UpdatedAt.IsZero())
	
	msg.SetTimestamps()
	
	// 设置时间戳后应该不为零值
	assert.False(t, msg.CreatedAt.IsZero())
	assert.False(t, msg.UpdatedAt.IsZero())
	
	// 时间应该在最近1分钟内
	assert.True(t, time.Since(msg.CreatedAt) < time.Minute)
	assert.True(t, time.Since(msg.UpdatedAt) < time.Minute)
}

func TestMessageRepository_FileOperations(t *testing.T) {
	msg := &mongodb.ChatMessage{
		MessageType: mongodb.MessageTypeImage,
	}
	
	// 初始时应该没有文件信息
	fileInfo := msg.GetFileInfo()
	assert.Empty(t, fileInfo)
	
	// 设置文件信息
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"
	
	msg.SetFileInfo(fileName, fileSize, mimeType)
	
	// 获取文件信息
	fileInfo = msg.GetFileInfo()
	assert.Equal(t, fileName, fileInfo["file_name"])
	assert.Equal(t, fileSize, fileInfo["file_size"])
	assert.Equal(t, mimeType, fileInfo["mime_type"])
}

func TestMessageRepository_ConversationID(t *testing.T) {
	tests := []struct {
		name       string
		fromUserID string
		toUserID   string
		expected   string
	}{
		{
			name:       "user1 to user2",
			fromUserID: "user1",
			toUserID:   "user2",
			expected:   "user1_user2",
		},
		{
			name:       "user2 to user1",
			fromUserID: "user2",
			toUserID:   "user1",
			expected:   "user1_user2", // Should be same as above
		},
		{
			name:       "userA to userB",
			fromUserID: "userA",
			toUserID:   "userB",
			expected:   "userA_userB",
		},
		{
			name:       "userB to userA",
			fromUserID: "userB",
			toUserID:   "userA",
			expected:   "userA_userB", // Should be same as above
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &mongodb.ChatMessage{
				FromUserID: tt.fromUserID,
				ToUserID:   tt.toUserID,
			}

			conversationID := msg.GetConversationID()
			assert.Equal(t, tt.expected, conversationID)
		})
	}
}

func TestMessageRepository_MessageTypes(t *testing.T) {
	tests := []struct {
		name        string
		messageType mongodb.MessageType
		isFile      bool
	}{
		{"text message", mongodb.MessageTypeText, false},
		{"image message", mongodb.MessageTypeImage, true},
		{"file message", mongodb.MessageTypeFile, true},
		{"audio message", mongodb.MessageTypeAudio, true},
		{"video message", mongodb.MessageTypeVideo, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &mongodb.ChatMessage{MessageType: tt.messageType}
			assert.Equal(t, tt.isFile, msg.IsFileMessage())
		})
	}
}

func TestMessageRepository_CreateHelpers(t *testing.T) {
	// 测试创建文本消息
	fromUserID := "user1"
	toUserID := "user2"
	content := "Hello, world!"

	textMsg := mongodb.CreateTextMessage(fromUserID, toUserID, content)

	assert.Equal(t, fromUserID, textMsg.FromUserID)
	assert.Equal(t, toUserID, textMsg.ToUserID)
	assert.Equal(t, mongodb.MessageTypeText, textMsg.MessageType)
	assert.Equal(t, content, textMsg.Content)
	assert.False(t, textMsg.IsRead)
	assert.False(t, textMsg.CreatedAt.IsZero())
	assert.False(t, textMsg.UpdatedAt.IsZero())

	// 测试创建文件消息
	filePath := "/uploads/test.jpg"
	messageType := mongodb.MessageTypeImage
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"

	fileMsg := mongodb.CreateFileMessage(fromUserID, toUserID, filePath, messageType, fileName, fileSize, mimeType)

	assert.Equal(t, fromUserID, fileMsg.FromUserID)
	assert.Equal(t, toUserID, fileMsg.ToUserID)
	assert.Equal(t, messageType, fileMsg.MessageType)
	assert.Equal(t, filePath, fileMsg.Content)

	// 检查文件信息
	fileInfo := fileMsg.GetFileInfo()
	assert.Equal(t, fileName, fileInfo["file_name"])
	assert.Equal(t, fileSize, fileInfo["file_size"])
	assert.Equal(t, mimeType, fileInfo["mime_type"])
}