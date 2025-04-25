package httpserver

import "chat/internal/domain"

type GetChatsResponse struct {
	Chats []int `json:"chats"`
}

type NewChatRequest struct {
	UserID2 int `json:"user_id_2"`
}

type NewChatResponse struct {
	ChatID int `json:"chat_id"`
}

type SendMessageRequest struct {
	Text string `json:"text"`
}

type SendMessageResponse struct {
	MessageID int `json:"message_id"`
}

type GetMessagesResponse []domain.Message
