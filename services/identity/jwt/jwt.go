package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/romashorodok/stream-source/services/identity/types"
)

type JWT struct {
	SecretKey string

	AccessTokenDurationInSec  uint32
	RefreshTokenDurationInSec uint32
}

type AccessTokenClaims struct {
	jwt.StandardClaims

	Username string
	UserId   string
}

type RefreshTokenClaims struct {
	jwt.StandardClaims

	Username string
	UserId   string
}

func (s *JWT) FillAuthentication(user *types.User, auth *types.UserAuthentication, role string) (*types.UserAuthentication, error) {
	refreshToken, refreshTokenExpireAt, err := s.GenerateRefreshToken(user)
	if err != nil {
		return nil, errors.New("unable to generate refresh token")
	}

	accessToken, accessTokenTokenExpireAt, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, errors.New("unable to generate access token")
	}

	auth.RefreshToken = refreshToken
	auth.RefreshTokenExpireAt = refreshTokenExpireAt
	auth.AccessToken = accessToken
	auth.AccessTokenExpireAt = accessTokenTokenExpireAt
	auth.Role = role

	return auth, nil
}

func (s *JWT) ProvideAuthentication(user *types.User, role string) error {
	auth, err := s.FillAuthentication(user, &types.UserAuthentication{}, role)

	if err != nil {
		return err
	}

	user.UserAuthentications = append(user.UserAuthentications, *auth)

	return nil
}

func (m *JWT) GetAccessTokenClaims(accessToken string) (*AccessTokenClaims, error) {

	var parseFunc jwt.Keyfunc = func(token *jwt.Token) (interface{}, error) {

		_, ok := token.Method.(*jwt.SigningMethodHMAC)

		if !ok {
			return nil, errors.New("undefined signing method")
		}

		return []byte(m.SecretKey), nil
	}

	token, err := jwt.ParseWithClaims(accessToken, &AccessTokenClaims{}, parseFunc)

	if err != nil {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*AccessTokenClaims)

	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (m *JWT) GetRefreshTokenClaims(refreshToken string) (*RefreshTokenClaims, error) {

	var parseFunc jwt.Keyfunc = func(token *jwt.Token) (interface{}, error) {

		_, ok := token.Method.(*jwt.SigningMethodHMAC)

		if !ok {
			return nil, errors.New("undefined signing method")
		}

		return []byte(m.SecretKey), nil
	}

	token, err := jwt.ParseWithClaims(refreshToken, &RefreshTokenClaims{}, parseFunc)

	if err != nil {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)

	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (m *JWT) GenerateAccessToken(user *types.User) (string, time.Time, error) {
	duration := time.Duration(m.AccessTokenDurationInSec) * time.Second
	expireAt := time.Now().Add(duration)

	claims := &AccessTokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireAt.Unix(),
		},
		Username: user.Username,
		UserId:   user.UserId.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, err := token.SignedString([]byte(m.SecretKey))
	return tokenstring, expireAt, err
}

func (m *JWT) GenerateRefreshToken(user *types.User) (string, time.Time, error) {
	duration := time.Duration(m.RefreshTokenDurationInSec) * time.Second
	expireAt := time.Now().Add(duration)

	claims := &RefreshTokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireAt.Unix(),
		},
		Username: user.Username,
		UserId:   user.UserId.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, err := token.SignedString([]byte(m.SecretKey))
	return tokenstring, expireAt, err
}
