package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	store "xtest/storage"
	"xtest/storage/model"

	"github.com/gorilla/mux"
)

type Server struct {
	Router *mux.Router
	Store  *store.Store
	Config *Config
	Srv    http.Server
}

func NewServer(config *Config) *Server {
	server := &Server{
		Router: mux.NewRouter(),
		Config: config,
	}

	return server
}
func (s *Server) Start() error {
	store, err := store.New(s.Config.Store.DBUrl, s.Config.Store.DBName)
	if err != nil {
		return err
	}
	s.Store = store
	if err = s.Store.Connect(); err != nil {
		return err
	}
	s.configureRouter()
	s.Srv.Handler = s.Router
	s.Srv.Addr = s.Config.Server.Port
	go s.serve()
	go s.update()
	return nil
}

func (s *Server) configureRouter() {
	http.Handle("/", s.Router)
	s.Router.HandleFunc("/courses", s.VerifyCourses).Methods(http.MethodPost)
}
func (s *Server) serve() {
	if err := s.Srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Print("Listen error", err)
	}
}
func (s *Server) update() {
	for {
		<-time.After(30 * time.Second)
		resp, err := http.Get("https://api.blockchain.com/v3/exchange/tickers")
		if err != nil {
			log.Println(err)
			continue
		}
		mod := []model.Model{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			continue
		}
		resp.Body.Close()
		err = json.Unmarshal(body, &mod)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, v := range mod {
			err := v.Save(s.Store)
			if err != nil {
				log.Println(err)
			}
		}
		log.Println("Courses updated")
	}
}
