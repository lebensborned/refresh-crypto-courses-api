package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"xtest/storage/model"
)

type Response struct {
	Price          float64 `json:"price"`
	Volume         float64 `json:"volume"`
	LastTradePrice float64 `json:"last_trade_price"`
}

func (s *Server) VerifyCourses(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("Bad model"))
		return
	}
	models := []model.Model{}
	err = json.Unmarshal(body, &models)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Cant unmarshal JSON"))
		return
	}
	resp := map[string]Response{}
	for _, v := range models {
		res, err := model.FindBySymbol(s.Store, v.Symbol)
		if err != nil {
			log.Println(err)
			continue
		}
		if len(res) == 0 {
			log.Println("skip cycle iteration, bad symbol name:", v.Symbol)
			continue
		}
		resp[res[0].Symbol] = Response{
			Price:          res[0].Price,
			Volume:         res[0].Volume,
			LastTradePrice: res[0].LastTradePrice,
		}
	}
	data, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Cant marshal JSON"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
