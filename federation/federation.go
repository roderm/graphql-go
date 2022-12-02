package federation

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

type FederatedSchemaConfig struct {
	EntitiesFieldResolver graphql.FieldResolveFn
	EntityTypeResolver    graphql.ResolveTypeFn
	graphql.SchemaConfig
}

// federated types

type _Any map[string]interface{}

var _AnyType = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "_Any",
	Description: "The `_Any` scalar is used to pass representations of entities from external services into the root _entities field for execution.",
	Serialize: func(value interface{}) interface{} {
		switch value := value.(type) {
		case _Any:
			return value
		case *_Any:
			return *value
		default:
			return nil
		}
	},
	ParseValue: func(value interface{}) interface{} {
		return value
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.ListValue:
			return valueAST.Values
		case *ast.BooleanValue:
			return valueAST.Value
		case *ast.EnumValue:
			return valueAST.Value
		case *ast.FloatValue:
			return valueAST.Value
		case *ast.IntValue:
			return valueAST.Value
		case *ast.ObjectValue:
			literalValue := make(map[string]ast.Value)
			for _, field := range valueAST.Fields {
				literalValue[field.Name.Value] = field.Value
			}
			return literalValue
		case *ast.StringValue:
			return valueAST.Value
		default:
			return nil
		}
	},
})

func coerceString(value interface{}) interface{} {
	if v, ok := value.(*string); ok {
		if v == nil {
			return nil
		}
		return *v
	}
	return fmt.Sprintf("%v", value)
}

var _FieldSetType = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "FieldSet",
	Description: "String-serialized scalar represents a set of fields that's passed to a federated directive, such as @key, @requires, or @provides",
	// coercing logic is the same as String
	Serialize: func(value interface{}) interface{} {
		return coerceString(value)
	},
	ParseValue: func(value interface{}) interface{} {
		return coerceString(value)
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			return valueAST.Value
		}
		return nil
	},
})

type _Service struct {
	SDL string `json:"sdl"`
}

var _ServiceType = graphql.NewObject(graphql.ObjectConfig{
	Name: "_Service",
	Fields: graphql.Fields{
		"sdl": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
})

func findEntityTypes(schema graphql.Schema) []*graphql.Object {
	entities := make([]*graphql.Object, 0)

	for _, gqlType := range schema.TypeMap() {
		obj, ok := gqlType.(*graphql.Object)
		if ok && isEntity(obj) {
			entities = append(entities, obj)
		}
	}
	return entities
}

func isEntity(t *graphql.Object) bool {
	if t.Extend {
		return false
	}
	for _, directive := range t.AppliedDirectives {
		if directive.Name == "key" {
			return true
		}
	}
	return false
}

// @link(import : ["@composeDirective", "@external", "@inaccessible", "@key", "@override", "@provides", "@requires", "@shareable", "@tag", "@FieldSet"], url : "https://specs.apollo.dev/federation/v2.1")
var federationLinkAppliedDirective = LinkAppliedDirective(
	"https://specs.apollo.dev/federation/v2.1",
	[]string{"@composeDirective", "@external", "@inaccessible", "@key", "@override", "@provides", "@requires", "@shareable", "@tag", "FieldSet"},
)

// new schema

func NewFederatedSchema(config FederatedSchemaConfig) (graphql.Schema, error) {
	// add federated directives
	config.Directives = append(config.Directives,
		// built-in directives
		graphql.DeprecatedDirective,
		graphql.IncludeDirective,
		graphql.SkipDirective,
		// federated directives
		ComposeDirectiveDefinition,
		ExternalDirectiveDefinition,
		InaccessibleDirectiveDefinition,
		KeyDirectiveDefinition,
		LinkDirectiveDefinition,
		OverrideDirectiveDefinition,
		ProvidesDirectiveDefinition,
		RequiresDirectiveDefinition,
		ShareableDirectiveDefinition,
		TagDirectiveDefinition,
	)

	// add @link directive to the schema
	config.AppliedDirectives = append(config.AppliedDirectives, federationLinkAppliedDirective)
	// add federated types
	// scalar _Any
	// scalar FieldSet
	if config.Types == nil {
		config.Types = make([]graphql.Type, 0)
	}
	config.Types = append(config.Types, _AnyType, _FieldSetType, _ServiceType)
	// ensure there is a valid query type
	query := config.Query
	if query == nil {
		query = graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"_service": &graphql.Field{
					Name: "_service",
					Type: _ServiceType,
				},
			},
		})
		config.Query = query
	} else {
		query.AddFieldConfig("_service", &graphql.Field{
			Name: "_service",
			Type: _ServiceType,
		})
	}

	schema, err := graphql.NewSchema(config.SchemaConfig)
	if err != nil {
		panic("failure to create schema" + err.Error())
	}

	// find entities
	entities := findEntityTypes(schema)
	if len(entities) == 0 {
		entities = append(entities, graphql.NewObject(graphql.ObjectConfig{
			Name: "_ExtendHelper",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Name: "id",
					Type: graphql.NewNonNull(graphql.ID),
					AppliedDirectives: []*graphql.AppliedDirective{
						ExternalAppliedDirective,
					},
				},
			},
		}))
	}

	entityType := graphql.NewUnion(
		graphql.UnionConfig{
			Name:        "_Entity",
			Types:       entities,
			ResolveType: config.EntityTypeResolver,
		},
	)
	schema.TypeMap()["_Entity"] = entityType

	schema.QueryType().AddFieldConfig("_entities", &graphql.Field{
		Name: "_entities",
		Type: graphql.NewNonNull(graphql.NewList(entityType)),
		Args: graphql.FieldConfigArgument{
			"representations": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(_AnyType))),
			},
		},
		Resolve: config.EntitiesFieldResolver,
	})

	sdl := PrintSchema(schema, DefaultPrinterOptions)

	schema.QueryType().AddFieldConfig("_service", &graphql.Field{
		Name: "_service",
		Type: _ServiceType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return &_Service{SDL: sdl}, nil
		},
	})
	return schema, err
}
