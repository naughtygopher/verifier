package stores

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/fatih/structs"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/naughtygopher/verifier"
)

// structToMapStringWithTag converts a struct to map[string]interface{}, where keys are fetched from
// provided tag values
func structToMapStringWithTag(tag string, source interface{}) (map[string]interface{}, error) {
	converter := structs.New(source)
	converter.TagName = tag
	return converter.Map(), nil
}

// PostgresConfig holds all configuration required for postgres
type PostgresConfig struct {
	Host      string `json:"host,omitempty"`
	Port      string `json:"port,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	StoreName string `json:"storeName,omitempty"`
	PoolSize  int    `json:"poolSize,omitempty"`
	SSLMode   string `json:"sslMode,omitempty"`

	DialTimeout  time.Duration `json:"dialTimeoutSecs,omitempty"`
	ReadTimeout  time.Duration `json:"readTimeoutSecs,omitempty"`
	WriteTimeout time.Duration `json:"writeTimeoutSecs,omitempty"`
	IdleTimeout  time.Duration `json:"idleTimeoutSecs,omitempty"`

	TableName string `json:"tableName,omitempty"`
}

// ConnURL returns the connection URL
func (pgcfg *PostgresConfig) ConnURL() string {
	sslMode := strings.TrimSpace(pgcfg.SSLMode)
	if sslMode == "" {
		sslMode = "disable"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		pgcfg.Username,
		pgcfg.Password,
		pgcfg.Host,
		pgcfg.Port,
		pgcfg.StoreName,
		sslMode,
	)
}

// Postgres implements the verifier store functions using Postgresql as the persistence layer
type Postgres struct {
	cfg       *PostgresConfig
	tableName string
	pqdriver  *pgxpool.Pool
	qbuilder  squirrel.StatementBuilderType
}

func ctxWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(
		ctx,
		timeout,
	)
}

// Create creates a new entry of verifier request
func (pgs *Postgres) Create(req *verifier.Request) (*verifier.Request, error) {
	reqmap, err := structToMapStringWithTag("json", req)
	if err != nil {
		return nil, err
	}

	query, args, err := pgs.qbuilder.Insert(pgs.tableName).SetMap(reqmap).ToSql()
	if err != nil {
		return nil, err
	}

	ctx, _ := ctxWithTimeout(context.Background(), pgs.cfg.WriteTimeout)
	_, err = pgs.pqdriver.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// ReadLastPending reads the last pending verification request of the commtype + recipient
func (pgs *Postgres) ReadLastPending(ctype verifier.CommType, recipient string) (*verifier.Request, error) {
	query, args, err := pgs.qbuilder.Select(
		"id",
		"type",
		"sender",
		"recipient",
		"data",
		"secret",
		"secretExpiry",
		"attempts",
		"commStatus",
		"status",
		"createdAt",
		"updatedAt",
	).From(
		pgs.tableName,
	).OrderBy(
		"autoID DESC",
	).Limit(
		1,
	).Where(squirrel.Eq{
		"type":   ctype,
		"status": verifier.VerStatusPending,
	}).ToSql()
	if err != nil {
		return nil, err
	}

	ctx, _ := ctxWithTimeout(context.Background(), pgs.cfg.ReadTimeout)
	row := pgs.pqdriver.QueryRow(
		ctx,
		query,
		args...,
	)

	req := &verifier.Request{
		SecretExpiry: new(time.Time),
		CreatedAt:    new(time.Time),
		UpdatedAt:    new(time.Time),
		Data:         map[string]string{},
		CommStatus:   make([]verifier.CommStatus, 0, 10),
	}

	id := new(sql.NullString)
	commtype := new(sql.NullString)
	sender := new(sql.NullString)
	storedRecipient := new(sql.NullString)
	secret := new(sql.NullString)
	attempts := new(sql.NullInt32)

	err = row.Scan(
		id,
		commtype,
		sender,
		storedRecipient,
		&req.Data,
		secret,
		req.SecretExpiry,
		attempts,
		&req.CommStatus,
		&req.Status,
		req.CreatedAt,
		req.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	req.ID = id.String
	req.Type = verifier.CommType(commtype.String)
	req.Sender = sender.String
	req.Recipient = storedRecipient.String
	req.Secret = secret.String
	req.Attempts = int(attempts.Int32)
	req.Sender = sender.String

	return req, nil
}

// Update updates a verification request for the given verification ID & the payload
func (pgs *Postgres) Update(verID string, req *verifier.Request) (*verifier.Request, error) {
	vermap, err := structToMapStringWithTag("json", req)
	if err != nil {
		return nil, err
	}

	query, args, err := pgs.qbuilder.Update(
		pgs.tableName,
	).SetMap(
		vermap,
	).Where(
		squirrel.Eq{"id": verID},
	).ToSql()
	if err != nil {
		return nil, err
	}

	_, err = pgs.pqdriver.Exec(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPostgres returns a new instance of Postgres with all the required fields initialized
func NewPostgres(cfg *PostgresConfig) (*Postgres, error) {
	poolcfg, err := pgxpool.ParseConfig(cfg.ConnURL())
	if err != nil {
		return nil, err
	}

	poolcfg.MaxConnLifetime = cfg.IdleTimeout
	poolcfg.MaxConns = int32(cfg.PoolSize)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolcfg)
	if err != nil {
		return nil, err
	}

	pg := &Postgres{
		cfg:       cfg,
		tableName: cfg.TableName,
		pqdriver:  pool,
		qbuilder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return pg, nil
}
