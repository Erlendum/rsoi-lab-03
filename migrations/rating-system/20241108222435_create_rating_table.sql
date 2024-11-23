-- +goose Up
-- +goose StatementBegin
CREATE TABLE rating
(
    id       SERIAL PRIMARY KEY,
    username VARCHAR(80) NOT NULL,
    stars    INT         NOT NULL
    CHECK (stars BETWEEN 0 AND 100)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rating;
-- +goose StatementEnd
