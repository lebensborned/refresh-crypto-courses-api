package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
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
	wg     sync.WaitGroup
	ch     chan struct{}
	ticker *time.Ticker
}

func NewServer(config *Config) *Server {
	server := &Server{
		Router: mux.NewRouter(),
		Config: config,
	}

	return server
}
func (s *Server) Start() error {
	ch := make(chan struct{})
	s.ch = ch
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
	s.ticker = time.NewTicker(30 * time.Second)
	s.wg.Add(2)
	go s.serve()
	go s.update()
	log.Println("Server started")
	return nil
}

func (s *Server) configureRouter() {
	http.Handle("/", s.Router)
	s.Router.HandleFunc("/courses", s.VerifyCourses).Methods(http.MethodPost)
}
func (s *Server) Close() error {
	s.ch <- struct{}{}
	s.wg.Wait()
	if err := s.Store.Disconnect(); err != nil {
		return err
	}
	return nil
}
func (s *Server) serve() {
	defer s.wg.Done()
	if err := s.Srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Print("Listen error", err)
	}
}
func (s *Server) update() {
	defer s.wg.Done()
UPDLOOP:
	for {
		select {
		case <-s.ticker.C:
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
		case <-s.ch:
			s.ticker.Stop()
			break UPDLOOP
		}
	}
}
