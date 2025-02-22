package common

import "time"

type Common struct {
	ID        uint      `json:"-" form:"id"`
	CreatedAt time.Time `json:"-" form:"created_at"`
	UpdatedAt time.Time `json:"-" form:"updated_at"`
}
