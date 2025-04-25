package postgresql_test

import (
	"chat/internal/storage/postgresql"
	"context"
	"database/sql"
	"github.com/testcontainers/testcontainers-go"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS chats (
			id SERIAL PRIMARY KEY,
			user_1_id BIGINT NOT NULL,
			user_2_id BIGINT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_1_id, user_2_id)
		);
		
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
			sender_id BIGINT NOT NULL,
			text TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			is_read BOOLEAN DEFAULT FALSE
		);
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	return pool
}

func TestCreateChat(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t)
	storage := postgresql.New(pool)

	t.Run("successful chat create", func(t *testing.T) {
		user1 := 1
		user2 := 2
		chatID, err := storage.CreateChat(ctx, user1, user2)
		require.NoError(t, err)
		assert.Greater(t, chatID, 0)

		chats1, err := storage.GetUserChats(ctx, user1)
		require.NoError(t, err)
		assert.Contains(t, chats1, chatID)

		chats2, err := storage.GetUserChats(ctx, user2)
		require.NoError(t, err)
		assert.Contains(t, chats2, chatID)
	})
	t.Run("create chat with same userID", func(t *testing.T) {
		user1 := 3
		_, err := storage.CreateChat(ctx, user1, user1)
		require.Equal(t, sql.ErrNoRows, err)
	})

	t.Run("create multiple chats for one user", func(t *testing.T) {
		user1 := 4
		user2 := 5
		user3 := 6
		chat1, err := storage.CreateChat(ctx, user1, user2)
		require.NoError(t, err)

		chat2, err := storage.CreateChat(ctx, user1, user3)
		require.NoError(t, err)

		chats, err := storage.GetUserChats(ctx, user1)
		require.NoError(t, err)
		assert.ElementsMatch(t, []int{chat1, chat2}, chats)
	})
}

func TestGetUserChats(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t)
	storage := postgresql.New(pool)
	user1 := 1
	user2 := 2
	chatID, err := storage.CreateChat(ctx, user1, user2)
	require.NoError(t, err)
	t.Run("successful get chatID", func(t *testing.T) {
		require.NoError(t, err)
		assert.Greater(t, chatID, 0)

		chats1, err := storage.GetUserChats(ctx, user1)
		require.NoError(t, err)
		assert.Contains(t, chats1, chatID)

		chats2, err := storage.GetUserChats(ctx, user2)
		require.NoError(t, err)
		assert.Contains(t, chats2, chatID)
	})
	t.Run("multiple chats for user", func(t *testing.T) {
		userID := 3
		otherUsers := []int{4, 5, 6}
		var expectedChats []int

		for _, otherUser := range otherUsers {
			chatID, err := storage.CreateChat(ctx, userID, otherUser)
			require.NoError(t, err)
			expectedChats = append(expectedChats, chatID)
		}
		actualChats, err := storage.GetUserChats(ctx, userID)
		require.NoError(t, err)
		assert.ElementsMatch(t, expectedChats, actualChats)
	})
	t.Run("zero chats", func(t *testing.T) {
		userID := 7
		chats, err := storage.GetUserChats(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, chats)
	})
}

func TestSendMessageWithGetMessages(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t)
	storage := postgresql.New(pool)

	user1 := 1
	user2 := 2

	chatID, err := storage.CreateChat(ctx, user1, user2)
	require.NoError(t, err)

	msgID, err := storage.SendMessage(ctx, user1, chatID, "Test message")
	require.NoError(t, err)
	assert.Greater(t, msgID, 0)

	messages, err := storage.GetMessages(ctx, chatID, 10, 0)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	assert.Equal(t, "Test message", messages[0].Text)
}
