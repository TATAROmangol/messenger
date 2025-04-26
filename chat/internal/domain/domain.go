package domain

import "time"

type Message struct {
	SenderId string   `json:"sender_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	IsRead    bool      `json:"is_read"`
}
