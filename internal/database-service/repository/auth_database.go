package repository

import (
	"context"
	"core/model"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) AuthRepository {
	return &AuthRepositoryImpl{pool: pool}
}

func (repository *AuthRepositoryImpl) CreateUser(ctx context.Context, credential *model.AuthCredential) (*model.User, error) {
	user := credential.User
	err := repository.pool.QueryRow(ctx,
		"INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id",
		user.Name,
		user.Email,
		credential.PasswordHash,
	).Scan(&user.Id)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return nil, model.ErrEmailAlreadyExists
		}
		return nil, err
	}
	return &user, nil
}

func (repository *AuthRepositoryImpl) GetCredentialByEmail(ctx context.Context, email string) (*model.AuthCredential, error) {
	credential := &model.AuthCredential{}
	err := repository.pool.QueryRow(ctx,
		"SELECT id, name, email, password_hash FROM users WHERE LOWER(email) = LOWER($1) AND password_hash IS NOT NULL",
		email,
	).Scan(
		&credential.User.Id,
		&credential.User.Name,
		&credential.User.Email,
		&credential.PasswordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrCredentialNotFound
	}
	if err != nil {
		return nil, err
	}
	return credential, nil
}

func (repository *AuthRepositoryImpl) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	user := &model.User{}
	err := repository.pool.QueryRow(ctx,
		"SELECT id, name, email FROM users WHERE id = $1 AND email IS NOT NULL",
		userID,
	).Scan(&user.Id, &user.Name, &user.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrCredentialNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
