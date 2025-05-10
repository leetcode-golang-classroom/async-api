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