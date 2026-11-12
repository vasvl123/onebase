package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/ivantit66/onebase/internal/auth"
	"github.com/ivantit66/onebase/internal/dsl/interpreter"
	"github.com/ivantit66/onebase/internal/runtime"
	"github.com/ivantit66/onebase/internal/storage"
	"github.com/ivantit66/onebase/internal/ui"
)

// shutdownFixture собирает полноценный сервер (frontend + api), чтобы
// администраторская команда «Выключить сервер» проходила через публичный
// вход Handler(), а не через прямой вызов хендлера.
func shutdownFixture(t *testing.T) (*Server, *auth.Repo, context.Context) {
	t.Helper()
	ctx := context.Background()
	db, err := storage.ConnectSQLite(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	authRepo := auth.NewRepo(db)
	if err := authRepo.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	frontend := ui.New(runtime.NewRegistry(), db, interpreter.New(), authRepo, ui.Config{}, nil)
	srv := New(runtime.NewRegistry(), db, interpreter.New(), authRepo, "", 0, Config{}, frontend, nil)
	return srv, authRepo, ctx
}

func postAsUser(t *testing.T, srv *Server, authRepo *auth.Repo, ctx context.Context, login string, admin bool, path string) *httptest.ResponseRecorder {
	t.Helper()
	u, err := authRepo.CreateManaged(ctx, login, "secret12345", login, admin)
	if err != nil {
		t.Fatal(err)
	}
	token, err := authRepo.CreateSession(ctx, u.ID, auth.SessionMeta{Kind: auth.SessionKindEnterprise})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.AddCookie(&http.Cookie{Name: "onebase_session", Value: token})
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func TestAdminShutdown_AdminTriggersGracefulStop(t *testing.T) {
	srv, authRepo, ctx := shutdownFixture(t)
	rec := postAsUser(t, srv, authRepo, ctx, "root", true, "/ui/admin/shutdown")
	if rec.Code != http.StatusOK {
		t.Fatalf("admin shutdown: %d: %s", rec.Code, rec.Body.String())
	}
	select {
	case <-srv.Done():
	default:
		t.Fatal("аутентифицированная команда администратора не закрыла Done")
	}
}

func TestAdminShutdown_NonAdminForbidden(t *testing.T) {
	srv, authRepo, ctx := shutdownFixture(t)
	// Первый пользователь обязан быть администратором, затем заводим
	// неадминистратора и проверяем, что его команда заперта.
	if _, err := authRepo.CreateManaged(ctx, "root", "secret12345", "root", true); err != nil {
		t.Fatal(err)
	}
	rec := postAsUser(t, srv, authRepo, ctx, "kassir", false, "/ui/admin/shutdown")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("не-админ: ожидался 403, получено %d: %s", rec.Code, rec.Body.String())
	}
	select {
	case <-srv.Done():
		t.Fatal("запретил администратору — не должен был закрыть Done")
	default:
	}
}
