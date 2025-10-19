package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/Platinaa777/DataCat/internal/entity"
	"github.com/Platinaa777/DataCat/internal/usecases/contract"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrations embed.FS

type queryer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
type Store struct {
	db   queryer
	pool *pgxpool.Pool
}

func Open(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	s := &Store{pool, pool}
	if err = s.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return s, nil
}
func (s *Store) Close()                         { s.pool.Close() }
func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }
func (s *Store) Pool() *pgxpool.Pool            { return s.pool }
func (s *Store) Migrate(ctx context.Context) error {
	return s.Within(ctx, func(store contract.Store) error {
		tx := store.(*Store)
		if _, err := tx.db.Exec(ctx, "SELECT pg_advisory_xact_lock(68323011)"); err != nil {
			return err
		}
		files, err := migrations.ReadDir("migrations")
		if err != nil {
			return err
		}
		for _, file := range files {
			b, err := migrations.ReadFile("migrations/" + file.Name())
			if err != nil {
				return err
			}
			if _, err = tx.db.Exec(ctx, string(b)); err != nil {
				return fmt.Errorf("migration %s: %w", file.Name(), err)
			}
		}
		return nil
	})
}
func (s *Store) Within(ctx context.Context, fn func(contract.Store) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	if err = fn(&Store{db: tx, pool: s.pool}); err != nil {
		return err
	}
	return translate(tx.Commit(ctx))
}
func (s *Store) Get(ctx context.Context, ns, kind, id string) (entity.Record, error) {
	var r entity.Record
	err := s.db.QueryRow(ctx, `SELECT namespace,kind,id,name,data,version FROM datacat.resources WHERE namespace=$1 AND kind=$2 AND (id=$3 OR name=$3) ORDER BY (id=$3) DESC LIMIT 1`, ns, kind, id).Scan(&r.Namespace, &r.Kind, &r.ID, &r.Name, &r.Data, &r.Version)
	return r, translate(err)
}
func (s *Store) List(ctx context.Context, ns, kind string) ([]entity.Record, error) {
	rows, err := s.db.Query(ctx, `SELECT namespace,kind,id,name,data,version FROM datacat.resources WHERE ($1='*' OR namespace=$1) AND kind=$2 ORDER BY namespace,id`, ns, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.Record, 0)
	for rows.Next() {
		var r entity.Record
		if err = rows.Scan(&r.Namespace, &r.Kind, &r.ID, &r.Name, &r.Data, &r.Version); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
func (s *Store) Create(ctx context.Context, r entity.Record) error {
	_, err := s.db.Exec(ctx, `INSERT INTO datacat.resources(namespace,kind,id,name,data) VALUES($1,$2,$3,$4,$5)`, r.Namespace, r.Kind, r.ID, r.Name, r.Data)
	return translate(err)
}
func (s *Store) Update(ctx context.Context, r entity.Record) error {
	tag, err := s.db.Exec(ctx, `UPDATE datacat.resources SET name=$4,data=$5,version=version+1,updated_at=now() WHERE namespace=$1 AND kind=$2 AND id=$3 AND version=$6`, r.Namespace, r.Kind, r.ID, r.Name, r.Data, r.Version)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() != 1 {
		return entity.ErrConflict
	}
	return nil
}
func (s *Store) Delete(ctx context.Context, ns, kind, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM datacat.resources WHERE namespace=$1 AND kind=$2 AND id=$3`, ns, kind, id)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() != 1 {
		return entity.ErrNotFound
	}
	return nil
}
func (s *Store) NextID(ctx context.Context) (int64, error) {
	var n int64
	err := s.db.QueryRow(ctx, "SELECT nextval('datacat.resource_id')").Scan(&n)
	return n, err
}
func translate(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.ErrNotFound
	}
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && (pgerr.Code == "23505" || pgerr.Code == "40001") {
		return entity.ErrConflict
	}
	return err
}
