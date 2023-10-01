package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"graphql_test/db"
	"graphql_test/queries"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v", result.Errors)
	}
	return result
}

type ReqBody struct {
	Query string `json:"query"`
}

func main() {

	db.ClientOptions = options.Client().ApplyURI(
		"mongodb://admin:secret@localhost:27017/",
	).SetDirect(true)

	db.Ctx = context.Background()
	db.Client, db.Err = mongo.Connect(db.Ctx, db.ClientOptions)
	if db.Err != nil {
		log.Fatal(db.Err)
	}
	defer db.Client.Disconnect(db.Ctx)

	fmt.Println("Connected to MongoDB")

	db.Database = db.Client.Database("shamim")
	db.CollectionBook = db.Database.Collection("Book")
	db.CollectionAuthor = db.Database.Collection("Author")
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		var t ReqBody
		err = json.Unmarshal(body, &t)
		if err != nil {
			panic(err)
		}
		result := executeQuery(t.Query, queries.GetRootSchema())
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			log.Println(err)
		}
	})
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	fmt.Println("Access the web app via browser at 'http://localhost:8080'")

	http.ListenAndServe(":8080", nil)
}
