-- +goose Up
CREATE TABLE users (
    "id" UUID NOT NULL PRIMARY KEY,
    "created_at" TIMESTAMP NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP NOT NULL DEFAULT now(),
    "email" TEXT UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE users;

