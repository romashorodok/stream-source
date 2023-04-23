package types

import (
	"errors"
	"time"

	"github.com/google/uuid"
	identitypb "github.com/romashorodok/stream-source/pb/go/identity/v1"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserId uuid.UUID `gorm:"primarykey;type:uuid;default:gen_random_uuid()"`

	Username string
	Password string

	UserAuthentications []UserAuthentication `gorm:"foreignKey:UserId;references:UserId"`
}

func (m *User) IsValidPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(m.Password), []byte(password))
	return err == nil
}

func (m *User) FromProto(user *identitypb.User) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.GetPassword()), bcrypt.DefaultCost)

	if err != nil {
		return errors.New("cannot hash password")
	}

	m.Password = string(hashedPass)
	m.Username = user.GetUsername()

	return nil
}

type UserAuthentication struct {
	UserAuthenticationId uuid.UUID `gorm:"primarykey;type:uuid;default:gen_random_uuid()"`
	UserId               uuid.UUID `gorm:"type:uuid;null"`

	Role                 string
	RefreshToken         string
	AccessToken          string
	RefreshTokenExpireAt time.Time
	AccessTokenExpireAt  time.Time
	Blacklisted          bool `gorm:"default:false"`
}
