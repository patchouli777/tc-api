package setup

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"main/internal/endpoint/auth"
	"main/internal/endpoint/category"
	"main/internal/endpoint/follow"
	"main/internal/endpoint/livestream"
	"main/internal/endpoint/user"
	livestreamApi "main/pkg/api/model/livestream"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Populate(ctx context.Context,
	pool *pgxpool.Pool,
	as *auth.ServiceImpl,
	ls *livestream.StreamServerAdapter,
	cr *category.RepositoryImpl,
	fs *follow.ServiceImpl,
	us *user.ServiceImpl) {
	addTags(ctx, pool)
	addUsers(ctx, pool)
	addCategories(ctx, cr)
	addFollows(ctx, fs)
	startLivestreams(ctx, cr, ls)
}

func addCategories(ctx context.Context, cr *category.RepositoryImpl) {
	slog.Info("adding categories")

	categoriesLen := min(categoriesCount, len(categories))
	categories = categories[:categoriesLen]

	for _, cat := range categories {
		err := cr.Create(ctx, category.CategoryCreate{
			Thumbnail: cat.Thumbnail,
			Name:      cat.Name,
			Link:      cat.Link,
			Tags:      []int32{1, 2},
		})

		if err != nil {
			log.Fatalf("unable to add category: %v", err)
		}
	}

	slog.Info("categories added")
}

func addUsers(ctx context.Context, pool *pgxpool.Pool) {
	slog.Info("adding users")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("unable to acquire connection: %v", err)
		return
	}

	r, err := conn.Query(ctx, usersToSQL(users))
	if err != nil {
		fmt.Println(usersToSQL(users))
		log.Fatalf("unable to add users: %v", err)
	}
	r.Close()

	slog.Info("users added")
}

func startLivestreams(ctx context.Context, cr *category.RepositoryImpl, ls *livestream.StreamServerAdapter) {
	slog.Info("starting livestreams")

	cl := livestreamApi.NewStreamServerClient()

	for i := range streamsCount {
		_, err := cl.Start(users[i].Name)
		if err != nil {
			log.Fatalf("unable to start livestream: %v", err)
		}
	}

	slog.Info("livestreams started")
}

func addTags(ctx context.Context, pool *pgxpool.Pool) {
	slog.Info("adding tags")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("unable to acquire connection: %v", err)
		return
	}

	r, err := conn.Query(ctx, tagsToSQL(tags))
	if err != nil {
		log.Fatalf("unable to add tags: %v", err)
	}
	r.Close()

	slog.Info("tags added")
}

func addFollows(ctx context.Context, fs *follow.ServiceImpl) {
	slog.Info("adding follows")

	for i := range followCount {
		err := fs.Follow(ctx, "user1", fmt.Sprintf("user%d", i))
		if err != nil {
			log.Fatalf("unable to follow a user: %v", err)
		}
	}

	slog.Info("follows added")
}

func usersToSQL(users []setupUser) string {
	var sql bytes.Buffer
	sql.WriteString(`INSERT INTO tc_user(name, password, avatar, description, links, tags) VALUES `)

	for _, user := range users {
		sql.WriteString(fmt.Sprintf("('%s', '%s', '%s', '%s', '{instagram, telegram}', '{tag1, tag2}'),",
			user.Name, user.Password, user.Avatar, user.Description))
	}
	sql.Truncate(sql.Len() - 1)
	sql.WriteString(";")
	return sql.String()
}

func tagsToSQL(tags []string) string {
	var sql bytes.Buffer
	sql.WriteString(`INSERT INTO tc_tag(name) VALUES`)

	for _, t := range tags {
		sql.WriteString(fmt.Sprintf("('%s'),", t))
	}
	sql.Truncate(sql.Len() - 1)
	sql.WriteString(";")
	return sql.String()
}
