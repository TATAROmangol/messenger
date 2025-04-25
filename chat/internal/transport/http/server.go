package httpserver

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type Service interface {
	ChatService
}

type Server struct {
	ctx        context.Context
	httpServer *http.Server
	Handler    *Handler
	Auther     Auther
}

func New(ctx context.Context, cfg Config, service Service, auther Auther) *Server {
	r := mux.NewRouter()
	chatHandler := NewHandler(service)
	s := &Server{
		ctx:     ctx,
		Handler: chatHandler,
		Auther:  auther,
	}
	setupRouter(r, s)
	s.httpServer = &http.Server{
		Addr:         cfg.Address(),
		Handler:      r,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	return s
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}
func (s *Server) MustRun() {
	if err := s.Run(); err != nil {
		panic(err)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func setupRouter(r *mux.Router, s *Server) {
	r.Use(InitLoggerCtxMiddleware(s.ctx))
	r.Use(AuthMiddleware(s.Auther))
	r.Handle("/chat/", s.Handler.GetChatsHandler()).Methods("GET")
	r.Handle("/chat/create", s.Handler.NewChatHandler()).Methods("POST")
	r.Handle("/chat/{chat_id:[0-9]+}", s.Handler.SendMessageHandler()).Methods("POST")
	r.Handle("/chat/{chat_id:[0-9]+}/messages", s.Handler.GetMessagesHandler()).Methods("GET")
}
