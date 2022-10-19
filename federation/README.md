# Apollo Federation module for graphql-go/graphql

This module provides [Apollo Federation](https://www.apollographql.com/docs/federation/) support to the `graphql-go/graphql` library.

## Usage

Apollo Federation relies on the usage of directives to specify relationships between various subgraphs. All federated directives are provided in the `federation` package.

Once our GraphQL schema is defined, then we can then generate Federation compatible schema using `federation.NewFederatedSchema(config FederatedSchemaConfig)` function. 
In order to be able to resolve the federated `Product` type, we need to provide `EntityTypeResolver` to resolve `_Entity` union type and a `EntitiesFieldResolver` to resolve 
`_entities` query.

```golang
type Product struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

productType := graphql.NewObject(graphql.ObjectConfig{
    Name: "Product",
    Fields: graphql.Fields{
        "id": &graphql.Field{
            Name: "id",
            Type: graphql.NewNonNull(graphql.ID),
        },
        "description": &graphql.Field{
            Name: "description",
            Type: graphql.String,
        },
    },
    AppliedDirectives: []*graphql.AppliedDirective{
        federation.KeyAppliedDirective("id", true),
    },
})

schema, err := federation.NewFederatedSchema(federation.FederatedSchemaConfig{
    EntitiesFieldResolver: func(p graphql.ResolveParams) (interface{}, error) {
        representations, ok := p.Args["representations"].([]interface{})
        results := make([]interface{}, 0)
        if ok {
            for _, representation := range representations {
                raw, isAny := representation.(map[string]interface{})
                if isAny {
                    typeName, _ := raw["__typename"].(string)
                    if typeName == "Product" {
                        id, _ := raw["id"].(string)
                        product := &Product{
                            ID:          id,
                            Description: "Federated Description",
                        }
                        results = append(results, product)
                    }
                }
            }
        }
        return results, nil
    },
    EntityTypeResolver: func(p graphql.ResolveTypeParams) *graphql.Object {
        if _, ok := p.Value.(*Product); ok {
            return productType
        }
        return nil
    },
    SchemaConfig: schemaConfig = graphql.SchemaConfig{
        // root query can be omitted when creating federated schema
        // Query: graphql.NewObject(graphql.ObjectConfig{...}) // specify your query object
        Types: []graphql.Type{
            // if entities are not resolvable by queries, they have to be explicitly specified here
            productType,
        },
    },
})
```

Above code will generate the following GraphQL schema (directive definitions and scalars are omitted for brevity):

```graphql
schema @link(url: "https://specs.apollo.dev/federation/v2.1", import: ["composeDirective", "external", "inaccessible", "key", "override", "provides", "requires", "shareable", "tag", "FieldSet"]) {
  query: Query
}

type Product @key(fields: "id", resolvable: true) {
  description: String
  id: ID!
}

type Query {
  _entities(representations: [_Any!]!): [_Entity]
  _service: _Service
}

type _Service {
  sdl: String!
}

union _Entity = Product
```
