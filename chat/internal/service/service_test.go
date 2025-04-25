package service

import (
	"chat/internal/domain"
	"chat/internal/service/mock"
	"chat/pkg/logger"
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestChatSvc_StartChat(t *testing.T) {
	cr := mock.NewMockChatRepo(gomock.NewController(t))

	l := logger.New()
	ctx := logger.InitFromCtx(context.Background(), l)

	type args struct {
		userID1 int
		userID2 int
	}

	type MockBehavor func(user1, user2 int)

	tests := []struct {
		name        string
		args        args
		MockBehavor MockBehavor
		want        int
		wantErr     bool
	}{
		{
			name: "ok",
			args: args{
				userID1: 1,
				userID2: 2,
			},
			MockBehavor: func(user1, user2 int) {
				cr.EXPECT().
					CreateChat(ctx, user1, user2).
					Return(1, nil)
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "with u1==u2",
			args: args{
				userID1: 1,
				userID2: 1,
			},
			MockBehavor: func(user1, user2 int) {},
			want:        -1,
			wantErr:     true,
		},
		{
			name: "with err repo",
			args: args{
				userID1: 1,
				userID2: 2,
			},
			MockBehavor: func(user1, user2 int) {
				cr.EXPECT().
					CreateChat(ctx, user1, user2).
					Return(-1, fmt.Errorf("test err"))
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ChatSvc{
				ChatRepo: cr,
			}
			tt.MockBehavor(tt.args.userID1, tt.args.userID2)

			got, err := s.StartChat(ctx, tt.args.userID1, tt.args.userID2)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChatSvc.StartChat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ChatSvc.StartChat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChatSvc_PostMessage(t *testing.T) {
	cr := mock.NewMockChatRepo(gomock.NewController(t))

	l := logger.New()
	ctx := logger.InitFromCtx(context.Background(), l)

	type MockBehavor func(chatID, userID int, text string)

	type args struct {
		chatID int
		userID int
		text   string
	}
	tests := []struct {
		name        string
		args        args
		MockBehavor MockBehavor
		want        int
		wantErr     bool
	}{
		{
			name: "ok",
			args: args{
				chatID: 1,
				userID: 2,
				text:   "test",
			},
			MockBehavor: func(chatID, userID int, text string) {
				cr.EXPECT().
					SendMessage(ctx, userID, chatID, text).
					Return(1, nil)
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "with err repo",
			args: args{
				chatID: 1,
				userID: 2,
				text:   "test",
			},
			MockBehavor: func(chatID, userID int, text string) {
				cr.EXPECT().
					SendMessage(gomock.Any(), userID, chatID, text).
					Return(-1, fmt.Errorf("test err"))
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ChatSvc{
				ChatRepo: cr,
			}
			tt.MockBehavor(tt.args.chatID, tt.args.userID, tt.args.text)
			got, err := s.PostMessage(ctx, tt.args.chatID, tt.args.userID, tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChatSvc.PostMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ChatSvc.PostMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChatSvc_GetUserChats(t *testing.T) {
	cr := mock.NewMockChatRepo(gomock.NewController(t))

	l := logger.New()
	ctx := logger.InitFromCtx(context.Background(), l)

	type MockBehavor func(userId int)

	type args struct {
		userID int
	}
	tests := []struct {
		name        string
		args        args
		MockBehavor MockBehavor
		want        []int
		wantErr     bool
	}{
		{
			name: "ok",
			args: args{
				userID: 1,
			},
			MockBehavor: func(userId int) {
				cr.EXPECT().
					GetUserChats(gomock.Any(), userId).
					Return([]int{1, 2}, nil)
			},
			want:    []int{1, 2},
			wantErr: false,
		},
		{
			name: "with err repo",
			args: args{
				userID: 1,
			},
			MockBehavor: func(userId int) {
				cr.EXPECT().
					GetUserChats(gomock.Any(), userId).
					Return(nil, fmt.Errorf("test err"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "with userID<0",
			args: args{
				userID: -1,
			},
			MockBehavor: func(userId int) {},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ChatSvc{
				ChatRepo: cr,
			}
			tt.MockBehavor(tt.args.userID)
			got, err := s.GetUserChats(ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChatSvc.GetUserChats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChatSvc.GetUserChats() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChatSvc_GetMessages(t *testing.T) {
	cr := mock.NewMockChatRepo(gomock.NewController(t))

	l := logger.New()
	ctx := logger.InitFromCtx(context.Background(), l)

	type MockBehavor func(chatID, limit, offset int)

	type args struct {
		chatID int
		limit  int
		offset int
	}
	tests := []struct {
		name    string
		args    args
		MockBehavor MockBehavor	
		want    []domain.Message
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				chatID: 1,
				limit:  10,
				offset: 0,
			},
			MockBehavor: func(chatID, limit, offset int) {
				cr.EXPECT().
					GetMessages(gomock.Any(), chatID, limit, offset).
					Return([]domain.Message{{"test", time.Unix(1,1), true}}, nil)
			},
			want:    []domain.Message{{"test", time.Unix(1,1), true}},
			wantErr: false,
		},
		{
			name: "with err repo",
			args: args{
				chatID: 1,
				limit:  10,
				offset: 0,
			},
			MockBehavor: func(chatID, limit, offset int) {
				cr.EXPECT().
					GetMessages(gomock.Any(), chatID, limit, offset).
					Return(nil, fmt.Errorf("test err"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "with offset<0",
			args: args{
				chatID: 1,
				limit:  10,
				offset: -1,
			},
			MockBehavor: func(chatID, limit, offset int) {
				cr.EXPECT().
					GetMessages(gomock.Any(), chatID, limit, 0).
					Return([]domain.Message{{"test", time.Unix(1,1), true}}, nil)
			},
			want:    []domain.Message{{"test", time.Unix(1,1), true}},
			wantErr: false,
		},
		{
			name: "with limit<0",
			args: args{
				chatID: 1,
				limit:  -1,
				offset: 0,
			},
			MockBehavor: func(chatID, limit, offset int) {	
				cr.EXPECT().
					GetMessages(gomock.Any(), chatID, 10, offset).
					Return([]domain.Message{{"test", time.Unix(1,1), true}}, nil)
			},
			want:    []domain.Message{{"test", time.Unix(1,1), true}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ChatSvc{
				ChatRepo: cr,
			}
			tt.MockBehavor(tt.args.chatID, tt.args.limit, tt.args.offset)
			got, err := s.GetMessages(ctx, tt.args.chatID, tt.args.limit, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChatSvc.GetMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChatSvc.GetMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}
