package httpserver

import (
	"bytes"
	"chat/internal/transport/http/mock"
	"chat/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGetChatsHandler(t *testing.T) {
	cs := mock.NewMockChatService(gomock.NewController(t))

	type mockBehavior func(userID int)

	tests := []struct {
		name         string
		userId       int
		mockBehavior mockBehavior
		wantStatus   int
		want         GetChatsResponse
	}{
		{
			name:   "ok",
			userId: 1,
			mockBehavior: func(userID int) {
				cs.EXPECT().
					GetUserChats(gomock.Any(), userID).
					Return([]int{1, 2}, nil)
			},
			want: GetChatsResponse{
				Chats: []int{1, 2},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "error service",
			userId: 1,
			mockBehavior: func(userID int) {
				cs.EXPECT().
					GetUserChats(gomock.Any(), userID).
					Return(nil, fmt.Errorf("error"))
			},
			want: GetChatsResponse{
				Chats: nil,
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "error user id",
			userId: -1,
			mockBehavior: func(userID int) {
				cs.EXPECT().
					GetUserChats(gomock.Any(), userID).
					Return(nil, fmt.Errorf("error"))
			},
			want: GetChatsResponse{
				Chats: nil,
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.userId)

			h := NewHandler(cs)

			rr := httptest.NewRecorder()

			req := httptest.NewRequest("GET", "/chat", nil)
			ctx := context.WithValue(req.Context(), UserIdKey, tt.userId)
			l := logger.New()
			ctx = logger.InitFromCtx(ctx, l)
			req = req.WithContext(ctx)

			h.GetChatsHandler().ServeHTTP(rr, req)

			if tt.wantStatus != rr.Code {
				t.Errorf("GetChatsHandler status got %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			var resp GetChatsResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Errorf("GetChatsHandler response got error %v", err)
			}

			assert.Equal(t, tt.want, resp)
		})
	}
}

func TestHandler_NewChatHandler(t *testing.T) {
	cs := mock.NewMockChatService(gomock.NewController(t))

	type mockBehavior func(user1, user2 int)

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		host         int
		req          NewChatRequest
		want         NewChatResponse
		wantStatus   int
	}{
		{
			name: "ok",
			host: 1,
			req: NewChatRequest{
				UserID2: 2,
			},
			mockBehavior: func(user1, user2 int) {
				cs.EXPECT().
					StartChat(gomock.Any(), user1, user2).
					Return(1, nil)
			},
			want: NewChatResponse{
				ChatID: 1,
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "error service",
			host: 1,
			req: NewChatRequest{
				UserID2: 2,
			},
			mockBehavior: func(user1, user2 int) {
				cs.EXPECT().
					StartChat(gomock.Any(), user1, user2).
					Return(0, fmt.Errorf("error"))
			},
			want: NewChatResponse{
				ChatID: 0,
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "error user id",
			host: 1,
			req: NewChatRequest{
				UserID2: -1,
			},
			mockBehavior: func(user1, user2 int) {
				cs.EXPECT().
					StartChat(gomock.Any(), user1, user2).
					Return(0, fmt.Errorf("error"))
			},
			want: NewChatResponse{
				ChatID: 0,
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(tt.host, tt.req.UserID2)

			h := NewHandler(cs)

			rr := httptest.NewRecorder()

			requestBody, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest("POST", "/chat/create", bytes.NewBuffer(requestBody))
			ctx := context.WithValue(req.Context(), UserIdKey, tt.host)
			l := logger.New()
			ctx = logger.InitFromCtx(ctx, l)
			req = req.WithContext(ctx)

			h.NewChatHandler().ServeHTTP(rr, req)

			if tt.wantStatus != rr.Code {
				t.Errorf("GetChatsHandler status got %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantStatus != 200 && tt.wantStatus != 201 {
				return
			}

			var resp NewChatResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Errorf("GetChatsHandler response got error %v", err)
			}

			assert.Equal(t, tt.want, resp)
		})
	}
}

func TestHandler_SendMessageHandler(t *testing.T) {
	cs := mock.NewMockChatService(gomock.NewController(t))

	type mockBehavior func(chatId, userId int, text string)

	tests := []struct {
		name   string
		mockBehavior mockBehavior
		req 	SendMessageRequest
		resp 	SendMessageResponse
		wantStatus int
	}{
		{
			name: "ok",
			req: SendMessageRequest{
				Text: "Hello",
			},
			mockBehavior: func(chatId, userId int, text string) {
				cs.EXPECT().
					PostMessage(gomock.Any(), chatId, userId, text).
					Return(8, nil)
			},
			resp: SendMessageResponse{
				MessageID: 8,
			},
			wantStatus: http.StatusCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(1, 1, tt.req.Text)

			h := NewHandler(cs)
			router := mux.NewRouter()
			router.Handle("/send/{chat_id:[0-9]+}", h.SendMessageHandler())

			rr := httptest.NewRecorder()

			requestBody, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest("POST", "/send/1", bytes.NewBuffer(requestBody))
			ctx := context.WithValue(req.Context(), UserIdKey, 1)
			l := logger.New()
			ctx = logger.InitFromCtx(ctx, l)
			req = req.WithContext(ctx)

			router.ServeHTTP(rr, req)

			if tt.wantStatus != rr.Code {
				t.Errorf("GetChatsHandler status got %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantStatus != 200 && tt.wantStatus != 201 {
				return
			}

			var resp SendMessageResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Errorf("GetChatsHandler response got error %v", err)
			}

			assert.Equal(t, tt.resp, resp)
		})
	}
}
