package schema

import (
	"github.com/graphql-go/graphql"
	// "go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	// db "graphql_test/db/queries"
	// "graphql_test/domain"
)

var BookType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Book",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"title": &graphql.Field{
			Type: graphql.String,
		},
		"author_ids": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		// "authors": &graphql.Field{
		// 	Type: graphql.NewList(AuthorType),
		// 	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		// 		source := p.Source.(*domain.Book)
		// 		var oIds []interface{}
		// 		for _, item := range source.AuthorIds {
		// 			oid, err := primitive.ObjectIDFromHex(item)
		// 			if err == nil {
		// 				oIds = append(oIds, oid)
		// 			}
		// 		}
		// 		return db.GetDataFromAuthorCollection(bson.M{"_id": bson.M{"$in": oIds}})
		// 	},
		// },
	},
})
