package dataset

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const csvContent = "step,type,amount\n1,PAYMENT,9839.64\n"

// zipWith arma en memoria un .zip como el que entrega Kaggle.
func zipWith(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func serve(t *testing.T, status int, body []byte) (*httptest.Server, *int) {
	t.Helper()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(status)
		w.Write(body)
	}))
	t.Cleanup(server.Close)
	return server, &calls
}

func TestDownload_ExtractsCSVAndRemovesZip(t *testing.T) {
	body := zipWith(t, map[string]string{FileName: csvContent, "LEEME.txt": "x"})
	server, _ := serve(t, http.StatusOK, body)
	dir := filepath.Join(t.TempDir(), "raw") // la carpeta aun no existe

	path, err := Download(server.URL, dir, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, FileName) {
		t.Fatalf("path = %s", path)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != csvContent {
		t.Fatalf("contenido = %q, %v", got, err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("solo debe quedar el CSV, quedaron %d archivos", len(entries))
	}
}

func TestDownload_SkipsWhenCSVAlreadyExists(t *testing.T) {
	server, calls := serve(t, http.StatusOK, nil)
	dir := t.TempDir()
	existing := filepath.Join(dir, FileName)
	if err := os.WriteFile(existing, []byte(csvContent), 0o644); err != nil {
		t.Fatal(err)
	}
	path, err := Download(server.URL, dir, io.Discard)
	if err != nil || path != existing {
		t.Fatalf("Download = %s, %v", path, err)
	}
	if *calls != 0 {
		t.Fatal("no debe volver a descargar si el CSV ya existe")
	}
}

func TestDownload_Errors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   []byte
	}{
		{"http error", http.StatusNotFound, []byte("not found")},
		{"not a zip", http.StatusOK, []byte("esto no es un zip")},
		{"zip without csv", http.StatusOK, zipWith(t, map[string]string{"otro.txt": "x"})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server, _ := serve(t, tc.status, tc.body)
			dir := t.TempDir()
			if _, err := Download(server.URL, dir, io.Discard); err == nil {
				t.Fatal("expected an error")
			}
			// Un fallo no debe dejar un CSV a medias que luego parezca valido.
			if _, err := os.Stat(filepath.Join(dir, FileName)); !os.IsNotExist(err) {
				t.Fatal("no debe quedar el CSV tras un error")
			}
		})
	}
}
