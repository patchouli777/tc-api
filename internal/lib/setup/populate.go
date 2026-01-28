package setup

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"
	authStorage "twitchy-api/internal/auth/storage"
	categoryDomain "twitchy-api/internal/category/domain"
	categoryStorage "twitchy-api/internal/category/storage"
	"twitchy-api/internal/external/streamserver"
	streamservermock "twitchy-api/internal/external/streamserver/mock"
	followStorage "twitchy-api/internal/follow/storage"
	userStorage "twitchy-api/internal/user/storage"

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
	slog.Info("adding tags")
	t := withTime(func() {
		AddTags(ctx, pool)
	})
	slog.Info("tags added", slog.Duration("took", t))

	slog.Info("adding categories")
	t = withTime(func() {
		AddCategories(ctx, cr)
	})
	slog.Info("categories added", slog.Duration("took", t))

	slog.Info("adding users")
	t = withTime(func() {
		addUsers(ctx, pool)
	})
	slog.Info("users added", slog.Duration("took", t))

	slog.Info("adding follows")
	t = withTime(func() {
		AddFollows(ctx, fs)
	})
	slog.Info("follows added", slog.Duration("took", t))

	slog.Info("starting livestreams")
	t = withTime(func() {
		StartLivestreams(ls, streamServerBaseUrl)
	})
	slog.Info("livestreams started", slog.Duration("took", t))
}

func withTime(f func()) time.Duration {
	now := time.Now()
	f()
	return time.Since(now)
}

func AddCategories(ctx context.Context, cr *categoryStorage.RepositoryImpl) {
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
}

func addUsers(ctx context.Context, pool *pgxpool.Pool) {
	r, err := pool.Query(ctx, usersToSQL(users))
	if err != nil {
		fmt.Println(usersToSQL(users))
		log.Fatalf("unable to add users: %v", err)
	}
	r.Close()
}

func StartLivestreams(ls *streamserver.Adapter, baseUrl string) {
	cl := streamservermock.NewStreamServerClient(baseUrl)

	for i := range streamsCount {
		_, err := cl.Start(users[i].Name)
		if err != nil {
			log.Fatalf("unable to start livestream: %v", err)
		}
	}
}

func AddTags(ctx context.Context, pool *pgxpool.Pool) {
	r, err := pool.Query(ctx, tagsToSQL(tags))
	if err != nil {
		log.Fatalf("unable to add tags: %v", err)
	}
	r.Close()
}

// TODO: sql instead of repo
func AddFollows(ctx context.Context, fs *followStorage.RepositoryImpl) {
	for i := range followCount {
		err := fs.Follow(ctx, "user1", fmt.Sprintf("user%d", i))
		if err != nil {
			log.Fatalf("unable to follow a user: %v", err)
		}
	}
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
