package repository

import (
	"context"
	"core/model"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &UserRepositoryImpl{pool: pool}
}

func (repository *UserRepositoryImpl) Create(user *model.User) (*model.User, error) {
	err := repository.pool.QueryRow(
		context.Background(),
		"INSERT INTO users (name, email) VALUES ($1, NULLIF($2, '')) RETURNING id",
		user.Name,
		user.Email,
	).Scan(&user.Id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repository *UserRepositoryImpl) GetById(id int64) (*model.User, error) {
	user := &model.User{}
	err := repository.pool.QueryRow(
		context.Background(),
		"SELECT id, name, COALESCE(email, '') FROM users WHERE id = $1",
		id,
	).Scan(&user.Id, &user.Name, &user.Email)
	if err != nil {
		return nil, mapUserLookupError(err)
	}
	return user, nil
}

func (repository *UserRepositoryImpl) Update(user *model.User) (*model.User, error) {
	err := repository.pool.QueryRow(
		context.Background(),
		"UPDATE users SET name = $2 WHERE id = $1 RETURNING id, name, COALESCE(email, '')",
		user.Id,
		user.Name,
	).Scan(&user.Id, &user.Name, &user.Email)
	if err != nil {
		return nil, mapUserLookupError(err)
	}
	return user, nil
}

func (repository *UserRepositoryImpl) Delete(id int64) error {
	commandTag, err := repository.pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return mapUserDeleteError(err)
	}
	if commandTag.RowsAffected() == 0 {
		return model.ErrUserNotFound
	}
	return nil
}

func mapUserDeleteError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && (postgresError.Code == "23001" || postgresError.Code == "23503") {
		return model.ErrUserInUse
	}
	if strings.Contains(err.Error(), "SQLSTATE 23001") || strings.Contains(err.Error(), "SQLSTATE 23503") {
		return model.ErrUserInUse
	}
	return err
}

func mapUserLookupError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrUserNotFound
	}
	return err
}
