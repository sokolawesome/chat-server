package models

import "time"

type User struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	HashedPassword string    `json:"-"`
	CreatedTime    time.Time `json:"created_time"`
}
