package queries

import (
	"graphql_test/resolvers"
	"graphql_test/schema"

	"github.com/graphql-go/graphql"
)

var RootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: graphql.Fields{
		"getBooks": &graphql.Field{
			Type:        graphql.NewList(schema.BookType),
			Description: "Returns  all books",
			//Args:        graphql.FieldConfigArgument{},
			Resolve: resolvers.GetBooks,
		},

		"getAuthors": &graphql.Field{
			Type:        graphql.NewList(schema.AuthorType),
			Description: "Returns  all Authors",
			//Args:        graphql.FieldConfigArgument{},
			Resolve: resolvers.GetAuthors,
		},

		"getAuthorAndBooks": &graphql.Field{
			Type: graphql.NewObject(graphql.ObjectConfig{
				Name: "AuthorAndBooks",
				Fields: graphql.Fields{
					"author": &graphql.Field{
						Type: schema.AuthorType,
					},
					"books": &graphql.Field{
						Type: graphql.NewList(schema.BookType),
					},
				},
			}),
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			//Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			//	//fmt.Println("I am here")
			//	fmt.Println(p)
			//	return resolvers.ResolveGetAuthorAndBooks(p)
			//},

			Resolve: resolvers.ResolveGetAuthorAndBooks,
		},
	},
})
