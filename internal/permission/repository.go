package permission

import "context"

type Repository interface {
	CommandPermission(ctx context.Context, chatID int64, name string) (Status, error)
	SetCommandPermission(ctx context.Context, chatID int64, key string, status Status) error
}
