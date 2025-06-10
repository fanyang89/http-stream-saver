package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/DataDog/zstd"
)

var flagBaseDir string

func saveStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileName := filepath.Base(r.URL.Path)
	if fileName == "" || fileName == "." {
		http.Error(w, "Invalid file name", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(flagBaseDir, fileName+".zst")
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating file: %v", err), http.StatusInternalServerError)
		return
	}
	z := zstd.NewWriter(file)
	defer func() {
		_ = z.Flush()
		_ = file.Close()
	}()

	_, err = io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error writing file: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "File saved to %s\n", filePath)
}

func main() {
	var err error

	listen := flag.String("listen", ":5000", "listen address")
	flag.StringVar(&flagBaseDir, "base-dir", "/tmp", "data directory")
	flag.Parse()

	flagBaseDir, err = filepath.Abs(flagBaseDir)
	if err != nil {
		panic(err)
	}
	fmt.Printf("listening on %s, base directory: %s\n", *listen, flagBaseDir)

	http.HandleFunc("/", saveStream)
	err = http.ListenAndServe(*listen, nil)
	if err != nil {
		panic(err)
	}
}
