DROP TABLE refresh_sessions;
DROP INDEX users_email_normalized_unique;

ALTER TABLE users
    DROP COLUMN password_hash,
    DROP COLUMN email;
