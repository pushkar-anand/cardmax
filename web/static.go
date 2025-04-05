package web

import (
	"embed"
	"github.com/pushkar-anand/build-with-go/logger"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
)

//go:embed static/*
var staticFiles embed.FS

func StaticFilesHandler() http.Handler {
	return http.FileServer(http.FS(staticFiles))
}

func NamedFileHandler(
	log *slog.Logger,
	pathToFile ...string,
) http.HandlerFunc {
	if len(pathToFile) == 0 {
		panic("pathToFile must not be empty")
	}

	fileName := filepath.Join("static", filepath.Join(pathToFile...))

	return func(w http.ResponseWriter, r *http.Request) {
		file, err := staticFiles.Open(fileName)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to open file", slog.String("filename", fileName), logger.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}

		defer func() { _ = file.Close() }()

		_, err = io.Copy(w, file)
		if err != nil {
			log.ErrorContext(r.Context(), "failed to write named file", slog.String("filename", fileName), logger.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return
		}
	}
}
