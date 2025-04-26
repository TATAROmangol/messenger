package postgresql

import (
	"chat/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatStorage struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *ChatStorage {
	return &ChatStorage{db: db}
}

func (s *ChatStorage) GetUserChats(ctx context.Context, userID int) ([]int, error) {
	query := `
        SELECT id
        FROM chats
        WHERE user_1_id = $1 OR user_2_id = $1
        ORDER BY id`
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chats: %w", err)
	}
	defer rows.Close()
	var chats []int
	for rows.Next() {
		var chatID int
		if err := rows.Scan(&chatID); err != nil {
			return nil, fmt.Errorf("failed to scan chat ID: %w", err)
		}
		chats = append(chats, chatID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return chats, nil
}

func (s *ChatStorage) CreateChat(ctx context.Context, userid1, userid2 int) (int, error) {
	if userid1 == userid2 {
		return 0, sql.ErrNoRows
	}
	var chatID int

	query := `
        INSERT INTO chats (user_1_id, user_2_id) 
        VALUES ($1, $2)
        RETURNING id
    `

	err := s.db.QueryRow(ctx, query, userid1, userid2).Scan(&chatID)
	if err != nil {
		return 0, err
	}

	return chatID, nil
}

func (s *ChatStorage) SendMessage(ctx context.Context, userId, chatId int, text string) (int, error) {
	var messId int

	query := `
        INSERT INTO messages (chat_id, sender_id, text) 
        VALUES ($1, $2, $3)
        RETURNING id
    `

	err := s.db.QueryRow(ctx, query, chatId, userId, text).Scan(&messId)
	if err != nil {
		return 0, err
	}

	return messId, nil
}

func (s *ChatStorage) GetMessages(ctx context.Context, chatID int, limit, offset int) ([]domain.Message, error) {
	query := `
		SELECT sender_id,text, created_at, is_read 
		FROM messages 
		WHERE chat_id = $1 
		ORDER BY created_at ASC 
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(ctx, query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		err := rows.Scan(&msg.SenderId, &msg.Text, &msg.CreatedAt, &msg.IsRead)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
