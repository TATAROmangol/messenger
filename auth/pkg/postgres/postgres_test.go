package postgres

import (
	"context"
	"testing"
	"auth_service/internal/config"
)

func TestNew(t *testing.T) {
	cfg, err := config.New("../../config/config_test.env")
	if err != nil {
		t.Error("Error when loading config: ", err)
		return
	}

	_, err = New(context.Background(), (*cfg).Postgres, "../../db/migrations")
	if err != nil {
		t.Error("Error when executing New(): ", err)
		return
	}
	defer PGXPool.Close()

	err = PGXPool.Ping(context.Background())
	if err != nil {
		t.Error("Error when executing Ping(): ", err)
		return
	}
}

func TestInsertUser(t *testing.T) {
	cfg, err := config.New("../../config/config_test.env")
	if err != nil {
		t.Error("Error when loading config: ", err)
		return
	}

	_, err = New(context.Background(), (*cfg).Postgres, "../../db/migrations")
	if err != nil {
		t.Error("Error when executing New(): ", err)
		return
	}
	defer PGXPool.Close()

	id, err := InsertUser("testtoken", "testlogin", "testemail", "testpass", "testname")
	if err != nil {
		t.Error("Error when executing InsertUser(): ", err)
		return
	}
	defer PGXPool.Exec(context.Background(), "DELETE FROM auth_schema.users WHERE id=$1;", id)

	type Record struct {
		Token	string
		Login	string
		Email	string
		Pass	string
		Name	string
    }
	var ret Record;
	err = PGXPool.QueryRow(context.Background(), "SELECT token, login, email, pass, name FROM auth_schema.users WHERE id=$1;", id).Scan(&ret.Token, &ret.Login, &ret.Email, &ret.Pass, &ret.Name)
	if err != nil {
		t.Error("Error when executing QueryRow(): ", err)
		return
	}

	if ret.Token != "testtoken" || ret.Login != "testlogin" || ret.Email != "testemail" || ret.Pass != "testpass" || ret.Name != "testname" {
		t.Error("Unexpected record in the database:", ret.Token, ret.Login, ret.Email, ret.Pass, ret.Name)
		return
	}
}

func TestGetIdByToken(t *testing.T) {
	cfg, err := config.New("../../config/config_test.env")
	if err != nil {
		t.Error("Error when loading config: ", err)
		return
	}

	_, err = New(context.Background(), (*cfg).Postgres, "../../db/migrations")
	if err != nil {
		t.Error("Error when executing New(): ", err)
		return
	}
	defer PGXPool.Close()

	id := 0
	err = PGXPool.QueryRow(context.Background(), "INSERT INTO auth_schema.users (token, login, email, pass, name) VALUES ('testtoken', 'testlogin', 'testemail', 'testpass', 'testname') RETURNING id;").Scan(&id)
	if err != nil {
		t.Error("Error when inserting user: ", err)
		return
	}
	defer PGXPool.Exec(context.Background(), "DELETE FROM auth_schema.users WHERE id=$1;", id)

	dbid, err := GetIdByToken("testtoken")
	if err != nil {
		t.Error("Error when executing GetIdByToken(): ", err)
		return
	}

	if dbid != id {
		t.Error("Expected:", id, "got", dbid)
		return
	}
}

func TestGetTokenByCredentialAndPassword(t *testing.T) {
	cfg, err := config.New("../../config/config_test.env")
	if err != nil {
		t.Error("Error when loading config: ", err)
		return
	}

	_, err = New(context.Background(), (*cfg).Postgres, "../../db/migrations")
	if err != nil {
		t.Error("Error when executing New(): ", err)
		return
	}
	defer PGXPool.Close()

	_, err = PGXPool.Exec(context.Background(), "INSERT INTO auth_schema.users (token, login, email, pass, name) VALUES ('testtoken', 'testlogin', 'testemail', 'testpass', 'testname');")
	if err != nil {
		t.Error("Error when inserting user: ", err)
		return
	}
	defer PGXPool.Exec(context.Background(), "DELETE FROM auth_schema.users WHERE token='testtoken';")

	dbtoken, err := GetTokenByCredentialAndPassword("testlogin", "testpass")
	if err != nil {
		t.Error("Error when executing GetTokenByCredentialAndPassword(): ", err)
		return
	}

	if dbtoken != "testtoken" {
		t.Error("Expected: testtoken, got", dbtoken)
		return
	}
}

func TestDeleteUserByToken(t *testing.T) {
	cfg, err := config.New("../../config/config_test.env")
	if err != nil {
		t.Error("Error when loading config: ", err)
		return
	}

	_, err = New(context.Background(), (*cfg).Postgres, "../../db/migrations")
	if err != nil {
		t.Error("Error when executing New(): ", err)
		return
	}
	defer PGXPool.Close()

	id := 0
	err = PGXPool.QueryRow(context.Background(), "INSERT INTO auth_schema.users (token, login, email, pass, name) VALUES ('testtoken', 'testlogin', 'testemail', 'testpass', 'testname') RETURNING id;").Scan(&id)
	if err != nil {
		t.Error("Error when inserting user: ", err)
		return
	}

	err = DeleteUserByToken("testtoken")
	if err != nil {
		t.Error("Error when executing DeleteUserByToken(): ", err)
		return
	}

	dbid := 0
	if err = PGXPool.QueryRow(context.Background(), "SELECT id FROM auth_schema.users WHERE token='testtoken';").Scan(&dbid); err == nil || dbid == id {
		t.Error("The record hasn't deleted")
		PGXPool.Exec(context.Background(), "DELETE FROM auth_schema.users WHERE id=$1;", id)
		return
	}
}

func TestGetRecordByToken(t *testing.T) {
	cfg, err := config.New("../../config/config_test.env")
	if err != nil {
		t.Error("Error when loading config: ", err)
		return
	}

	_, err = New(context.Background(), (*cfg).Postgres, "../../db/migrations")
	if err != nil {
		t.Error("Error when executing New(): ", err)
		return
	}
	defer PGXPool.Close()

	id := 0
	err = PGXPool.QueryRow(context.Background(), "INSERT INTO auth_schema.users (token, login, email, pass, name) VALUES ('testtoken', 'testlogin', 'testemail', 'testpass', 'testname') RETURNING id;").Scan(&id)
	if err != nil {
		t.Error("Error when inserting user: ", err)
		return
	}
	defer PGXPool.Exec(context.Background(), "DELETE FROM auth_schema.users WHERE id=$1;", id)

	dblogin, dbemail, dbname, err := GetRecordByToken("testtoken")
	if err != nil {
		t.Error("Error when executing GetRecorddByToken(): ", err)
		return
	}

	if dblogin != "testlogin" || dbemail != "testemail" || dbname != "testname" {
		t.Error("Expected: dblogin,dbemail dbname, got", dblogin, dbemail, dbname)
		return
	}
}