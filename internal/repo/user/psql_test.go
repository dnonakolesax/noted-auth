package user

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dbsql "github.com/dnonakolesax/noted-auth/internal/db/sql"

	"github.com/dnonakolesax/noted-auth/internal/mocks"
)

/* ----------------------------- test helpers ----------------------------- */

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

// newPGXResponse injects pgx.Rows into dbsql.PGXResponse despite unexported field `rows`.
func newPGXResponse(rows pgx.Rows) *dbsql.PGXResponse {
	pr := &dbsql.PGXResponse{}
	v := reflect.ValueOf(pr).Elem().FieldByName("rows")
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(rows))
	return pr
}

// setPtr sets a pointer destination to a value using reflection.
// If `value` can't be assigned/converted, it falls back to a deterministic value by kind.
func setPtr(ptr any, value any) any {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return nil
	}

	ev := rv.Elem()

	if value != nil {
		vv := reflect.ValueOf(value)
		if vv.IsValid() {
			if vv.Type().AssignableTo(ev.Type()) {
				ev.Set(vv)
				return ev.Interface()
			}
			if vv.Type().ConvertibleTo(ev.Type()) {
				ev.Set(vv.Convert(ev.Type()))
				return ev.Interface()
			}
		}
	}

	switch ev.Kind() {
	case reflect.String:
		ev.SetString("test-value")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ev.SetInt(123)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		ev.SetUint(123)
	case reflect.Bool:
		ev.SetBool(true)
	default:
		// leave zero
	}
	return ev.Interface()
}

/* ----------------------------- pgx.Rows stub ----------------------------- */

// rowsStub implements pgx.Rows interface.
type rowsStub struct {
	nextSeq []bool
	nextIdx int

	scanFn func(dest ...any) error
	err    error

	closed bool
}

func (r *rowsStub) Close() { r.closed = true }
func (r *rowsStub) Err() error {
	return r.err
}
func (r *rowsStub) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }
func (r *rowsStub) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}
func (r *rowsStub) Next() bool {
	if r.nextIdx >= len(r.nextSeq) {
		return false
	}
	v := r.nextSeq[r.nextIdx]
	r.nextIdx++
	return v
}
func (r *rowsStub) Scan(dest ...any) error {
	if r.scanFn == nil {
		return nil
	}
	return r.scanFn(dest...)
}
func (r *rowsStub) Values() ([]any, error) { return nil, nil }
func (r *rowsStub) RawValues() [][]byte    { return nil }
func (r *rowsStub) Conn() *pgx.Conn        { return nil }

/* ----------------------------- tests: NewUserRepo ----------------------------- */

func TestNewUserRepo_OK_LoadsSQL(t *testing.T) {
	tmp := t.TempDir()

	// NewUserRepo вызывает LoadSQLRequests(requestsPath + "user")
	// => делаем <tmp>/req/user/*.sql
	reqRoot := filepath.Join(tmp, "req")
	userDir := filepath.Join(reqRoot, thisDomainName)
	require.NoError(t, os.MkdirAll(userDir, 0o755))

	getUserSQL := "select login, first_name, last_name from users where id=$1 and realm_id=$2;"
	getUserByNameSQL := "select id from users where login=$1 and realm_id=$2;"

	require.NoError(t, os.WriteFile(filepath.Join(userDir, getUserFileName+".sql"), []byte(getUserSQL), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(userDir, getUserByNameFileName+".sql"),
		[]byte(getUserByNameSQL), 0o644))

	mw := mocks.NewMockIPGXWorker(t)

	repo, err := NewUserRepo(mw, "realm-1", reqRoot+string(os.PathSeparator), testLogger())
	require.NoError(t, err)
	require.NotNil(t, repo)

	require.Equal(t, getUserSQL, repo.requests[getUserFileName])
	require.Equal(t, getUserByNameSQL, repo.requests[getUserByNameFileName])
}

func TestNewUserRepo_ErrorOnMissingDir(t *testing.T) {
	mw := mocks.NewMockIPGXWorker(t)

	_, err := NewUserRepo(mw, "realm-1", filepath.Join(t.TempDir(), "nope")+string(os.PathSeparator), testLogger())
	require.Error(t, err)
}

/* ----------------------------- tests: GetUser ----------------------------- */

func TestUserRepo_GetUser_QueryError(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserFileName: "SQL_GET_USER"},
	}

	qErr := errors.New("db down")

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER", "u1", "realm").
		Return((*dbsql.PGXResponse)(nil), qErr).
		Once()

	_, err := ur.GetUser(ctx, "u1")
	require.ErrorIs(t, err, qErr)
}

func TestUserRepo_GetUser_NotFound(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserFileName: "SQL_GET_USER"},
	}

	rows := &rowsStub{nextSeq: []bool{false}}
	resp := newPGXResponse(rows)

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER", "u1", "realm").
		Return(resp, nil).
		Once()

	_, err := ur.GetUser(ctx, "u1")
	require.Error(t, err)
	require.Equal(t, "not found", err.Error())
}

func TestUserRepo_GetUser_ScanError(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserFileName: "SQL_GET_USER"},
	}

	scErr := errors.New("scan failed")

	rows := &rowsStub{
		nextSeq: []bool{true},
		scanFn:  func(_ ...any) error { return scErr },
	}
	resp := newPGXResponse(rows)

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER", "u1", "realm").
		Return(resp, nil).
		Once()

	_, err := ur.GetUser(ctx, "u1")
	require.Error(t, err)
	require.ErrorIs(t, err, scErr) // PGXResponse.Scan wraps with %w
}

