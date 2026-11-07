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

var _ AuthRepository = (*AuthRepositoryImpl)(nil)

func NewAuthRepository(pool *pgxpool.Pool) AuthRepository {
	return &AuthRepositoryImpl{pool: pool}
}

func (repository *AuthRepositoryImpl) CreateIdentity(ctx context.Context, credential *model.AuthCredential) (*model.AuthIdentity, error) {
	identity := credential.Identity
	err := repository.pool.QueryRow(ctx,
		"INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id",
		identity.Name,
		identity.Email,
		credential.PasswordHash,
	).Scan(&identity.UserID)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return nil, model.ErrEmailAlreadyExists
		}
		return nil, err
	}
	return &identity, nil
}

func (repository *AuthRepositoryImpl) GetCredentialByEmail(ctx context.Context, email string) (*model.AuthCredential, error) {
	credential := &model.AuthCredential{}
	err := repository.pool.QueryRow(ctx,
		"SELECT id, name, email, password_hash FROM users WHERE LOWER(email) = LOWER($1) AND password_hash IS NOT NULL",
		email,
	).Scan(
		&credential.Identity.UserID,
		&credential.Identity.Name,
		&credential.Identity.Email,
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

func (repository *AuthRepositoryImpl) GetIdentityByID(ctx context.Context, userID int64) (*model.AuthIdentity, error) {
	identity := &model.AuthIdentity{}
	err := repository.pool.QueryRow(ctx,
		"SELECT id, name, email FROM users WHERE id = $1 AND email IS NOT NULL",
		userID,
	).Scan(&identity.UserID, &identity.Name, &identity.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrCredentialNotFound
	}
	if err != nil {
		return nil, err
	}
	return identity, nil
}

func (repository *AuthRepositoryImpl) CreateRefreshSession(ctx context.Context, session *model.RefreshSession) error {
	_, err := repository.pool.Exec(ctx,
		"INSERT INTO refresh_sessions (user_id, token_hash, expires_at) VALUES ($1, $2, $3)",
		session.UserID,
		session.TokenHash,
		session.ExpiresAt,
	)
	return err
}

func (repository *AuthRepositoryImpl) RotateRefreshSession(ctx context.Context, currentTokenHash string, replacement *model.RefreshSession) (*model.AuthIdentity, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var sessionID int64
	var userID int64
	err = tx.QueryRow(ctx,
		`SELECT id, user_id
FROM refresh_sessions
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
FOR UPDATE`,
		currentTokenHash,
	).Scan(&sessionID, &userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrRefreshSessionInvalid
	}
	if err != nil {
		return nil, err
	}

	var replacementID int64
	err = tx.QueryRow(ctx,
		"INSERT INTO refresh_sessions (user_id, token_hash, expires_at) VALUES ($1, $2, $3) RETURNING id",
		userID,
		replacement.TokenHash,
		replacement.ExpiresAt,
	).Scan(&replacementID)
	if err != nil {
		return nil, err
	}

	commandTag, err := tx.Exec(ctx,
		"UPDATE refresh_sessions SET revoked_at = NOW(), replaced_by_id = $2 WHERE id = $1 AND revoked_at IS NULL",
		sessionID,
		replacementID,
	)
	if err != nil {
		return nil, err
	}
	if commandTag.RowsAffected() != 1 {
		return nil, model.ErrRefreshSessionInvalid
	}

	identity := &model.AuthIdentity{}
	err = tx.QueryRow(ctx,
		"SELECT id, name, email FROM users WHERE id = $1 AND email IS NOT NULL",
		userID,
	).Scan(&identity.UserID, &identity.Name, &identity.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrCredentialNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return identity, nil
}

func (repository *AuthRepositoryImpl) RevokeRefreshSession(ctx context.Context, tokenHash string) error {
	_, err := repository.pool.Exec(ctx,
		"UPDATE refresh_sessions SET revoked_at = COALESCE(revoked_at, NOW()) WHERE token_hash = $1",
		tokenHash,
	)
	return err
}
