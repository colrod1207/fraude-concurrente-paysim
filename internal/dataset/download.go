// Package dataset descarga el CSV crudo de PaySim desde Kaggle.
package dataset

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// KaggleURL es la API publica de descarga de Kaggle para PaySim. El dataset
// es publico, asi que no hace falta cuenta ni token.
const KaggleURL = "https://www.kaggle.com/api/v1/datasets/download/ealaxi/paysim1"

// FileName es el CSV dentro del .zip; es la ruta por defecto que usa main.go.
const FileName = "PS_20174392719_1491204439457_log.csv"

// Download baja el .zip desde url, extrae FileName en dir y borra el .zip.
// Si el CSV ya existe no descarga nada. Escribe el avance en progress.
// Los archivos se escriben con sufijo .part y se renombran al final, para
// que una descarga interrumpida nunca deje un CSV incompleto.
func Download(url, dir string, progress io.Writer) (string, error) {
	csvPath := filepath.Join(dir, FileName)
	if _, err := os.Stat(csvPath); err == nil {
		fmt.Fprintf(progress, "El dataset ya existe: %s\n", csvPath)
		return csvPath, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	zipPath := filepath.Join(dir, "paysim1.zip.part")
	defer os.Remove(zipPath)
	if err := fetch(url, zipPath, progress); err != nil {
		return "", err
	}
	fmt.Fprintln(progress, "Descomprimiendo...")
	if err := extract(zipPath, FileName, csvPath); err != nil {
		return "", err
	}
	return csvPath, nil
}

func fetch(url, path string, progress io.Writer) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("descargar dataset: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("descargar dataset: Kaggle respondio %s", resp.Status)
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	counter := &progressWriter{out: progress, total: resp.ContentLength}
	_, copyErr := io.Copy(out, io.TeeReader(resp.Body, counter))
	closeErr := out.Close()
	fmt.Fprintln(progress)
	if copyErr != nil {
		return fmt.Errorf("descargar dataset: %w", copyErr)
	}
	return closeErr
}

func extract(zipPath, name, dest string) error {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("abrir zip: %w", err)
	}
	defer archive.Close()
	for _, file := range archive.File {
		if file.Name != name {
			continue
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()
		tmp := dest + ".part"
		out, err := os.Create(tmp)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, src)
		closeErr := out.Close()
		if copyErr != nil || closeErr != nil {
			os.Remove(tmp)
			return fmt.Errorf("extraer %s: %w", name, errors.Join(copyErr, closeErr))
		}
		return os.Rename(tmp, dest)
	}
	return fmt.Errorf("el zip no contiene %s", name)
}

// progressWriter imprime el avance cada 10 MB descargados.
type progressWriter struct {
	out         io.Writer
	total, done int64
	lastPrinted int64
}

const progressStep = 10 << 20

func (p *progressWriter) Write(b []byte) (int, error) {
	p.done += int64(len(b))
	if p.done-p.lastPrinted >= progressStep {
		p.lastPrinted = p.done
		if p.total > 0 {
			fmt.Fprintf(p.out, "\rDescargando: %d / %d MB", p.done>>20, p.total>>20)
		} else {
			fmt.Fprintf(p.out, "\rDescargando: %d MB", p.done>>20)
		}
	}
	return len(b), nil
}
