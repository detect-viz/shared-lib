package common

type Common struct {
	ID        uint `json:"-" form:"id"`
	CreatedAt uint `json:"-" form:"created_at"`
	UpdatedAt uint `json:"-" form:"updated_at"`
}
