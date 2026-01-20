package user

import (
	"context"
	"errors"
	"main/internal/external/db"
	d "main/internal/user/domain"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *RepositoryImpl {
	return &RepositoryImpl{pool: pool}
}

func (r *RepositoryImpl) Get(ctx context.Context, id int32) (*d.User, error) {
	q := queriesAdapter{queries: db.New(r.pool)}

	res, err := q.Select(ctx, id)
	if err != nil {
		return nil, err
	}

	return &d.User{
		Id:              res.ID,
		Name:            res.Name,
		IsBanned:        res.IsBanned.Bool,
		IsPartner:       res.IsPartner.Bool,
		FirstLivestream: res.FirstLivestream.Time,
		LastLivestream:  res.LastLivestream.Time,
		Pfp:             res.Pfp.String}, nil
}

func (r *RepositoryImpl) Create(ctx context.Context, u d.UserCreate) error {
	q := queriesAdapter{queries: db.New(r.pool)}

	err := q.Insert(ctx, db.UserInsertParams{
		Name:     u.Name,
		Password: u.Password,
		Pfp:      pgtype.Text{String: u.Pfp.Value, Valid: u.Pfp.Explicit && !u.Pfp.IsNull}})

	return err
}

func (r *RepositoryImpl) Update(ctx context.Context, id int32, upd d.UserUpdate) error {
	q := queriesAdapter{queries: db.New(r.pool)}

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

		PfpDoUpdate: upd.Pfp.Explicit && !upd.Pfp.IsNull,
		Pfp:         pgtype.Text{String: upd.Pfp.Value, Valid: upd.Pfp.IsNull},
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryImpl) Delete(ctx context.Context, id int32) error {
	q := queriesAdapter{queries: db.New(r.pool)}

	return q.Delete(ctx, id)
}

func (r *RepositoryImpl) List(ctx context.Context, ul d.UserList) ([]d.User, error) {
	return nil, errors.New("not implemented")
}
