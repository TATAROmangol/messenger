package transport

import (
	"testing"
	"net/http/httptest"
	"auth_service/internal/config"
	"auth_service/pkg/postgres"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func TestRegister(t *testing.T) {
	cfg, err := config.New("../../config/config_test.env")
	if err != nil {
		t.Error("Error when loading config: ", err)
		return
	}

	pool, err := postgres.New(context.Background(), (*cfg).Postgres, "../../db/migrations")
	if err != nil {
		t.Error("Error when executing postgres.New(): ", err)
		return
	}
	defer pool.Close()

	r := httptest.NewRequest(http.MethodGet, "/auth/register", bytes.NewBuffer([]byte(`{"login":"testlogin","email":"testemail","pass":"testpass","name":"testname"}`)))
    w := httptest.NewRecorder()
    Register(w, r)
	defer pool.Exec(context.Background(), "DELETE FROM auth_schema.users WHERE login='testlogin';")
    
	res := w.Result()
    defer res.Body.Close()

	var data RegisterResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Error("Error when decoding response: ", err)
		return
	}

	type Record struct {
		Login string
		Email string
		Name string
	}
	var dbdata Record
	err = pool.QueryRow(context.Background(), "SELECT login, email, name FROM auth_schema.users WHERE token=$1;", data.Token).Scan(&dbdata.Login, &dbdata.Email, &dbdata.Name)
	if err != nil {
		t.Error("Error when selecting data from the database: ", err)
		return
	}
	if dbdata.Login != "testlogin" || dbdata.Email != "testemail" || dbdata.Name != "testname" {
		t.Error("Invalid login in the database. Expected testlogin testemail, testname, found:", dbdata.Login, dbdata.Email, dbdata.Name)
		return
	}
}

func TestLogin(t *testing.T) {
	cfg, err := config.New("../../config/config_test.env")
	if err != nil {
		t.Error("Error when loading config: ", err)
		return
	}

	pool, err := postgres.New(context.Background(), (*cfg).Postgres, "../../db/migrations")
	if err != nil {
		t.Error("Error when executing postgres.New(): ", err)
		return
	}
	defer pool.Close()

	id := 0
	err = pool.QueryRow(context.Background(), "INSERT INTO auth_schema.users (token, login, email, pass, name) VALUES ('testtoken', 'testlogin', 'testemail', '179ad45c6ce2cb97cf1029e212046e81', 'testname') RETURNING id;").Scan(&id)
	if err != nil {
		t.Error("Error when inserting user: ", err)
		return
	}
	defer pool.Exec(context.Background(), "DELETE FROM auth_schema.users WHERE id=$1;", id)
	

	r := httptest.NewRequest(http.MethodGet, "/auth/login", bytes.NewBuffer([]byte(`{"credential":"testlogin","pass":"testpass"}`)))
    w := httptest.NewRecorder()
    Login(w, r)
	res := w.Result()
    defer res.Body.Close()

	var data LoginResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Error("Error when decoding response: ", err)
		return
	}

	type Record struct {
		Login string
		Email string
		Name string
	}
	var dbdata Record
	err = pool.QueryRow(context.Background(), "SELECT login, email, name FROM auth_schema.users WHERE (token=$1 AND id=$2);", data.Token, id).Scan(&dbdata.Login, &dbdata.Email, &dbdata.Name)
	if err != nil {
		t.Error("Error when selecting data from the database: ", err)
		return
	}
	if dbdata.Login != "testlogin" || dbdata.Email != "testemail" || dbdata.Name != "testname" {
		t.Error("Invalid login in the database. Expected testlogin testemail, testname, found:", dbdata.Login, dbdata.Email, dbdata.Name)
		return
	}
}

func TestLogout(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/auth/logout", nil)
	r.AddCookie(&http.Cookie{
		Name:    JWTCookieName,
		Value:   "testtoken",
		Expires: time.Now().Add(1*time.Hour),
	})
    w := httptest.NewRecorder()
    Logout(w, r)
	res := w.Result()
    defer res.Body.Close()

	cookies := res.Cookies()
	if len(cookies) > 0 {
		if cookies[0].Value != "" {
			t.Error("Cookie isn't deleted. Value: \"" + cookies[0].Value + "\"")
			return
		}
	}
}

func TestDelete(t *testing.T) {
	cfg, err := config.New("../../config/config_test.env")
	if err != nil {
		t.Error("Error when loading config: ", err)
		return
	}

	pool, err := postgres.New(context.Background(), (*cfg).Postgres, "../../db/migrations")
	if err != nil {
		t.Error("Error when executing postgres.New(): ", err)
		return
	}
	defer pool.Close()

	id := 0
	err = pool.QueryRow(context.Background(), "INSERT INTO auth_schema.users (token, login, email, pass, name) VALUES ('testtoken', 'testlogin', 'testemail', '179ad45c6ce2cb97cf1029e212046e81', 'testname') RETURNING id;").Scan(&id)
	if err != nil {
		t.Error("Error when inserting user: ", err)
		return
	}

	r := httptest.NewRequest(http.MethodGet, "/auth/delete", nil)
	r.AddCookie(&http.Cookie{
		Name:    JWTCookieName,
		Value:   "testtoken",
		Expires: time.Now().Add(1*time.Hour),
	})
    w := httptest.NewRecorder()
    Delete(w, r)
	res := w.Result()
    defer res.Body.Close()

	cookies := res.Cookies()
	if len(cookies) > 0 {
		if cookies[0].Value != "" {
			t.Error("Cookie isn't deleted. Value: \"" + cookies[0].Value + "\"")
			return
		}
	}

	dblogin := ""
	err = pool.QueryRow(context.Background(), "SELECT login FROM auth_schema.users WHERE id=$1;", id).Scan(&dblogin)
	if err == nil || dblogin != "" {
		t.Error("Expected error (db record should not be found after deleting), found: ", dblogin)
		return
	}
}