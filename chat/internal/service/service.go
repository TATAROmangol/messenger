package service

import (
	"chat/internal/domain"
	"context"
	"errors"
)

//go:generate mockgen -destination=./mock/mock.go -package=mock -source=service.go

type ChatRepo interface{
	CreateChat(ctx context.Context, userID1, userID2 int) (int, error)
	SendMessage(ctx context.Context, userID, chatID int, text string) (int, error)
	GetMessages(ctx context.Context, chatID int, limit, offset int) ([]domain.Message, error)
	GetUserChats(ctx context.Context, userID int) ([]int, error)
}

type ChatSvc struct {
	ChatRepo ChatRepo
}

func New(chatRepo ChatRepo) *ChatSvc {
	return &ChatSvc{ChatRepo: chatRepo}
}

func (s *ChatSvc) StartChat(ctx context.Context, userID1, userID2 int) (int, error) {
	if userID1 == userID2 {
		return -1, errors.New("cannot create chat with yourself")
	}
	return s.ChatRepo.CreateChat(ctx, userID1, userID2)
}

func (s *ChatSvc) PostMessage(ctx context.Context, chatID, userID int, text string) (int, error) {
	return s.ChatRepo.SendMessage(ctx, userID, chatID, text)
}

func (s *ChatSvc) GetUserChats(ctx context.Context, userID int) ([]int, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	return s.ChatRepo.GetUserChats(ctx, userID)
}

func (s *ChatSvc) GetMessages(ctx context.Context, chatID int, limit, offset int) ([]domain.Message, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.ChatRepo.GetMessages(ctx, chatID, limit, offset)
}
