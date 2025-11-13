package server

import (
	_ "embed"
	"fmt"
	"goaria/internal/ariarpc"
	"goaria/internal/handlers"
	"goaria/internal/middleware"
	"net/http"
)

//go:embed index.html
var indexHTML []byte

type Server struct {
	mux    *http.ServeMux
	hm     *handlers.HandlerManager
	port   string
	server *http.Server
}

// Create new server with all relevant configuration, mainly port
func NewServer(port string, ariaClient *ariarpc.AriaClient) *Server {

	s := &Server{
		mux:  http.NewServeMux(),
		hm:   handlers.NewHandlerManager(ariaClient),
		port: port,
	}

	s.setupMuxRoutes()
	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: s.mux,
	}

	return s

}

func (s *Server) setupMuxRoutes() {

	// Unprotected for login and logout
	s.mux.HandleFunc("POST /login", s.hm.LoginHandler)
	s.mux.HandleFunc("POST /logout", s.hm.LogoutHandler)

	// Serves .html file embedded in binary
	s.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexHTML)
	})

	// Protected routes (require session)
	s.mux.Handle("GET /downloads/active",
		middleware.SessionMiddleware(http.HandlerFunc(s.hm.ActiveDownloadsHandler)),
	)
	s.mux.Handle("POST /download",
		middleware.SessionMiddleware(http.HandlerFunc(s.hm.DownloadHandler)),
	)

	s.mux.Handle("POST /pause/{gid}",
		middleware.SessionMiddleware(http.HandlerFunc(s.hm.PauseDownloadHandler)),
	)

	s.mux.Handle("POST /unpause/{gid}",
		middleware.SessionMiddleware(http.HandlerFunc(s.hm.UnpauseDownloadHandler)),
	)

	s.mux.Handle("POST /remove/{gid}",
		middleware.SessionMiddleware(http.HandlerFunc(s.hm.RemoveDownloadHandler)),
	)

	s.mux.Handle("GET /getdownloaddir",
		middleware.SessionMiddleware(http.HandlerFunc(s.hm.GetDownloadDirHandler)),
	)

}

// Runs server in separate goroutine, if done channel receives we know that
// aria2c process was properly finished by user
func (s *Server) Run(done chan error) {

	go func() {
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("[server] crash errpr: %v\n", err)
		}

	}()

	<-done

	if err := s.server.Close(); err != nil {
		fmt.Printf("[server] close error: %v\n", err)

	} else {
		fmt.Printf("[server] closed successfully")

	}
}
