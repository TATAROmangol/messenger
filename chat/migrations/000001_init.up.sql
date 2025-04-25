-- Таблица чатов
CREATE TABLE chats (
                       id SERIAL PRIMARY KEY,
                       user_1_id BIGINT NOT NULL,
                       user_2_id BIGINT NOT NULL,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                       updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                       UNIQUE (user_1_id, user_2_id)
);

CREATE TABLE messages (
                          id SERIAL PRIMARY KEY,
                          chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
                          sender_id BIGINT NOT NULL,
                          text TEXT NOT NULL,
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                          is_read BOOLEAN DEFAULT FALSE
);