package service

import "time"

type AccessTokenManager interface {
	Issue(userID int64) (string, time.Time, error)
	Verify(token string) (int64, error)
}
