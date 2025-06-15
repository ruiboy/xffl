module xffl/services/ffl

go 1.23.0

toolchain go1.24.1

require (
	github.com/99designs/gqlgen v0.17.74
	github.com/rs/cors v1.11.1
	github.com/vektah/gqlparser/v2 v2.5.27
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.25.12
	xffl/pkg v0.0.0-00010101000000-000000000000
)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.2 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/text v0.26.0 // indirect
)

replace xffl/pkg => ../../pkg
