package model

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     Role   `json:"role" gorm:"foreignKey:RoleID"`
	RoleID   int64  `json:"role_id"`
}