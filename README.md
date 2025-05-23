## golang-async-api

This repository is for demo how to implementation async api with golang

## architecture concept

![async-api-with-sqs](async-api-with-sqs.png)

## setup postgresql with docker-compose

```yaml
services:
  postgres:
    restart: always
    image: postgres:16
    container_name: postgres-docker-instance-for-async-api
    volumes:
      - ${HOST_DIR}:/var/lib/postgresql/data
    expose:
      - 5432
    ports:
      - ${POSTGRES_PORT}:5432
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    logging:
      driver: "json-file"
      options:
        max-size: "1k"
        max-file: "3"
```

## setup db migration with go migration

https://github.com/golang-migrate/migrate

```yaml
tasks:
  db_create_migration:
    cmds:
      - migrate create -ext sql -dir migrations -seq {{.table_name}}
    silent: true
    requires:
      vars: [table_name]

  db_migrate:
    cmds:
      - migrate -database $DB_URL -path migrations {{.action}}
    silent: true
    requires:
      vars: [action]
    ## action should be up or down
```

## setup custom handler response with wrapper function

```golang
// Handler - handler that will handle error message
func Handler(fn func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// handle error message
		if err := fn(w, r); err != nil {
			status := http.StatusInternalServerError
			msg := http.StatusText(status)
			if e, ok := err.(*ErrWithStatus); ok {
				status = e.status
				msg = http.StatusText(e.status)
				if status == http.StatusBadRequest || status == http.StatusConflict {
					msg = e.err.Error()
				}
			}
			log := logger.FromContext(r.Context())
			log.ErrorContext(r.Context(),
				"error executing handler",
				slog.Any("err", err),
				slog.Int("status", status),
				slog.String("msg", msg),
			)
			w.Header().Set("Content-Type", "application/json;charset=utf-8")
			w.WriteHeader(status)
			if err := json.NewEncoder(w).Encode(response.ApiResponse[struct{}]{
				Message: msg,
			}); err != nil {
				log.ErrorContext(r.Context(), "error encoding response", slog.Any("err", err))
			}
		}
	}
}
```

## setup decode function response with generic

```golang
type Validator interface {
	Validate(_validator *validator.Validate) error
}

// Decode - decode and vaildate input request body
func Decode[T Validator](r *http.Request, _validator *validator.Validate) (T, error) {
	var t T
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		return t, fmt.Errorf("decoding request body: %w", err)
	}
	if err := t.Validate(_validator); err != nil {
		return t, err
	}
	return t, nil
}
```

## setup encode function

```golang
// Encode - encode response body
func Encode[T any](v T, status int, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encoding response: %w", err)
	}
	return nil
}
```

## setup jwt 

```golang
var signingMethod = jwt.SigningMethodHS256

type JWTManager struct {
	config *config.Config
}

type TokenPair struct {
	AccessToken  *jwt.Token
	RefreshToken *jwt.Token
}

type CustomClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func NewJWTManager(config *config.Config) *JWTManager {
	return &JWTManager{
		config: config,
	}
}

// Parse - parse token to token.Claim, so that the custom claim could load setup
func (jwtManager *JWTManager) Parse(token string) (*jwt.Token, error) {
	parser := jwt.NewParser()
	jwtToken, err := parser.Parse(token, func(t *jwt.Token) (any, error) {
		if t.Method != signingMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtManager.config.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	return jwtToken, nil
}

// GenerateTokenPair - generate accessToken, refreshToken
func (jwtManager *JWTManager) GenerateTokenPair(userID uuid.UUID) (*TokenPair, error) {
	now := time.Now()
	issuer := fmt.Sprintf("http://%s:%s", jwtManager.config.JWTServerHost, jwtManager.config.Port)
	jwtAccessToken := jwt.NewWithClaims(signingMethod,
		CustomClaims{
			TokenType: "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userID.String(),
				Issuer:    issuer,
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute * 15)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		})

	key := []byte(jwtManager.config.JWTSecret)
	signedAccessToken, err := jwtAccessToken.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}
	accessToken, err := jwtManager.Parse(signedAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	jwtRefreshToken := jwt.NewWithClaims(signingMethod,
		CustomClaims{
			TokenType: "refresh",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userID.String(),
				Issuer:    issuer,
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24 * 30)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		})
	signedRefreshToken, err := jwtRefreshToken.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}
	refreshToken, err := jwtManager.Parse(signedRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}

// IsAccessToken - check if this token is access token
func (jwtManager *JWTManager) IsAccessToken(token *jwt.Token) bool {
	jwtClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	if tokenType, ok := jwtClaims["token_type"]; ok {
		return tokenType == "access"
	}
	return false
}
```

## setup terraorm

https://docs.localstack.cloud/user-guide/integrations/terraform/

https://developer.hashicorp.com/terraform/language/values/variables

https://registry.terraform.io/providers/rgeraskin/aws2/latest/docs/resources/sqs_queue

```yaml=
variable "aws_secret_access_key" {
  type = string
}

variable "aws_access_key_id" {
  type = string
}

variable "aws_default_region" {
  type = string
}

variable "s3_bucket" {
  type = string
}

variable "sqs_queue" {
  type = string
}


provider "aws" {
  access_key                  = var.aws_access_key_id
  secret_key                  = var.aws_secret_access_key
  region                      = var.aws_default_region

  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    s3             = "http://s3.localhost.localstack.cloud:4566"
    sqs            = "http://localhost:4566"
  }
}

resource "aws_s3_bucket" "reports-s3-bucket" {
  bucket = var.s3_bucket
}

resource "aws_sqs_queue" "reports_sqs_queue" {
  name                      = var.sqs_queue
  delay_seconds             = 5
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}
```

## run up terraform

```shell=
terraform init
task terraform_plan
task terraform_apply
```

## golang package for s3 bucket and sqs

```shell=
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/service/sqs
```