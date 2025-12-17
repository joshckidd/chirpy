-- +goose Up
CREATE TABLE chirps (
    id UUID NOT NULL PRIMARY KEY
    ,created_at TIMESTAMP NOT NULL
    ,updated_at TIMESTAMP NOT NULL
    ,body TEXT NOT NULL
    ,user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;