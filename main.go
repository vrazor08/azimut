package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed static templates favicon.ico
var content embed.FS

var tmpl = template.Must(template.ParseFS(content, "templates/index2.html"))

func main() {
	if err := run(); err != nil {
		log.Fatalf("[ERROR] main run error, %v", err)
	}
}

// TODO: add systemd and add more advanced logs
func run() error {
	// log.Printf("Starting...\n")
	f, err := os.Create("server.log")
	if err != nil {
		return err
	}
	defer f.Close()
	log.SetOutput(f)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(content)))
	mux.HandleFunc("/", index())
	mux.HandleFunc("/favicon.ico", GiveFavicon)

	srv := &http.Server{
		Addr:              ":80",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	ec := make(chan error, 1)
	go func() {
		log.Println("[INFO] Activate web server: http://195.24.66.249:80")
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			ec <- fmt.Errorf("server error: %w", err)
		}
		log.Print("[INFO] Stopped serving new connections")
		ec <- nil
	}()

	<-ctx.Done()
	stop()

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(timeoutCtx); err != nil {
		log.Printf("[WARN] Server shutdown error: %v", err)
	}

	log.Println("[INFO] Graceful shutdown complete")
	return <-ec
}

func index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] serving request: %s %s %s %s %s %s", r.Method, r.URL, r.Proto, r.Host, r.RemoteAddr, r.UserAgent())

		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("[WARN] failed to execute template, %v", err)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
}

func RedirectFunction(redirectPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] redirect request: %s %s %s %s %s %s on %s", r.Method, r.URL, r.Proto, r.Host, r.RemoteAddr, r.UserAgent(), redirectPath)
		http.Redirect(w, r, redirectPath, http.StatusFound)
	}
}

func GiveFavicon(w http.ResponseWriter, r *http.Request) {
	data, err := content.ReadFile("favicon.ico")
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("content-type", "image/x-icon")
	w.Write(data)
}
