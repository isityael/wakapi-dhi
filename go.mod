module git.m0sh1.cc/m0sh1/wakapi-dhi

go 1.26.4

require (
	github.com/alitto/pond/v2 v2.7.1
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/coreos/go-oidc/v3 v3.19.0
	github.com/dchest/captcha v1.1.0
	github.com/descope/virtualwebauthn v1.0.5
	github.com/duke-git/lancet/v2 v2.3.9
	github.com/emersion/go-sasl v0.0.0-20241020182733-b788ff22d5a6
	github.com/emersion/go-smtp v0.24.0
	github.com/getsentry/sentry-go v0.47.0
	github.com/go-chi/chi/v5 v5.3.0
	github.com/go-chi/httprate v0.15.0
	github.com/go-webauthn/webauthn v0.17.4
	github.com/gofrs/uuid/v5 v5.4.0
	github.com/gohugoio/hashstructure v0.6.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/gorilla/schema v1.4.1
	github.com/gorilla/securecookie v1.1.2
	github.com/gorilla/sessions v1.4.0
	github.com/mileusna/useragent v1.3.5
	github.com/muety/wakapi v0.0.0-20260703070800-2493da214f2e
	github.com/ncruces/go-sqlite3/gormlite v0.34.0
	github.com/oauth2-proxy/mockoidc v0.0.0-20240214162133-caebfff84d25
	github.com/robfig/cron/v3 v3.0.1
	github.com/stretchr/testify v1.11.1
	github.com/stripe/stripe-go/v76 v76.25.0
	github.com/swaggo/http-swagger/v2 v2.0.2
	github.com/swaggo/swag v1.16.6
	go.yaml.in/yaml/v3 v3.0.4
	golang.org/x/crypto v0.53.0
	golang.org/x/oauth2 v0.36.0
	gorm.io/driver/mysql v1.6.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.2
)

require (
	github.com/fxamacker/cbor/v2 v2.9.2 // indirect
	github.com/go-jose/go-jose/v3 v3.0.5 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-openapi/swag/conv v0.26.1 // indirect
	github.com/go-openapi/swag/jsonname v0.26.1 // indirect
	github.com/go-openapi/swag/jsonutils v0.26.1 // indirect
	github.com/go-openapi/swag/loading v0.26.1 // indirect
	github.com/go-openapi/swag/stringutils v0.26.1 // indirect
	github.com/go-openapi/swag/typeutils v0.26.1 // indirect
	github.com/go-openapi/swag/yamlutils v0.26.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/go-webauthn/x v0.2.6 // indirect
	github.com/google/go-tpm v0.9.8 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/ncruces/go-sqlite3 v0.35.0 // indirect
	github.com/ncruces/go-sqlite3-wasm/v2 v2.6.35302 // indirect
	github.com/ncruces/go-sqlite3-wasm/v3 v3.1.35302 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/swaggo/files/v2 v2.0.2 // indirect
	github.com/tinylib/msgp v1.6.4 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	golang.org/x/mod v0.37.0 // indirect
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-openapi/jsonpointer v0.23.1 // indirect
	github.com/go-openapi/jsonreference v0.21.6 // indirect
	github.com/go-openapi/spec v0.22.6 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/samber/lo v1.53.0 // indirect
	github.com/samber/slog-common v0.22.0 // indirect
	github.com/samber/slog-multi v1.8.0
	github.com/samber/slog-sentry/v2 v2.11.0
	github.com/stretchr/objx v0.5.3 // indirect
	golang.org/x/exp v0.0.0-20260611194520-c48552f49976 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	golang.org/x/tools v0.46.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/muety/wakapi => .

godebug x509negativeserial=1 // https://stackoverflow.com/a/79062183/3112139
