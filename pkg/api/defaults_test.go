package api

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestWriteKeyRejectsNonNumericIntegerWithoutExecuting(t *testing.T) {
	oldRun := runCommand
	t.Cleanup(func() { runCommand = oldRun })
	called := false
	runCommand = func(string, ...string) (string, error) { called = true; return "", nil }
	body := []byte(`{"domain":"com.example","key":"Count","type":"int","value":"1; touch /tmp/bad"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/defaults/key", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handleWriteKey(response, req)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
	if called {
		t.Fatal("command executed for invalid integer")
	}
}

func TestWriteKeyPassesStringAsLiteralArgument(t *testing.T) {
	oldRun := runCommand
	oldBackup := saveDefaultsBackup
	t.Cleanup(func() { runCommand = oldRun; saveDefaultsBackup = oldBackup })
	var gotName string
	var gotArgs []string
	runCommand = func(name string, args ...string) (string, error) {
		gotName, gotArgs = name, append([]string(nil), args...)
		return "", nil
	}
	saveDefaultsBackup = func(string, string) error { return nil }
	body := []byte(`{"domain":"com.example","key":"Greeting","type":"string","value":"hello'; touch /tmp/bad"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/defaults/key", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handleWriteKey(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	want := []string{"write", "com.example", "Greeting", "-string", "hello'; touch /tmp/bad"}
	if gotName != "/usr/bin/defaults" || !reflect.DeepEqual(gotArgs, want) {
		t.Fatalf("command = %q %#v, want direct args %#v", gotName, gotArgs, want)
	}
}

func TestWriteKeyRequiresJSONContentType(t *testing.T) {
	response := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/defaults/key", bytes.NewBufferString(`{}`))
	handleWriteKey(response, req)
	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want 415", response.Code)
	}
}

func TestWriteKeyStopsWhenBackupFails(t *testing.T) {
	oldRun := runCommand
	oldBackup := saveDefaultsBackup
	t.Cleanup(func() { runCommand = oldRun; saveDefaultsBackup = oldBackup })
	called := false
	runCommand = func(string, ...string) (string, error) { called = true; return "", nil }
	saveDefaultsBackup = func(string, string) error { return errors.New("backup failed") }
	body := []byte(`{"domain":"com.example","key":"Enabled","type":"bool","value":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/defaults/key", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handleWriteKey(response, req)
	if response.Code != http.StatusInternalServerError || called {
		t.Fatalf("status = %d, command called = %v", response.Code, called)
	}
}
