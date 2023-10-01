package resolvers

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"graphql_test/db"
	"graphql_test/db/models"
	db2 "graphql_test/db/queries"
	//"graphql_test/domain"
	"log"

	"github.com/graph-gophers/dataloader"
	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var booksLoader = dataloader.NewBatchedLoader(batchLoadBooksByAuthorIDs)

func GetAuthors(p graphql.ResolveParams) (interface{}, error) {
	return db2.GetDataFromAuthorCollection(bson.M{})
}

func CreateNewAuthor(p graphql.ResolveParams) (interface{}, error) {
	var name string

	if val, ok := p.Args["name"].(string); ok {
		name = val
	}

	return db2.InsertAuthor(&models.Author{
		ID:   primitive.NewObjectID(),
		Name: name,
	})
}

//////////////////////////// Data Loader /////////////////

func batchLoadBooksByAuthorIDs(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	booksByAuthor := make(map[string][]models.Book)

	var authorIDs []primitive.ObjectID
	for _, key := range keys {
		authorID := key.String()
		authorHexID, err := primitive.ObjectIDFromHex(authorID)
		if err != nil {
			log.Printf("Error parsing author ID: %v", err)
			continue
		}
		authorIDs = append(authorIDs, authorHexID)
	}
	books, err := fetchBooksByAuthorIDs(ctx, authorIDs)
	if err != nil {
		log.Printf("Error fetching todos: %v", err)
	}

	for _, book := range books {
		for _, id := range book.AuthorIds {
			booksByAuthor[id] = append(booksByAuthor[id], book)
		}
	}

	var results []*dataloader.Result
	for _, key := range keys {
		authorID := key.String()
		books := booksByAuthor[authorID]
		results = append(results, &dataloader.Result{Data: books, Error: nil})
	}

	return results
}

func fetchBooksByAuthorIDs(ctx context.Context, authorIDs []primitive.ObjectID) ([]models.Book, error) {
	var books []models.Book
	var AuthorIDsHex []string
	for _, ids := range authorIDs {
		AuthorIDsHex = append(AuthorIDsHex, ids.Hex())
	}

	query := bson.M{"author_ids": bson.M{"$in": AuthorIDsHex}}

	cursor, err := db.CollectionBook.Find(context.TODO(), query)

	if err != nil {
		return nil, fmt.Errorf("error fetching books: %v", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("error closing cursor: %v", err)
		}
	}()

	for cursor.Next(ctx) {
		var domainBook models.Book
		if err := cursor.Decode(&domainBook); err != nil {
			return nil, fmt.Errorf("error decoding domain.Book: %v", err)
		}

		auhtorIDHex, _ := primitive.ObjectIDFromHex(domainBook.ID.Hex())

		books = append(books, models.Book{
			ID:        auhtorIDHex,
			Title:     domainBook.Title,
			AuthorIds: domainBook.AuthorIds,
		})
	}

	return books, nil
}

func ResolveGetAuthorAndBooks(params graphql.ResolveParams) (interface{}, error) {
	authorID, isOK := params.Args["id"].(string)

	//fmt.Print("In ResolveGetAuthorAndBooks ", authorID, " ", isOK, "\n")

	if !isOK {
		return nil, nil
	}
	//fmt.Print("Here\n")
	loaderResult := booksLoader.Load(params.Context, dataloader.StringKey(authorID))

	loadedBooks, err := loaderResult()
	if err != nil {
		log.Printf("Error loading todos with DataLoader: %v", err)
		return nil, err
	}

	// Convert the result to the appropriate type (slice of AuthorTodo)
	books, ok := loadedBooks.([]models.Book)
	if !ok {
		log.Printf("Error asserting todo type from DataLoader: %v", err)
		return nil, fmt.Errorf("failed to assert todo type")
	}

	authorHexID, err := primitive.ObjectIDFromHex(authorID)
	if err != nil {
		log.Printf("Error parsing author ID: %v", err)
		return nil, err
	}

	//var author models.Author
	//authorQuery := bson.M{"_id": authorHexID}
	//fmt.Println(authorHexID)
	//author, err := db2.GetDataFromAuthorCollection(authorQuery)

	filter := bson.M{"_id": authorHexID}

	// Find the author document using the filter
	var author models.Author
	err = db.CollectionAuthor.FindOne(context.TODO(), filter).Decode(&author)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("Author not found")
		} else {
			log.Fatal(err)
		}
	}

	if err != nil {
		log.Printf("Error fetching author: %v", err)
		return nil, err
	}
	//fmt.Print("Here\n")
	// Combine the author and todos into the result
	result := map[string]interface{}{
		"author": author,
		"books":  books,
	}
	return result, nil

}
