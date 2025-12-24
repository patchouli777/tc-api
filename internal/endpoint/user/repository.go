package user

import (
	"context"
	"errors"
	"main/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{pool: pool}
}

func (r *RepositoryImpl) Get(ctx context.Context, username string) (*User, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	queries := db.New(conn)
	q := NewDBAdapter(queries)

	res, err := q.SelectByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return &User{Id: int(res.ID),
		Name:            res.Name,
		IsBanned:        res.IsBanned.Bool,
		IsPartner:       res.IsPartner.Bool,
		FirstLivestream: res.FirstLivestream.Time,
		LastLivestream:  res.LastLivestream.Time,
		Avatar:          res.Avatar.String}, nil
}

func (r *RepositoryImpl) Create(ctx context.Context, u UserCreate) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	queries := db.New(conn)
	q := NewDBAdapter(queries)

	_, err = q.Insert(ctx, db.UserInsertParams{Name: u.Name,
		Password: u.Password,
		Avatar:   pgtype.Text{String: u.Avatar, Valid: true}})

	return err
}

func (r *RepositoryImpl) Update(ctx context.Context, u UserUpdate) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return errors.New("not implemented")
}

func (r *RepositoryImpl) Delete(ctx context.Context, id int) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	queries := db.New(conn)
	q := NewDBAdapter(queries)

	return q.Delete(ctx, int32(id))
}

func (r *RepositoryImpl) List(ctx context.Context, ul UserList) ([]User, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	return nil, errors.New("not implemented")
}
