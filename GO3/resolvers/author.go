package resolvers

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"graphql_test/db"
	"graphql_test/db/models"
	db2 "graphql_test/db/queries"
	//"graphql_test/domain"
	//"log"

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
			fmt.Printf("Error parsing author ID: %v", err)
			continue
		}
		authorIDs = append(authorIDs, authorHexID)
	}
	books, err := fetchBooksByAuthorIDs(ctx, authorIDs)
	if err != nil {
		fmt.Printf("Error fetching todos: %v", err)
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
		return nil, fmt.Printf("error fetching books: %v", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			fmt.Printf("error closing cursor: %v", err)
		}
	}()

	for cursor.Next(ctx) {
		var domainBook models.Book
		if err := cursor.Decode(&domainBook); err != nil {
			return nil, fmt.Printf("error decoding domain.Book: %v", err)
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

	if !isOK {
		return nil, nil
	}
	loaderResult := booksLoader.Load(params.Context, dataloader.StringKey(authorID))

	loadedBooks, err := loaderResult()
	if err != nil {
		fmt.Printf("Error loading todos with DataLoader: %v", err)
		return nil, err
	}

	books, ok := loadedBooks.([]models.Book)
	if !ok {
		fmt.Printf("Error from DataLoader: %v", err)
		return nil, fmt.Printf("failed to assert todo type")
	}

	authorHexID, err := primitive.ObjectIDFromHex(authorID)
	if err != nil {
		fmt.Printf("Error parsing author ID: %v", err)
		return nil, err
	}

	filter := bson.M{"_id": authorHexID}

	var author models.Author
	err = db.CollectionAuthor.FindOne(context.TODO(), filter).Decode(&author)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("Author not found")
		} else {
			fmt.Fatal(err)
		}
	}

	if err != nil {
		fmt.Printf("Error fetching author: %v", err)
		return nil, err
	}

	result := map[string]interface{}{
		"author": author,
		"books":  books,
	}
	return result, nil

}
