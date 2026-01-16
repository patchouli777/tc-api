package user

import (
	"context"
	"errors"
	"fmt"
	"main/internal/external/db"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{pool: pool}
}

func (r *RepositoryImpl) Get(ctx context.Context, id int32) (*User, error) {
	q := QueriesAdapter{queries: db.New(r.pool)}

	res, err := q.Select(ctx, id)
	if err != nil {
		return nil, err
	}

	return &User{
		Id:              res.ID,
		Name:            res.Name,
		IsBanned:        res.IsBanned.Bool,
		IsPartner:       res.IsPartner.Bool,
		FirstLivestream: res.FirstLivestream.Time,
		LastLivestream:  res.LastLivestream.Time,
		Avatar:          res.Avatar.String}, nil
}

func (r *RepositoryImpl) Create(ctx context.Context, u UserCreate) error {
	q := QueriesAdapter{queries: db.New(r.pool)}

	err := q.Insert(ctx, db.UserInsertParams{Name: u.Name,
		Password: u.Password,
		Avatar:   pgtype.Text{String: u.Avatar, Valid: true}})

	return err
}

func (r *RepositoryImpl) Update(ctx context.Context, id int32, upd UserUpdate) error {
	q := QueriesAdapter{queries: db.New(r.pool)}

	err := q.Update(ctx, db.UserUpdateParams{
		ID: id,

		NameDoUpdate: upd.Name.Explicit && !upd.Name.IsNull,
		Name:         upd.Name.Value,

		PasswordDoUpdate: upd.Password.Explicit && !upd.Password.IsNull,
		Password:         upd.Password.Value,

		IsBannedDoUpdate: upd.IsBanned.Explicit && !upd.IsBanned.IsNull,
		IsBanned:         pgtype.Bool{Bool: upd.IsBanned.Value, Valid: true},

		IsPartnerDoUpdate: upd.IsPartner.Explicit && !upd.IsPartner.IsNull,
		IsPartner:         pgtype.Bool{Bool: upd.IsPartner.Value, Valid: true},

		AvatarDoUpdate: upd.Avatar.Explicit && !upd.Avatar.IsNull,
		Avatar:         pgtype.Text{String: upd.Avatar.Value, Valid: upd.Avatar.IsNull},
	})

	if err != nil {
		fmt.Println(err)

		return err
	}

	return nil
}

func (r *RepositoryImpl) Delete(ctx context.Context, id int32) error {
	q := QueriesAdapter{queries: db.New(r.pool)}

	return q.Delete(ctx, id)
}

func (r *RepositoryImpl) List(ctx context.Context, ul UserList) ([]User, error) {
	return nil, errors.New("not implemented")
}
