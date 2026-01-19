package setup

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	authStorage "main/internal/auth/storage"
	categoryDomain "main/internal/category/domain"
	categoryStorage "main/internal/category/storage"
	"main/internal/external/streamserver"
	streamservermock "main/internal/external/streamserver/mock"
	follow "main/internal/follow/storage"
	followStorage "main/internal/follow/storage"
	userStorage "main/internal/user/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Populate(ctx context.Context,
	pool *pgxpool.Pool,
	as *authStorage.ServiceImpl,
	ls *streamserver.Adapter,
	cr *categoryStorage.RepositoryImpl,
	fs *followStorage.RepositoryImpl,
	us *userStorage.RepositoryImpl,
	streamServerBaseUrl string) {
	addTags(ctx, pool)
	addCategories(ctx, cr)
	addUsers(ctx, pool)
	addFollows(ctx, fs)
	startLivestreams(cr, ls, streamServerBaseUrl)
}

func addCategories(ctx context.Context, cr *categoryStorage.RepositoryImpl) {
	slog.Info("adding categories")

	categoriesLen := min(categoriesCount, len(categories))
	categories = categories[:categoriesLen]

	for _, cat := range categories {
		err := cr.Create(ctx, categoryDomain.CategoryCreate{
			Thumbnail: cat.Thumbnail,
			Name:      cat.Name,
			Link:      cat.Link,
			Tags:      []int{1, 2},
		})

		if err != nil {
			log.Fatalf("unable to add category: %v", err)
		}
	}

	slog.Info("categories added")
}

func addUsers(ctx context.Context, pool *pgxpool.Pool) {
	slog.Info("adding users")

	r, err := pool.Query(ctx, usersToSQL(users))
	if err != nil {
		fmt.Println(usersToSQL(users))
		log.Fatalf("unable to add users: %v", err)
	}
	r.Close()

	slog.Info("users added")
}

func startLivestreams(cr *categoryStorage.RepositoryImpl, ls *streamserver.Adapter, baseUrl string) {
	slog.Info("starting livestreams")

	cl := streamservermock.NewStreamServerClient(baseUrl)

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

	r, err := pool.Query(ctx, tagsToSQL(tags))
	if err != nil {
		log.Fatalf("unable to add tags: %v", err)
	}
	r.Close()

	slog.Info("tags added")
}

func addFollows(ctx context.Context, fs *follow.RepositoryImpl) {
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
	sql.WriteString(`INSERT INTO tc_user(name, password, pfp, description, links, tags, id_category) VALUES `)

	for _, user := range users {
		str := fmt.Sprintf("('%s', '%s', '%s', '%s', '{instagram, telegram}', '{tag1, tag2}', '%d'),", user.Name, user.Password, user.Avatar, user.Description, 3)
		sql.WriteString(str)
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
