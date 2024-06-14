package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-workerd-api/core/generated/model"
)

func CreateUser(ctx context.Context, user *model.User) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(
		`INSERT INTO users (uuid, username, pass_hash, user_email, wallet_address, role, referrer, referral_code, created_at)
			VALUES (:uuid, :username, :pass_hash, :user_email, :wallet_address, :role, :referrer, :referral_code, :created_at);`,
	), user)
	return err
}

func ResetPassword(ctx context.Context, passHash, username string) error {
	_, err := DB.DB.ExecContext(ctx, fmt.Sprintf(
		`UPDATE users SET pass_hash = '%s', updated_at = now() WHERE username = '%s'`, passHash, username))
	return err
}

func GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var out model.User
	if err := DB.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM users WHERE username = ?`), username,
	).StructScan(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
