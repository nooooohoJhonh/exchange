package mongodb

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeFile  MessageType = "file"
	MessageTypeAudio MessageType = "audio"
	MessageTypeVideo MessageType = "video"
)

// ChatMessage 聊天消息模型
type ChatMessage struct {
	ID          primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	FromUserID  string                 `json:"from_user_id" bson:"from_user_id"`
	ToUserID    string                 `json:"to_user_id" bson:"to_user_id"`
	MessageType MessageType            `json:"message_type" bson:"message_type"`
	Content     string                 `json:"content" bson:"content"`
	Metadata    map[string]interface{} `json:"metadata" bson:"metadata"`
	IsRead      bool                   `json:"is_read" bson:"is_read"`
	CreatedAt   time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" bson:"updated_at"`
}

// CollectionName 返回集合名称
func (ChatMessage) CollectionName() string {
	return "chat_messages"
}

// Validate 验证消息数据
func (cm *ChatMessage) Validate() error {
	if cm.FromUserID == "" {
		return errors.New("from_user_id is required")
	}
	
	if cm.ToUserID == "" {
		return errors.New("to_user_id is required")
	}
	
	if cm.FromUserID == cm.ToUserID {
		return errors.New("cannot send message to yourself")
	}
	
	if cm.Content == "" {
		return errors.New("content is required")
	}
	
	// 验证消息类型
	validTypes := []MessageType{
		MessageTypeText,
		MessageTypeImage,
		MessageTypeFile,
		MessageTypeAudio,
		MessageTypeVideo,
	}
	
	isValidType := false
	for _, validType := range validTypes {
		if cm.MessageType == validType {
			isValidType = true
			break
		}
	}
	
	if !isValidType {
		return errors.New("invalid message type")
	}
	
	// 根据消息类型验证内容
	switch cm.MessageType {
	case MessageTypeText:
		if len(cm.Content) > 5000 {
			return errors.New("text message too long (max 5000 characters)")
		}
	case MessageTypeImage, MessageTypeFile, MessageTypeAudio, MessageTypeVideo:
		// 对于文件类型，content应该是文件URL或路径
		if len(cm.Content) > 500 {
			return errors.New("file path too long (max 500 characters)")
		}
	}
	
	return nil
}

// SetTimestamps 设置时间戳
func (cm *ChatMessage) SetTimestamps() {
	now := time.Now()
	if cm.CreatedAt.IsZero() {
		cm.CreatedAt = now
	}
	cm.UpdatedAt = now
}

// MarkAsRead 标记为已读
func (cm *ChatMessage) MarkAsRead() {
	cm.IsRead = true
	cm.UpdatedAt = time.Now()
}

// GetConversationID 获取会话ID（用于索引和查询）
func (cm *ChatMessage) GetConversationID() string {
	// 确保会话ID的一致性，较小的用户ID在前
	if cm.FromUserID < cm.ToUserID {
		return cm.FromUserID + "_" + cm.ToUserID
	}
	return cm.ToUserID + "_" + cm.FromUserID
}

// IsFileMessage 检查是否为文件消息
func (cm *ChatMessage) IsFileMessage() bool {
	return cm.MessageType == MessageTypeImage ||
		cm.MessageType == MessageTypeFile ||
		cm.MessageType == MessageTypeAudio ||
		cm.MessageType == MessageTypeVideo
}

// GetFileInfo 获取文件信息（从metadata中）
func (cm *ChatMessage) GetFileInfo() map[string]interface{} {
	if !cm.IsFileMessage() {
		return nil
	}
	
	fileInfo := make(map[string]interface{})
	if cm.Metadata != nil {
		if fileName, ok := cm.Metadata["file_name"]; ok {
			fileInfo["file_name"] = fileName
		}
		if fileSize, ok := cm.Metadata["file_size"]; ok {
			fileInfo["file_size"] = fileSize
		}
		if mimeType, ok := cm.Metadata["mime_type"]; ok {
			fileInfo["mime_type"] = mimeType
		}
	}
	
	return fileInfo
}

// SetFileInfo 设置文件信息
func (cm *ChatMessage) SetFileInfo(fileName string, fileSize int64, mimeType string) {
	if cm.Metadata == nil {
		cm.Metadata = make(map[string]interface{})
	}
	
	cm.Metadata["file_name"] = fileName
	cm.Metadata["file_size"] = fileSize
	cm.Metadata["mime_type"] = mimeType
}

// CreateTextMessage 创建文本消息
func CreateTextMessage(fromUserID, toUserID, content string) *ChatMessage {
	msg := &ChatMessage{
		FromUserID:  fromUserID,
		ToUserID:    toUserID,
		MessageType: MessageTypeText,
		Content:     content,
		IsRead:      false,
	}
	msg.SetTimestamps()
	return msg
}

// CreateFileMessage 创建文件消息
func CreateFileMessage(fromUserID, toUserID, filePath string, messageType MessageType, fileName string, fileSize int64, mimeType string) *ChatMessage {
	msg := &ChatMessage{
		FromUserID:  fromUserID,
		ToUserID:    toUserID,
		MessageType: messageType,
		Content:     filePath,
		IsRead:      false,
	}
	msg.SetFileInfo(fileName, fileSize, mimeType)
	msg.SetTimestamps()
	return msg
}