func TestUserRepo_GetUser_TooManyRows(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserFileName: "SQL_GET_USER"},
	}

	rows := &rowsStub{
		nextSeq: []bool{true, true}, // второй Next() => too many rows
		scanFn: func(dest ...any) error {
			_ = setPtr(dest[0], "bob")
			_ = setPtr(dest[1], "Bob")
			_ = setPtr(dest[2], "Builder")
			return nil
		},
	}
	resp := newPGXResponse(rows)

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER", "u1", "realm").
		Return(resp, nil).
		Once()

	_, err := ur.GetUser(ctx, "u1")
	require.Error(t, err)
	require.Equal(t, "too many rows", err.Error())
}

func TestUserRepo_GetUser_CloseError(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserFileName: "SQL_GET_USER"},
	}

	closeErr := errors.New("rows err after close")

	rows := &rowsStub{
		nextSeq: []bool{true, false},
		scanFn: func(dest ...any) error {
			_ = setPtr(dest[0], "bob")
			_ = setPtr(dest[1], "Bob")
			_ = setPtr(dest[2], "Builder")
			return nil
		},
		err: closeErr, // вернётся из rows.Err() внутри PGXResponse.Close()
	}
	resp := newPGXResponse(rows)

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER", "u1", "realm").
		Return(resp, nil).
		Once()

	_, err := ur.GetUser(ctx, "u1")
	require.ErrorIs(t, err, closeErr)
}

func TestUserRepo_GetUser_OK(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserFileName: "SQL_GET_USER"},
	}

	rows := &rowsStub{
		nextSeq: []bool{true, false},
		scanFn: func(dest ...any) error {
			_ = setPtr(dest[0], "bob")
			_ = setPtr(dest[1], "Bob")
			_ = setPtr(dest[2], "Builder")
			return nil
		},
	}
	resp := newPGXResponse(rows)

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER", "u1", "realm").
		Return(resp, nil).
		Once()

	got, err := ur.GetUser(ctx, "u1")
	require.NoError(t, err)

	require.Equal(t, "bob", got.Login)
	require.Equal(t, "Bob", got.FirstName)
	require.Equal(t, "Builder", got.LastName)
}

/* ----------------------------- tests: IDByName ----------------------------- */

func TestUserRepo_IDByName_QueryError(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserByNameFileName: "SQL_GET_USER_BY_NAME"},
	}

	qErr := errors.New("db down")

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER_BY_NAME", "alice", "realm").
		Return((*dbsql.PGXResponse)(nil), qErr).
		Once()

	_, err := ur.IDByName(ctx, "alice")
	require.ErrorIs(t, err, qErr)
}

func TestUserRepo_IDByName_NotFound(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserByNameFileName: "SQL_GET_USER_BY_NAME"},
	}

	rows := &rowsStub{nextSeq: []bool{false}}
	resp := newPGXResponse(rows)

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER_BY_NAME", "alice", "realm").
		Return(resp, nil).
		Once()

	_, err := ur.IDByName(ctx, "alice")
	require.Error(t, err)
	require.Equal(t, "not found", err.Error())
}

func TestUserRepo_IDByName_ScanError(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserByNameFileName: "SQL_GET_USER_BY_NAME"},
	}

	scErr := errors.New("scan failed")

	rows := &rowsStub{
		nextSeq: []bool{true},
		scanFn:  func(_ ...any) error { return scErr },
	}
	resp := newPGXResponse(rows)

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER_BY_NAME", "alice", "realm").
		Return(resp, nil).
		Once()

	_, err := ur.IDByName(ctx, "alice")
	require.Error(t, err)
	require.ErrorIs(t, err, scErr) // PGXResponse.Scan wraps with %w
}

func TestUserRepo_IDByName_TooManyRows(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserByNameFileName: "SQL_GET_USER_BY_NAME"},
	}

	rows := &rowsStub{
		nextSeq: []bool{true, true},
		scanFn: func(dest ...any) error {
			_ = setPtr(dest[0], nil)
			return nil
		},
	}
	resp := newPGXResponse(rows)

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER_BY_NAME", "alice", "realm").
		Return(resp, nil).
		Once()

	_, err := ur.IDByName(ctx, "alice")
	require.Error(t, err)
	require.Equal(t, "too many rows", err.Error())
}

func TestUserRepo_IDByName_CloseError(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserByNameFileName: "SQL_GET_USER_BY_NAME"},
	}

	closeErr := errors.New("rows err after close")

	rows := &rowsStub{
		nextSeq: []bool{true, false},
		scanFn: func(dest ...any) error {
			_ = setPtr(dest[0], nil)
			return nil
		},
		err: closeErr,
	}
	resp := newPGXResponse(rows)

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER_BY_NAME", "alice", "realm").
		Return(resp, nil).
		Once()

	_, err := ur.IDByName(ctx, "alice")
	require.ErrorIs(t, err, closeErr)
}

func TestUserRepo_IDByName_OK(t *testing.T) {
	ctx := context.Background()

	mw := mocks.NewMockIPGXWorker(t)
	ur := &UserRepo{
		worker:   mw,
		realmID:  "realm",
		logger:   testLogger(),
		requests: map[string]string{getUserByNameFileName: "SQL_GET_USER_BY_NAME"},
	}

	var expected any

	rows := &rowsStub{
		nextSeq: []bool{true, false},
		scanFn: func(dest ...any) error {
			// выставляем значение по фактическому типу user.ID
			expected = setPtr(dest[0], "u-1")
			return nil
		},
	}
	resp := newPGXResponse(rows)

	mw.EXPECT().
		Query(mock.Anything, "SQL_GET_USER_BY_NAME", "alice", "realm").
		Return(resp, nil).
		Once()

	got, err := ur.IDByName(ctx, "alice")
	require.NoError(t, err)

	require.Equal(t, fmt.Sprint(expected), got.ID)
}
