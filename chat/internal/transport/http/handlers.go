package httpserver

import (
	"chat/internal/domain"
	"chat/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

//go:generate mockgen -destination=./mock/mock_handlers.go -package=mock -source=handlers.go

type ChatService interface {
	StartChat(ctx context.Context, userID1, userID2 int) (int, error)
	PostMessage(ctx context.Context, chatID, userID int, text string) (int, error)
	GetMessages(ctx context.Context, chatID int, limit, offset int) ([]domain.Message, error)
	GetUserChats(ctx context.Context, userID int) ([]int, error)
}

type Handler struct {
	srv ChatService
}

func NewHandler(srv ChatService) *Handler {
	return &Handler{srv: srv}
}

func (h *Handler) GetChatsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sId := r.Context().Value(UserIdKey)
		id, err := strconv.Atoi(fmt.Sprintf("%v", sId))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		chats, err := h.srv.GetUserChats(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(GetChatsResponse{id, chats}); err != nil {
			logger.GetFromCtx(r.Context()).ErrorContext(r.Context(), "error in json encoder", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})
}

func (h *Handler) NewChatHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sId := r.Context().Value(UserIdKey)
		userId1, err := strconv.Atoi(fmt.Sprintf("%v", sId))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var req NewChatRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.GetFromCtx(r.Context()).ErrorContext(r.Context(), "error in json decoder", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		chatID, err := h.srv.StartChat(r.Context(), userId1, req.UserID2)
		if err != nil {
			http.Error(w, "Failed to create chat: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(NewChatResponse{ChatID: chatID}); err != nil {
			logger.GetFromCtx(r.Context()).ErrorContext(r.Context(), "error in json encoder", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})
}

func (h *Handler) SendMessageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		sUserId := r.Context().Value(UserIdKey)
		userId, err := strconv.Atoi(fmt.Sprintf("%v", sUserId))
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		sChatId := vars["chat_id"]
		chatId, _ := strconv.Atoi(sChatId)

		var req SendMessageRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.GetFromCtx(r.Context()).ErrorContext(r.Context(), "error in json decoder", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if chatId == 0 || userId == 0 || req.Text == "" {
			logger.GetFromCtx(r.Context()).ErrorContext(r.Context(), "Chat ID or User ID or text are required", nil)
			http.Error(w, "Chat ID or User ID or text are required", http.StatusBadRequest)
			return
		}

		messageID, err := h.srv.PostMessage(r.Context(), chatId, userId, req.Text)
		if err != nil {
			http.Error(w, "Failed to send message: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(SendMessageResponse{MessageID: messageID}); err != nil {
			logger.GetFromCtx(r.Context()).ErrorContext(r.Context(), "error in json encoder", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})
}

func (h *Handler) GetMessagesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		chatID, err := strconv.Atoi(vars["chat_id"])
		if err != nil {
			http.Error(w, "Invalid chat ID", http.StatusBadRequest)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		messages, err := h.srv.GetMessages(r.Context(), chatID, limit, offset)
		if err != nil {
			http.Error(w, "Failed to get messages: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(messages); err != nil {
			logger.GetFromCtx(r.Context()).ErrorContext(r.Context(), "error in json encoder", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})
}
