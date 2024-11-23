-include .env
export

LIBRARY_SYSTEM_MIGRATIONS_DIR:="migrations/library-system"
RATING_SYSTEM_MIGRATIONS_DIR:="migrations/rating-system"
RESERVATION_SYSTEM_MIGRATIONS_DIR:="migrations/reservation-system"

.PHONY: library-system-migrate-up
library-system-migrate-up:
	goose -dir $(LIBRARY_SYSTEM_MIGRATIONS_DIR) postgres "${LIBRARY_SYSTEM_POSTGRESQL_DSN}" up

.PHONY: library-system-migrate-down
library-system-migrate-down:
	goose -dir $(LIBRARY_SYSTEM_MIGRATIONS_DIR) postgres "${LIBRARY_SYSTEM_POSTGRESQL_DSN}" down

.PHONY: rating-system-migrate-up
rating-system-migrate-up:
	goose -dir $(RATING_SYSTEM_MIGRATIONS_DIR) postgres "${RATING_SYSTEM_POSTGRESQL_DSN}" up

.PHONY: rating-system-migrate-down
rating-system-migrate-down:
	goose -dir $(RATING_SYSTEM_MIGRATIONS_DIR) postgres "${RATING_SYSTEM_POSTGRESQL_DSN}" down

.PHONY: reservation-system-migrate-up
reservation-system-migrate-up:
	goose -dir $(RESERVATION_SYSTEM_MIGRATIONS_DIR) postgres "${RESERVATION_SYSTEM_POSTGRESQL_DSN}" up

.PHONY: reservation-system-migrate-down
reservation-system-migrate-down:
	goose -dir $(RESERVATION_SYSTEM_MIGRATIONS_DIR) postgres "${RESERVATION_SYSTEM_POSTGRESQL_DSN}" down

.PHONY: create-library-system-migration
create-library-system-migration:
ifeq ($(name),)
	@echo "You forgot to add migration name, example:\nmake create-migration name=create_users_table"
else
	goose -dir $(LIBRARY_SYSTEM_MIGRATIONS_DIR) create $(name) sql
endif

.PHONY: create-rating-system-migration
create-rating-system-migration:
ifeq ($(name),)
	@echo "You forgot to add migration name, example:\nmake create-migration name=create_users_table"
else
	goose -dir $(RATING_SYSTEM_MIGRATIONS_DIR) create $(name) sql
endif

.PHONY: create-reservation-system-migration
create-reservation-system-migration:
ifeq ($(name),)
	@echo "You forgot to add migration name, example:\nmake create-migration name=create_users_table"
else
	goose -dir $(RESERVATION_SYSTEM_MIGRATIONS_DIR) create $(name) sql
endif