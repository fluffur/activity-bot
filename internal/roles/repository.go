package roles

import (
	"context"
)

type Repository interface {
	CreateRoleReservation(
		ctx context.Context,
		chatID int64,
		userID int64,
		roleID int64,
	) error

	DeleteRoleReservation(
		ctx context.Context,
		chatID int64,
		roleID int64,
	) error

	ListRoleReservations(
		ctx context.Context,
		chatID int64,
	) ([]RoleReservation, error)

	CreateRoleTemplate(
		ctx context.Context,
		chatID int64,
		fandomName string,
		categories []Category,
	) error

	GetRoleByName(
		ctx context.Context,
		chatID int64,
		fandomName string,
		categoryName string,
		roleName string,
	) (Role, error)

	GetRoleByAlias(
		ctx context.Context,
		chatID int64,
		fandomName string,
		alias string,
	) (Role, error)

	GetRoleByNameOrAlias(
		ctx context.Context,
		chatID int64,
		fandomName string,
		name string,
	) (Role, error)

	CreateRole(
		ctx context.Context,
		categoryID int64,
		name string,
		emoji string,
	) (Role, error)

	CreateRoleAlias(
		ctx context.Context,
		roleID int64,
		name string,
	) error

	GetFandom(
		ctx context.Context,
		chatID int64,
		name string,
	) (Fandom, error)

	GetOrCreateFandom(
		ctx context.Context,
		chatID int64,
		name string,
	) (Fandom, error)

	GetRoleCategory(
		ctx context.Context,
		fandomID int64,
		name string,
	) (Category, error)

	ListRoleCategories(
		ctx context.Context,
		fandomID int64,
	) ([]Category, error)

	CreateRoleCategory(
		ctx context.Context,
		fandomID int64,
		name string,
	) (Category, error)

	GetRoleTemplate(
		ctx context.Context,
		chatID int64,
		fandomName string,
	) (Fandom, error)

	ListRoleTemplates(
		ctx context.Context,
		chatID int64,
	) ([]Fandom, error)
}
