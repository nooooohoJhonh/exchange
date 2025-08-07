package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"exchange/internal/models/mongodb"
	"exchange/internal/pkg/database"
)

// MessageRepository MongoDB消息Repository实现
type MessageRepository struct {
	db *database.MongoDBService
}

// NewMessageRepository 创建消息Repository
func NewMessageRepository(db *database.MongoDBService) *MessageRepository {
	return &MessageRepository{db: db}
}

// SaveMessage 保存消息（实现接口方法）
func (r *MessageRepository) SaveMessage(ctx context.Context, message *mongodb.ChatMessage) error {
	return r.Create(ctx, message)
}

// GetMessageByID 根据ID获取消息（实现接口方法）
func (r *MessageRepository) GetMessageByID(ctx context.Context, messageID string) (*mongodb.ChatMessage, error) {
	return r.GetByID(ctx, messageID)
}

// DeleteMessage 删除消息（实现接口方法）
func (r *MessageRepository) DeleteMessage(ctx context.Context, messageID string) error {
	return r.Delete(ctx, messageID)
}

// CountDocuments 统计文档数量
func (r *MessageRepository) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	// 将interface{}转换为bson.M
	var bsonFilter bson.M
	if filter == nil {
		bsonFilter = bson.M{}
	} else if f, ok := filter.(bson.M); ok {
		bsonFilter = f
	} else {
		// 如果不是bson.M类型，使用空过滤器
		bsonFilter = bson.M{}
	}
	return r.db.CountDocuments(mongodb.ChatMessage{}.CollectionName(), bsonFilter)
}

// Create 创建消息
func (r *MessageRepository) Create(ctx context.Context, message *mongodb.ChatMessage) error {
	// 设置时间戳
	message.SetTimestamps()

	// 验证消息
	if err := message.Validate(); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	// 插入到MongoDB
	result, err := r.db.InsertOne(message.CollectionName(), message)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// 设置生成的ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		message.ID = oid
	}

	return nil
}

// GetByID 根据ID获取消息
func (r *MessageRepository) GetByID(ctx context.Context, messageID string) (*mongodb.ChatMessage, error) {
	oid, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %w", err)
	}

	filter := bson.M{"_id": oid}
	var message mongodb.ChatMessage

	err = r.db.FindOne(message.CollectionName(), filter, &message)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &message, nil
}

