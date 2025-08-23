package client

import "time"

type Passport struct {
	ID                     string    `json:"id" db:"id"`
	UserID                 string    `json:"user_id" db:"user_id"`
	PhotoPathWithAuthor    string    `json:"photo_path_with_author" db:"photo_path_with_author"`
	PhotoPathWithoutAuthor string    `json:"photo_path_without_author" db:"photo_path_without_author"`
	Status                 string    `json:"status" db:"status"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

type Message struct {
	ID        string `json:"id" db:"id"`
	ChatID    string `json:"chat_id" db:"chat_id"`
	UserID    string `json:"user_id" db:"user_id"`
	Text      string `json:"text" db:"text"`
	CreatedAt string `json:"created_at" db:"created_at"`
}

type entry interface {
	Passport | Message
}

type GenericRepo[T entry] interface {
	Create(T) error
	Get(id string) (T, error)
	Update(T) error
	Delete(id string) error
}
