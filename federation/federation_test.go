package federation_test

import (
	"encoding/json"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/federation"
)

type Product struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

func buildSubgraph(includeQuery bool) (graphql.Schema, error) {
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

	var schemaConfig graphql.SchemaConfig
	if includeQuery {
		schemaConfig = graphql.SchemaConfig{
			Query: graphql.NewObject(graphql.ObjectConfig{
				Name: "Query",
				Fields: graphql.Fields{
					"product": &graphql.Field{
						Name: "product",
						Type: productType,
						Args: graphql.FieldConfigArgument{
							"id": {
								Type: graphql.NewNonNull(graphql.ID),
							},
						},
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							return &Product{ID: "1", Description: "Foo"}, nil
						},
					},
				},
			}),
			Types: []graphql.Type{
				productType,
			},
		}
	} else {
		schemaConfig = graphql.SchemaConfig{
			Types: []graphql.Type{
				productType,
			},
		}
	}

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
		SchemaConfig: schemaConfig,
	})
	return schema, err
}

func TestFederation_buildSubraphWithQuery(t *testing.T) {
	schema, _ := buildSubgraph(true)

	serviceSDLQuery := `query {
	  _service { sdl }
	}`

	params := graphql.Params{Schema: schema, RequestString: serviceSDLQuery}
	sdlResponse := graphql.Do(params)
	if len(sdlResponse.Errors) > 0 {
		t.Errorf("failed to execute _service { sdl } query, errors: %+v", sdlResponse.Errors)
	}

	data, _ := sdlResponse.Data.(map[string]interface{})
	_service, _ := data["_service"].(map[string]interface{})
	sdl, _ := _service["sdl"].(string)

	expected := `schema @link(url: "https://specs.apollo.dev/federation/v2.1", import: ["composeDirective", "external", "inaccessible", "key", "override", "provides", "requires", "shareable", "tag", "FieldSet"]) {
  query: Query
}

directive @composeDirective(name: String!) repeatable on SCHEMA

"Marks an element of a GraphQL schema as no longer supported."
directive @deprecated(reason: String) on FIELD_DEFINITION | ENUM_VALUE

"Marks target field as external meaning it will be resolved by federated schema"
directive @external on FIELD_DEFINITION

"Marks location within schema as inaccessible from the GraphQL Gateway"
directive @inaccessible on FIELD_DEFINITION | OBJECT | INTERFACE | UNION | ENUM | ENUM_VALUE | SCALAR | INPUT_OBJECT | INPUT_FIELD_DEFINITION | ARGUMENT_DEFINITION

"Directs the executor to include this field or fragment only when the ` + "`if`" + ` argument is true."
directive @include(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT

"Space separated list of primary keys needed to access federated object"
directive @key(fields: FieldSet!, resolvable: Boolean) repeatable on OBJECT | INTERFACE

directive @link(url: String!, import: [[String]]) repeatable on SCHEMA

"Overrides fields resolution logic from other subgraph. Used for migrating fields from one subgraph to another."
directive @override(from: String!) on FIELD_DEFINITION

"Specifies locally selectable fields on a given entity"
directive @provides(fields: FieldSet!) on FIELD_DEFINITION

"Specifies external federated fields required for computing this field value"
directive @requires(fields: FieldSet!) on FIELD_DEFINITION

"Indicates that given object and/or field can be resolved by multiple subgraphs"
directive @shareable on FIELD_DEFINITION | OBJECT

"Directs the executor to skip this field or fragment when the ` + "`if`" + ` argument is true."
directive @skip(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT

"Allows users to annotate fields and types with additional metadata information"
directive @tag(name: String!) repeatable on FIELD_DEFINITION | OBJECT | INTERFACE | UNION | ARGUMENT_DEFINITION | SCALAR | ENUM | ENUM_VALUE | INPUT_OBJECT | INPUT_FIELD_DEFINITION

type Product @key(fields: "id", resolvable: true) {
  description: String
  id: ID!
}

type Query {
  _entities(representations: [_Any!]!): [_Entity]
  _service: _Service
  product(id: ID!): Product
}

type _Service {
  sdl: String!
}

union _Entity = Product

"String-serialized scalar represents a set of fields that's passed to a federated directive, such as @key, @requires, or @provides"
scalar FieldSet

"The ` + "`_Any`" + ` scalar is used to pass representations of entities from external services into the root _entities field for execution."
scalar _Any`

	if expected != sdl {
		t.Fatalf("_service { sdl } query returned unexpected schema.\n\texpected = %q\n\n\tactual = %q", expected, sdl)
	}

	entityQuery := `query EntityQuery($_representations: [_Any!]!) {
	  _entities(representations: $_representations) {
	    ... on Product {
	      id
		  description
		}
	  }
	}`

	entityParams := graphql.Params{
		Schema:        schema,
		RequestString: entityQuery,
		VariableValues: map[string]interface{}{
			"_representations": []map[string]interface{}{
				{"__typename": "Product", "id": "1"},
			},
		}}
	entityResponse := graphql.Do(entityParams)
	if len(entityResponse.Errors) > 0 {
		t.Errorf("failed to execute entities query, errors: %+v", entityResponse.Errors)
	}

	entities, _ := json.Marshal(entityResponse)
	expected_entity := `{"data":{"_entities":[{"description":"Federated Description","id":"1"}]}}`

	if string(entities) != expected_entity {
		t.Fatalf("_entities query returned unexpected result.\n\texpected = %q\n\n\tactual = %q", expected_entity, entities)
	}
}

