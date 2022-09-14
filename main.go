package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var clientOptions = options.Client().ApplyURI("mongodb://localhost:27017")
var db, err = mongo.Connect(context.TODO(), clientOptions)
var todosCollection = db.Database("ToDo").Collection("todos")

func main() {
	if err != nil {
		panic(err)
	}

	if err := db.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	handleRequests()
}

func handleRequests() {
	http.HandleFunc("/todos", showList)
	http.HandleFunc("/new", addItem)
	http.HandleFunc("/delete", deleteItem)
	http.HandleFunc("/edit", editItem)
	log.Fatal(http.ListenAndServe(":4000", nil))
}

func showList(w http.ResponseWriter, r *http.Request) {
	todoSlice, err := todosCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}

	var todos []bson.M
	if err = todoSlice.All(context.TODO(), &todos); err != nil {
		panic(err)
	}

	json.NewEncoder(w).Encode(todos)
}

func addItem(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	json.NewDecoder(r.Body).Decode(&body)

	var newItem = bson.D{
		{Key: "note", Value: body["note"]},
		{Key: "date", Value: body["date"]},
	}

	result, err := todosCollection.InsertOne(context.TODO(), newItem)

	if err != nil {
		panic(err)
	}
	print(result)

	var response = map[string]string{
		"message": "Todo successfully added",
	}
	json.NewEncoder(w).Encode(response)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	json.NewDecoder(r.Body).Decode(&body)

	objectId, err := primitive.ObjectIDFromHex(body["id"])
	if err != nil {
		panic(err)
	}

	filter := bson.M{"_id": objectId}
	result, err := todosCollection.DeleteOne(context.TODO(), filter)

	if err != nil {
		panic(err)
	}
	print(result)

	var response = map[string]string{
		"message": "Todo successfully deleted",
	}
	json.NewEncoder(w).Encode(response)
}

func editItem(w http.ResponseWriter, r *http.Request) {
	var body map[string]string

	if json.NewDecoder(r.Body).Decode(&body); err != nil {
		panic(err)
	}
	defer r.Body.Close()

	var newItem = bson.D{
		{Key: "note", Value: body["note"]},
		{Key: "date", Value: body["date"]},
	}

	objectId, err := primitive.ObjectIDFromHex(body["id"])
	if err != nil {
		panic(err)
	}

	filter := bson.M{"_id": objectId}
	result, err := todosCollection.ReplaceOne(context.TODO(), filter, newItem)

	if err != nil {
		panic(err)
	}
	print(result)

	var response = map[string]string{
		"message": "Todo successfully edited",
	}
	json.NewEncoder(w).Encode(response)
}
