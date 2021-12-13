package model

import (
	"context"
	store "xtest/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CollectionName = "models"

var ctx = context.TODO()

type Model struct {
	Symbol         string  `json:"symbol" bson:"symbol"`
	Price          float64 `json:"price_24h" bson:"price_24h"`
	Volume         float64 `json:"volume_24h" bson:"volume_24h"`
	LastTradePrice float64 `json:"last_trade_price" bson:"last_trade_price"`
}

func (m *Model) Save(store store.Storage) error {

	f := bson.M{"symbol": m.Symbol}

	_, err := store.Database().Collection(CollectionName).UpdateOne(ctx, f, bson.M{"$set": m}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}
func FindBySymbol(store store.Storage, fltr string) ([]*Model, error) {

	f := bson.M{"symbol": fltr}

	cur, err := store.Database().Collection(CollectionName).Find(ctx, f, options.Find())
	if err != nil {
		return nil, err
	}

	var course []*Model

	if err := cur.All(ctx, &course); err != nil {
		return nil, err
	}

	return course, nil
}
