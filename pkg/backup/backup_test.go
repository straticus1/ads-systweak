package backup

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeDefaults struct {
	value       string
	valueType   ValueType
	readErr     error
	writes      []writeCall
	deleteCalls int
}

type writeCall struct {
	domain    string
	key       string
	valueType ValueType
	value     string
}

func (f *fakeDefaults) Read(domain, key string) (string, ValueType, error) {
	return f.value, f.valueType, f.readErr
}

func (f *fakeDefaults) Write(domain, key string, valueType ValueType, value string) error {
	f.writes = append(f.writes, writeCall{domain, key, valueType, value})
	return nil
}

func (f *fakeDefaults) Delete(domain, key string) error {
	f.deleteCalls++
	return nil
}

func newTestStore(t *testing.T, client DefaultsClient) *Store {
	t.Helper()
	return &Store{
		Dir:    t.TempDir(),
		Client: client,
		Now:    func() time.Time { return time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC) },
	}
}

func TestSaveIfAbsentPreservesFirstValue(t *testing.T) {
	fake := &fakeDefaults{value: "1", valueType: TypeBool}
	store := newTestStore(t, fake)

	if err := store.SaveIfAbsent("com.example", "Enabled"); err != nil {
		t.Fatalf("first SaveIfAbsent: %v", err)
	}
	fake.value = "0"
	if err := store.SaveIfAbsent("com.example", "Enabled"); err != nil {
		t.Fatalf("second SaveIfAbsent: %v", err)
	}

	entry, err := store.Load("com.example", "Enabled")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if entry.Value != "1" || entry.Type != TypeBool {
		t.Fatalf("entry = %#v, want original boolean value", entry)
	}
	info, err := os.Stat(store.path("com.example", "Enabled"))
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("backup mode = %o, want 600", got)
	}
}

func TestSaveIfAbsentRecordsMissingKey(t *testing.T) {
	fake := &fakeDefaults{readErr: ErrNotFound}
	store := newTestStore(t, fake)

	if err := store.SaveIfAbsent("com.example", "Missing"); err != nil {
		t.Fatalf("SaveIfAbsent: %v", err)
	}
	entry, err := store.Load("com.example", "Missing")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if entry.Exists {
		t.Fatalf("entry.Exists = true, want false")
	}
}

func TestSaveIfAbsentRejectsUnsupportedType(t *testing.T) {
	fake := &fakeDefaults{value: "(one, two)", valueType: ValueType("array")}
	store := newTestStore(t, fake)

	err := store.SaveIfAbsent("com.example", "Items")
	if !errors.Is(err, ErrUnsupportedType) {
		t.Fatalf("error = %v, want ErrUnsupportedType", err)
	}
	if _, statErr := os.Stat(store.path("com.example", "Items")); !os.IsNotExist(statErr) {
		t.Fatalf("backup unexpectedly created: %v", statErr)
	}
}

func TestRestoreWritesOriginalTypeAndConsumesReceipt(t *testing.T) {
	fake := &fakeDefaults{value: "42", valueType: TypeInt}
	store := newTestStore(t, fake)
	if err := store.SaveIfAbsent("com.example", "Count"); err != nil {
		t.Fatalf("SaveIfAbsent: %v", err)
	}

	if err := store.Restore("com.example", "Count"); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(fake.writes) != 1 || fake.writes[0].valueType != TypeInt || fake.writes[0].value != "42" {
		t.Fatalf("writes = %#v, want typed integer restoration", fake.writes)
	}
	if _, err := store.Load("com.example", "Count"); !errors.Is(err, ErrNoBackup) {
		t.Fatalf("Load after restore error = %v, want ErrNoBackup", err)
	}
}

func TestRestoreDeletesOriginallyMissingKey(t *testing.T) {
	fake := &fakeDefaults{readErr: ErrNotFound}
	store := newTestStore(t, fake)
	if err := store.SaveIfAbsent("com.example", "Missing"); err != nil {
		t.Fatalf("SaveIfAbsent: %v", err)
	}

	if err := store.Restore("com.example", "Missing"); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if fake.deleteCalls != 1 {
		t.Fatalf("delete calls = %d, want 1", fake.deleteCalls)
	}
}

func TestRestoreKeepsReceiptWhenWriteFails(t *testing.T) {
	dir := t.TempDir()
	client := &failingWriteDefaults{fakeDefaults: fakeDefaults{value: "old", valueType: TypeString}}
	store := &Store{Dir: dir, Client: client, Now: time.Now}
	if err := store.SaveIfAbsent("com.example", "Name"); err != nil {
		t.Fatalf("SaveIfAbsent: %v", err)
	}

	if err := store.Restore("com.example", "Name"); err == nil {
		t.Fatal("Restore returned nil error")
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.Base(store.path("com.example", "Name")))); err != nil {
		t.Fatalf("receipt removed after failed restore: %v", err)
	}
}

type failingWriteDefaults struct{ fakeDefaults }

func (f *failingWriteDefaults) Write(domain, key string, valueType ValueType, value string) error {
	return errors.New("write failed")
}
