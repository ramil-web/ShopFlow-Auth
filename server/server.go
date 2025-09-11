package server

import (
	"auth/proto/authpb"
	"context"
)

type AuthService struct {
	authpb.UnimplementedAuthServiceServer
}

// Реализация метода VerifyToken
func (s *AuthService) VerifyToken(ctx context.Context, req *authpb.AuthRequest) (*authpb.AuthResponse, error) {
	return &authpb.AuthResponse{
		Valid: true,
		Email: "user@example.com",
	}, nil
}
