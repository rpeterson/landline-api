package models

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type Message struct {
	Id        string    `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	RoomId    string    `db:"room_id" json:"room_id"`
	UserId    string    `db:"user_id" json:"user_id"`
	Body      string    `db:"body" json:"body"`
}

type MessageWithUser struct {
	Id        string    `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Body      string    `db:"body" json:"body"`
	Username  string    `db:"username" json:"username"`
	AvatarUrl string    `db:"avatar_url" json:"avatar_url"`
}

func NewMessageWithUser(message *Message, user *User) *MessageWithUser {
	return &MessageWithUser{
		Id:        message.Id,
		CreatedAt: message.CreatedAt,
		Body:      message.Body,
		Username:  user.Username,
		AvatarUrl: user.AvatarUrl,
	}
}

type UnreadAlert struct {
	Key        string    `json:"key"`
	Recipients *[]string `json:"recipients"`
}

func FindMessages(roomId string) ([]MessageWithUser, error) {
	var messages []MessageWithUser
	_, err := Db.Select(
		&messages,
		`SELECT messages.id, messages.created_at, body, username, avatar_url
		FROM messages INNER JOIN users ON (users.id = messages.user_id)
		WHERE messages.room_id = $1 ORDER BY messages.created_at ASC
		LIMIT 50`,
		roomId,
	)

	return messages, err
}

func CreateMessage(fields *Message) error {
	fields.CreatedAt = time.Now()
	fields.UpdatedAt = time.Now()
	err := Db.Insert(fields)

	if err != nil {
		return err
	}

	err = registerUnread(fields.RoomId)

	return err
}

func buildReadraptorRequestBody(roomId string) (*bytes.Reader, error) {
	subscribers, err := Subscribers(roomId)

	if err != nil {
		return nil, err
	}

	alert := UnreadAlert{
		Key:        roomId,
		Recipients: subscribers,
	}

	body, err := json.Marshal(alert)

	if err != nil {
		return nil, err
	}

	return bytes.NewReader(body), nil
}

func registerUnread(roomId string) error {
	body, err := buildReadraptorRequestBody(roomId)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		os.Getenv("RR_URL")+"/articles",
		body,
	)

	req.SetBasicAuth(os.Getenv("RR_PRIVATE_KEY"), "")

	client := &http.Client{}

	_, err = client.Do(req)

	return err
}