func TestFederation_buildSubgraphWithoutQuery(t *testing.T) {
	schema, _ := buildSubgraph(false)

	serviceSDLQuery := `query {
	  _service { sdl }
	}`

	params := graphql.Params{Schema: schema, RequestString: serviceSDLQuery}
	sdlResponse := graphql.Do(params)
	if len(sdlResponse.Errors) > 0 {
		t.Errorf("failed to execute _service { sdl } query, errors: %+v", sdlResponse.Errors)
	}

	data, _ := sdlResponse.Data.(map[string]interface{})
	_service, _ := data["_service"].(map[string]interface{})
	sdl, _ := _service["sdl"].(string)

	expected := `schema @link(url: "https://specs.apollo.dev/federation/v2.1", import: ["composeDirective", "external", "inaccessible", "key", "override", "provides", "requires", "shareable", "tag", "FieldSet"]) {
  query: Query
}

directive @composeDirective(name: String!) repeatable on SCHEMA

"Marks an element of a GraphQL schema as no longer supported."
directive @deprecated(reason: String) on FIELD_DEFINITION | ENUM_VALUE

"Marks target field as external meaning it will be resolved by federated schema"
directive @external on FIELD_DEFINITION

"Marks location within schema as inaccessible from the GraphQL Gateway"
directive @inaccessible on FIELD_DEFINITION | OBJECT | INTERFACE | UNION | ENUM | ENUM_VALUE | SCALAR | INPUT_OBJECT | INPUT_FIELD_DEFINITION | ARGUMENT_DEFINITION

"Directs the executor to include this field or fragment only when the ` + "`if`" + ` argument is true."
directive @include(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT

"Space separated list of primary keys needed to access federated object"
directive @key(fields: FieldSet!, resolvable: Boolean) repeatable on OBJECT | INTERFACE

directive @link(url: String!, import: [[String]]) repeatable on SCHEMA

"Overrides fields resolution logic from other subgraph. Used for migrating fields from one subgraph to another."
directive @override(from: String!) on FIELD_DEFINITION

"Specifies locally selectable fields on a given entity"
directive @provides(fields: FieldSet!) on FIELD_DEFINITION

"Specifies external federated fields required for computing this field value"
directive @requires(fields: FieldSet!) on FIELD_DEFINITION

"Indicates that given object and/or field can be resolved by multiple subgraphs"
directive @shareable on FIELD_DEFINITION | OBJECT

"Directs the executor to skip this field or fragment when the ` + "`if`" + ` argument is true."
directive @skip(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT

"Allows users to annotate fields and types with additional metadata information"
directive @tag(name: String!) repeatable on FIELD_DEFINITION | OBJECT | INTERFACE | UNION | ARGUMENT_DEFINITION | SCALAR | ENUM | ENUM_VALUE | INPUT_OBJECT | INPUT_FIELD_DEFINITION

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

"String-serialized scalar represents a set of fields that's passed to a federated directive, such as @key, @requires, or @provides"
scalar FieldSet

"The ` + "`_Any`" + ` scalar is used to pass representations of entities from external services into the root _entities field for execution."
scalar _Any`

	if expected != sdl {
		t.Fatalf("_service { sdl } query returned unexpected schema.\n\texpected = %q\n\n\tactual = %q", expected, sdl)
	}
}
