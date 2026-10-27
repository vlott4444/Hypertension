package ds

const CurrentUserID uint = 1

type User struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Username string `json:"username"`

	Password string `json:"-"`

	Likes []HypertensionLike `json:"-"`
}

func GetCurrentUser() *User {
	return &User{
		ID:       CurrentUserID,
		Username: "test_user",
	}
}
