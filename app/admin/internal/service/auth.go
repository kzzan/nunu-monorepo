package service

import (
	"context"
	"errors"
	"time"

	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// Login 校验账号密码并签发 JWT：用户不存在或密码不符时统一返回
// 401 语义的 ErrUnauthorized，不区分两种失败原因以避免账号枚举。
func (s *adminService) Login(ctx context.Context, req *v1.LoginRequest) (string, error) {
	user, err := s.adminRepository.GetAdminUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", v1.ErrUnauthorized
		}
		return "", v1.ErrInternalServerError
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", err
	}
	token, err := s.jwt.GenToken(user.ID, time.Now().Add(time.Hour*24*90))
	if err != nil {
		return "", err
	}

	return token, nil
}
