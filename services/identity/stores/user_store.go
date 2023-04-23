package stores

import (
	"github.com/romashorodok/stream-source/services/identity/jwt"
	"github.com/romashorodok/stream-source/services/identity/types"
	"gorm.io/gorm"
)

type UserStore interface {
	Save(user *types.User) error
	GetUserByAccessTokenClaims(token string, claims *jwt.AccessTokenClaims) (*types.User, error)
}

type UserStoreGORM struct {
	DB *gorm.DB
}

func (s *UserStoreGORM) Save(user *types.User) error {
	if err := s.DB.Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (s *UserStoreGORM) GetUserByAccessTokenClaims(token string, claims *jwt.AccessTokenClaims) (*types.User, error) {
	var user types.User

	err := s.DB.Joins("JOIN user_authentications ON user_authentications.user_id = ?", claims.UserId).
		Where("user_authentications.access_token = ?", token).
		Distinct().
		Preload("UserAuthentications").
		Find(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}
