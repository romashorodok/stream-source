package services

import (
	"context"
	"log"

	identitypb "github.com/romashorodok/stream-source/pb/go/identity/v1"
	"github.com/romashorodok/stream-source/services/identity/jwt"
	"github.com/romashorodok/stream-source/services/identity/stores"
	"github.com/romashorodok/stream-source/services/identity/types"

	"gorm.io/gorm"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	X_ACCESS_TOKEN = "x-access-token"

	DEFAULT_ROLE = "user"
)

type IdentityService struct {
	identitypb.UnimplementedIdentityServiceServer

	UserStore stores.UserStore
	JWT       *jwt.JWT
	DB        *gorm.DB
}

func (s *IdentityService) Login(ctx context.Context, req *identitypb.LoginRequest) (*identitypb.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}

func (s *IdentityService) CreateUser(ctx context.Context, req *identitypb.CreateUserRequest) (*identitypb.CreateUserResponse, error) {
	user := &types.User{}
	user.FromProto(req.GetUser())

	if err := s.UserStore.Save(user); err != nil {
		return nil, status.Error(codes.Unavailable, "Cannot save user")
	}

	if err := s.JWT.ProvideAuthentication(user, DEFAULT_ROLE); err != nil {
		log.Println(err)
		return nil, status.Error(codes.Unavailable, "Unable to create authentication data")
	}

	if err := s.UserStore.Save(user); err != nil {
		return nil, status.Error(codes.Unavailable, "Cannot save user")
	}

	return &identitypb.CreateUserResponse{
		AccessToken: &identitypb.AccessToken{
			Value: user.UserAuthentications[0].AccessToken,
		},
		RefreshToken: &identitypb.RefreshToken{
			Value: user.UserAuthentications[0].RefreshToken,
		},
	}, nil
}

func (s *IdentityService) RefreshAuthentication(ctx context.Context, req *identitypb.RefreshAuthenticationRequest) (*identitypb.RefreshAuthenticationResponse, error) {
	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Provide access token")
	}

	token := metadata.Get(X_ACCESS_TOKEN)[0]

	claims, err := s.JWT.GetAccessTokenClaims(token)

	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Unauthenticated, "Provided invalid or expired access token")
	}

	var userAuthentication *types.UserAuthentication
	user, err := s.UserStore.GetUserByAccessTokenClaims(token, claims)

	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Unauthenticated, "Unable to find user or token")
	}

	userAuthentication = &user.UserAuthentications[0]

	_, err = s.JWT.FillAuthentication(user, userAuthentication, DEFAULT_ROLE)

	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Unauthenticated, "Unable refresh tokens")
	}

	if err = s.DB.Updates(userAuthentication).Error; err != nil {
		log.Println(err)
		return nil, status.Error(codes.Unauthenticated, "Unable update token in db")
	}

	return &identitypb.RefreshAuthenticationResponse{
		AccessToken: &identitypb.AccessToken{
			Value: user.UserAuthentications[0].AccessToken,
		},
		RefreshToken: &identitypb.RefreshToken{
			Value: user.UserAuthentications[0].RefreshToken,
		},
	}, nil
}
