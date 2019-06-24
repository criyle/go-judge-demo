package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Model is the database model as well as transfer model
type Model struct {
	ID     *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Status string              `json:"status" bson:"status,omitempty"`
	Time   uint64              `json:"time,omitempty" bson:"time,omitempty"`
	Memory uint64              `json:"memory,omitempty" bson:"memory,omitempty"`
	Date   uint64              `json:"date,omitempty" bson:"date,omitempty"`
	Lang   string              `json:"language,omitempty" bson:"language,omitempty"`
	Code   string              `json:"code,omitempty" bson:"code,omitempty"`
	Stdin  string              `json:"stdin,omitempty" bson:"stdin,omitempty"`
	Stdout string              `json:"stdout,omitempty" bson:"stdout,omitempty"`
	Stderr string              `json:"stderr,omitempty" bson:"stderr,omitempty"`
}

type db struct {
	client   *mongo.Client
	database *mongo.Database

	insert     chan ClientSubmit
	insertDone chan Model
	update     chan Model
	updateDone chan Model
}

const (
	colName         = "submission"
	defaultURI      = "mongodb://localhost:27017/admin"
	defaultDatabase = "test1"
	envMongoURI     = "MONGOURI"
)

func getDB() *db {
	uri := defaultURI
	database := defaultDatabase
	if u := os.Getenv(envMongoURI); u != "" {
		uri = u
		database = uri[strings.LastIndex(uri, "/")+1:]
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	//defer client.Disconnect(ctx)

	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return &db{
		client:     client,
		database:   client.Database(database),
		insert:     make(chan ClientSubmit, 64),
		insertDone: make(chan Model, 64),
		update:     make(chan Model, 64),
		updateDone: make(chan Model, 64),
	}
}

func (d *db) loop() {
	for {
		select {
		case cs := <-d.insert:
			di, err := d.Add(&cs)
			if err != nil {
				log.Println("db add:", err)
				continue
			}
			d.insertDone <- *di

		case jd := <-d.update:
			ud, err := d.Update(&jd)
			if err != nil {
				log.Println("db update:", err)
				continue
			}
			d.updateDone <- *ud
		}
	}
}

func (d *db) Add(cs *ClientSubmit) (*Model, error) {
	c := d.database.Collection(colName)
	m := &Model{
		Status: "Submitted",
		Lang:   cs.Lang,
		Code:   cs.Code,
		Date:   uint64(time.Now().Unix()),
	}
	i, err := c.InsertOne(nil, m)
	if err != nil {
		return nil, err
	}
	id := i.InsertedID.(primitive.ObjectID)
	m.ID = &id
	return m, nil
}

func (d *db) Update(m *Model) (*Model, error) {
	c := d.database.Collection(colName)

	filter := bson.D{{Key: "_id", Value: m.ID}}
	update := bson.D{{Key: "$set", Value: m}}

	_, err := c.UpdateOne(nil, filter, update)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (d *db) Query(id string) ([]Model, error) {
	c := d.database.Collection(colName)

	findOption := options.Find()
	findOption.SetLimit(10)
	findOption.SetSort(bson.D{{Key: "_id", Value: -1}})

	filter := bson.D{}
	if len(id) > 0 {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		filter = append(filter, bson.E{
			Key:   "_id",
			Value: bson.D{{Key: "$lt", Value: oid}},
		})
	}

	cursor, err := c.Find(nil, filter, findOption)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(nil)
	rt := make([]Model, 0, 10)
	for cursor.Next(nil) {
		el := Model{}
		if err = cursor.Decode(&el); err != nil {
			return nil, err
		}
		rt = append(rt, el)
	}
	return rt, nil
}
