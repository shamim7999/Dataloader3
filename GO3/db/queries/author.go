package db

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"graphql_test/db"
	"graphql_test/db/models"
	"graphql_test/domain"
	"log"
)

func GetDataFromAuthorCollection(filter bson.M) ([]models.Author, error) {
	cursor, err := db.CollectionAuthor.Find(db.Ctx, filter)
	//fmt.Println("Cursor for Author: ", cursor)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Println(err)
		}
	}(cursor, db.Ctx)
	var authors []models.Author
	for cursor.Next(db.Ctx) {
		var author models.Author
		if err := cursor.Decode(&author); err != nil {
			log.Println(err)
		}
		authors = append(authors, models.Author{
			ID:   author.ID,
			Name: author.Name,
		})
	}
	return authors, nil
}

func InsertAuthor(param *models.Author) (*domain.Author, error) {
	if _, err := db.CollectionAuthor.InsertOne(db.Ctx, param); err != nil {
		return nil, err
	}

	return &domain.Author{
		ID:   param.ID.Hex(),
		Name: param.Name,
	}, nil
}