// Update 更新消息
func (r *MessageRepository) Update(ctx context.Context, message *mongodb.ChatMessage) error {
	if err := message.Validate(); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	message.UpdatedAt = time.Now()

	filter := bson.M{"_id": message.ID}
	update := bson.M{"$set": message}

	result, err := r.db.UpdateOne(message.CollectionName(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// Delete 删除消息
func (r *MessageRepository) Delete(ctx context.Context, messageID string) error {
	oid, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	filter := bson.M{"_id": oid}
	result, err := r.db.DeleteOne(mongodb.ChatMessage{}.CollectionName(), filter)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// List 获取消息列表
func (r *MessageRepository) List(ctx context.Context, limit, offset int) ([]*mongodb.ChatMessage, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	var messages []*mongodb.ChatMessage
	err := r.db.Find(mongodb.ChatMessage{}.CollectionName(), bson.M{}, &messages, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	return messages, nil
}

// GetConversationMessages 获取会话消息
func (r *MessageRepository) GetConversationMessages(ctx context.Context, userID1, userID2 string, limit, offset int) ([]*mongodb.ChatMessage, error) {
	// 构建查询条件：双向消息
	filter := bson.M{
		"$or": []bson.M{
			{"from_user_id": userID1, "to_user_id": userID2},
			{"from_user_id": userID2, "to_user_id": userID1},
		},
	}

	// 设置查询选项：按时间倒序，分页
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	var messages []*mongodb.ChatMessage
	err := r.db.Find(mongodb.ChatMessage{}.CollectionName(), filter, &messages, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	return messages, nil
}

// GetUserMessages 获取用户的所有消息
func (r *MessageRepository) GetUserMessages(ctx context.Context, userID string, limit, offset int) ([]*mongodb.ChatMessage, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"from_user_id": userID},
			{"to_user_id": userID},
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	var messages []*mongodb.ChatMessage
	err := r.db.Find(mongodb.ChatMessage{}.CollectionName(), filter, &messages, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get user messages: %w", err)
	}

	return messages, nil
}

// MarkAsRead 标记消息为已读
func (r *MessageRepository) MarkAsRead(ctx context.Context, messageID string) error {
	oid, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return fmt.Errorf("invalid message ID: %w", err)
	}

	filter := bson.M{"_id": oid}
	update := bson.M{
		"$set": bson.M{
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

// MarkConversationAsRead 标记会话中的所有未读消息为已读
func (r *MessageRepository) MarkConversationAsRead(ctx context.Context, fromUserID, toUserID string) (int64, error) {
	filter := bson.M{
		"from_user_id": fromUserID,
		"to_user_id":   toUserID,
		"is_read":      false,
	}

	update := bson.M{
		"$set": bson.M{
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

// GetUnreadCount 获取用户未读消息数量
func (r *MessageRepository) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	filter := bson.M{
		"to_user_id": userID,
		"is_read":    false,
	}

	count, err := r.db.CountDocuments(mongodb.ChatMessage{}.CollectionName(), filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread messages: %w", err)
	}

	return count, nil
}

// GetConversationUnreadCount 获取特定会话的未读消息数量
func (r *MessageRepository) GetConversationUnreadCount(ctx context.Context, fromUserID, toUserID string) (int64, error) {
	filter := bson.M{
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

// GetMessagesByTimeRange 根据时间范围获取消息
func (r *MessageRepository) GetMessagesByTimeRange(ctx context.Context, userID1, userID2 string, startTime, endTime time.Time) ([]*mongodb.ChatMessage, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"from_user_id": userID1, "to_user_id": userID2},
			{"from_user_id": userID2, "to_user_id": userID1},
		},
		"created_at": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})

	var messages []*mongodb.ChatMessage
	err := r.db.Find(mongodb.ChatMessage{}.CollectionName(), filter, &messages, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by time range: %w", err)
	}

	return messages, nil
}

// GetMessageStats 获取消息统计信息
func (r *MessageRepository) GetMessageStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"from_user_id": userID},
					{"to_user_id": userID},
				},
			},
		},
		{
			"$group": bson.M{
				"_id":            nil,
				"total_messages": bson.M{"$sum": 1},
				"sent_messages": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []string{"$from_user_id", userID}},
							1,
							0,
						},
					},
				},
				"received_messages": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{"$eq": []string{"$to_user_id", userID}},
							1,
							0,
						},
					},
				},
				"unread_messages": bson.M{
					"$sum": bson.M{
						"$cond": []interface{}{
							bson.M{
								"$and": []bson.M{
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

	var results []bson.M
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
	delete(stats, "_id") // 移除MongoDB的_id字段

	return stats, nil
}

// CreateIndexes 创建消息集合的索引
func (r *MessageRepository) CreateIndexes(ctx context.Context) error {
	collectionName := mongodb.ChatMessage{}.CollectionName()

	// 创建复合索引：from_user_id + to_user_id + created_at
	_, err := r.db.CreateIndex(collectionName, bson.D{
		{Key: "from_user_id", Value: 1},
		{Key: "to_user_id", Value: 1},
		{Key: "created_at", Value: -1},
	})
	if err != nil {
		return fmt.Errorf("failed to create conversation index: %w", err)
	}

	// 创建未读消息索引：to_user_id + is_read
	_, err = r.db.CreateIndex(collectionName, bson.D{
		{Key: "to_user_id", Value: 1},
		{Key: "is_read", Value: 1},
	})
	if err != nil {
		return fmt.Errorf("failed to create unread messages index: %w", err)
	}

	// 创建时间索引：created_at
	_, err = r.db.CreateIndex(collectionName, bson.D{
		{Key: "created_at", Value: -1},
	})
	if err != nil {
		return fmt.Errorf("failed to create time index: %w", err)
	}

	return nil
}
