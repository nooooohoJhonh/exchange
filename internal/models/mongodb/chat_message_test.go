package mongodb

import (
	"testing"
	"time"
)

func TestChatMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		message *ChatMessage
		wantErr bool
	}{
		{
			name: "valid text message",
			message: &ChatMessage{
				FromUserID:  "user1",
				ToUserID:    "user2",
				MessageType: MessageTypeText,
				Content:     "Hello, world!",
			},
			wantErr: false,
		},
		{
			name: "missing from_user_id",
			message: &ChatMessage{
				ToUserID:    "user2",
				MessageType: MessageTypeText,
				Content:     "Hello, world!",
			},
			wantErr: true,
		},
		{
			name: "missing to_user_id",
			message: &ChatMessage{
				FromUserID:  "user1",
				MessageType: MessageTypeText,
				Content:     "Hello, world!",
			},
			wantErr: true,
		},
		{
			name: "same from and to user",
			message: &ChatMessage{
				FromUserID:  "user1",
				ToUserID:    "user1",
				MessageType: MessageTypeText,
				Content:     "Hello, world!",
			},
			wantErr: true,
		},
		{
			name: "missing content",
			message: &ChatMessage{
				FromUserID:  "user1",
				ToUserID:    "user2",
				MessageType: MessageTypeText,
				Content:     "",
			},
			wantErr: true,
		},
		{
			name: "invalid message type",
			message: &ChatMessage{
				FromUserID:  "user1",
				ToUserID:    "user2",
				MessageType: "invalid",
				Content:     "Hello, world!",
			},
			wantErr: true,
		},
		{
			name: "text message too long",
			message: &ChatMessage{
				FromUserID:  "user1",
				ToUserID:    "user2",
				MessageType: MessageTypeText,
				Content:     string(make([]byte, 5001)), // 5001 characters
			},
			wantErr: true,
		},
		{
			name: "valid image message",
			message: &ChatMessage{
				FromUserID:  "user1",
				ToUserID:    "user2",
				MessageType: MessageTypeImage,
				Content:     "/uploads/image.jpg",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChatMessage_SetTimestamps(t *testing.T) {
	msg := &ChatMessage{
		FromUserID:  "user1",
		ToUserID:    "user2",
		MessageType: MessageTypeText,
		Content:     "Hello, world!",
	}

	// 初始时间戳应该为零值
	if !msg.CreatedAt.IsZero() {
		t.Error("CreatedAt should be zero initially")
	}
	if !msg.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be zero initially")
	}

	msg.SetTimestamps()

	// 设置时间戳后应该不为零值
	if msg.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero after SetTimestamps")
	}
	if msg.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero after SetTimestamps")
	}

	// 时间应该在最近1分钟内
	if time.Since(msg.CreatedAt) > time.Minute {
		t.Error("CreatedAt should be recent")
	}
	if time.Since(msg.UpdatedAt) > time.Minute {
		t.Error("UpdatedAt should be recent")
	}
}

func TestChatMessage_MarkAsRead(t *testing.T) {
	msg := &ChatMessage{
		FromUserID:  "user1",
		ToUserID:    "user2",
		MessageType: MessageTypeText,
		Content:     "Hello, world!",
		IsRead:      false,
	}

	if msg.IsRead {
		t.Error("Message should not be read initially")
	}

	msg.MarkAsRead()

	if !msg.IsRead {
		t.Error("Message should be marked as read")
	}

	if msg.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set when marking as read")
	}
}

func TestChatMessage_GetConversationID(t *testing.T) {
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
			msg := &ChatMessage{
				FromUserID: tt.fromUserID,
				ToUserID:   tt.toUserID,
			}

			conversationID := msg.GetConversationID()
			if conversationID != tt.expected {
				t.Errorf("GetConversationID() = %v, want %v", conversationID, tt.expected)
			}
		})
	}
}

func TestChatMessage_IsFileMessage(t *testing.T) {
	tests := []struct {
		name        string
		messageType MessageType
		expected    bool
	}{
		{"text message", MessageTypeText, false},
		{"image message", MessageTypeImage, true},
		{"file message", MessageTypeFile, true},
		{"audio message", MessageTypeAudio, true},
		{"video message", MessageTypeVideo, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &ChatMessage{MessageType: tt.messageType}
			if got := msg.IsFileMessage(); got != tt.expected {
				t.Errorf("IsFileMessage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestChatMessage_FileInfo(t *testing.T) {
	msg := &ChatMessage{
		MessageType: MessageTypeImage,
	}

	// 初始时应该没有文件信息
	fileInfo := msg.GetFileInfo()
	if len(fileInfo) != 0 {
		t.Error("File info should be empty initially")
	}

	// 设置文件信息
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"

	msg.SetFileInfo(fileName, fileSize, mimeType)

	// 获取文件信息
	fileInfo = msg.GetFileInfo()
	if fileInfo["file_name"] != fileName {
		t.Errorf("Expected file_name %s, got %v", fileName, fileInfo["file_name"])
	}
	if fileInfo["file_size"] != fileSize {
		t.Errorf("Expected file_size %d, got %v", fileSize, fileInfo["file_size"])
	}
	if fileInfo["mime_type"] != mimeType {
		t.Errorf("Expected mime_type %s, got %v", mimeType, fileInfo["mime_type"])
	}
}

func TestCreateTextMessage(t *testing.T) {
	fromUserID := "user1"
	toUserID := "user2"
	content := "Hello, world!"

	msg := CreateTextMessage(fromUserID, toUserID, content)

	if msg.FromUserID != fromUserID {
		t.Errorf("Expected FromUserID %s, got %s", fromUserID, msg.FromUserID)
	}
	if msg.ToUserID != toUserID {
		t.Errorf("Expected ToUserID %s, got %s", toUserID, msg.ToUserID)
	}
	if msg.MessageType != MessageTypeText {
		t.Errorf("Expected MessageType %s, got %s", MessageTypeText, msg.MessageType)
	}
	if msg.Content != content {
		t.Errorf("Expected Content %s, got %s", content, msg.Content)
	}
	if msg.IsRead {
		t.Error("Message should not be read initially")
	}
	if msg.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if msg.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestCreateFileMessage(t *testing.T) {
	fromUserID := "user1"
	toUserID := "user2"
	filePath := "/uploads/test.jpg"
	messageType := MessageTypeImage
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"

	msg := CreateFileMessage(fromUserID, toUserID, filePath, messageType, fileName, fileSize, mimeType)

	if msg.FromUserID != fromUserID {
		t.Errorf("Expected FromUserID %s, got %s", fromUserID, msg.FromUserID)
	}
	if msg.ToUserID != toUserID {
		t.Errorf("Expected ToUserID %s, got %s", toUserID, msg.ToUserID)
	}
	if msg.MessageType != messageType {
		t.Errorf("Expected MessageType %s, got %s", messageType, msg.MessageType)
	}
	if msg.Content != filePath {
		t.Errorf("Expected Content %s, got %s", filePath, msg.Content)
	}

	// 检查文件信息
	fileInfo := msg.GetFileInfo()
	if fileInfo["file_name"] != fileName {
		t.Errorf("Expected file_name %s, got %v", fileName, fileInfo["file_name"])
	}
	if fileInfo["file_size"] != fileSize {
		t.Errorf("Expected file_size %d, got %v", fileSize, fileInfo["file_size"])
	}
	if fileInfo["mime_type"] != mimeType {
		t.Errorf("Expected mime_type %s, got %v", mimeType, fileInfo["mime_type"])
	}
}