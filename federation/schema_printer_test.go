package federation_test

import (
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/federation"
)

func TestSchemaPrinter_printSingleLineTypeDescription(t *testing.T) {
	helloWorldQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"helloWorld": &graphql.Field{
				Type:        graphql.String,
				Description: "This is a single line description",
			},
		},
	})
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: helloWorldQuery,
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `type Query {
  "This is a single line description"
  helloWorld: String
}`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{})
	if actual != expected {
		t.Fatalf(`Unexpected single line description. expected = %q, actual = %q`, expected, actual)
	}
}

func TestSchemaPrinter_printMultiLineTypeDescription(t *testing.T) {
	helloWorldQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"helloWorld": &graphql.Field{
				Type:        graphql.String,
				Description: "This is a mutli line description.\nAdditional details",
			},
		},
	})
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: helloWorldQuery,
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `type Query {
  """
  This is a mutli line description.
  Additional details
  """
  helloWorld: String
}`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{})
	if actual != expected {
		t.Fatalf(`Unexpected multi line description. expected = %q, actual = %q`, expected, actual)
	}
}

func TestSchemaPrinter_printSchemaDefinition(t *testing.T) {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"helloWorld": &graphql.Field{
					Type: graphql.String,
				},
			},
		}),
		Mutation: graphql.NewObject(graphql.ObjectConfig{
			Name: "Mutation",
			Fields: graphql.Fields{
				"goodbyeWorld": &graphql.Field{
					Type: graphql.Boolean,
				},
			},
		}),
		AppliedDirectives: []*graphql.AppliedDirective{
			{
				Name: "foo",
			},
		},
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `schema @foo {
  query: Query
  mutation: Mutation
}

type Mutation {
  goodbyeWorld: Boolean
}

type Query {
  helloWorld: String
}`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{IncludeSchemaDefinition: true})
	if actual != expected {
		t.Fatalf(`Unexpected schema definition. expected = %q, actual = %q`, expected, actual)
	}
}

func TestSchemaPrinter_printDirectiveDefinition(t *testing.T) {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"helloWorld": &graphql.Field{
					Type: graphql.String,
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `"Marks an element of a GraphQL schema as no longer supported."
directive @deprecated(reason: String) on FIELD_DEFINITION | ENUM_VALUE

"Directs the executor to include this field or fragment only when the ` + "`if`" + ` argument is true."
directive @include(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT

"Directs the executor to skip this field or fragment when the ` + "`if`" + ` argument is true."
directive @skip(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT

type Query {
  helloWorld: String
}`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{IncludeDirectiveDefinition: true})
	if actual != expected {
		t.Fatalf(`Unexpected directive definition. expected = %q, actual = %q`, expected, actual)
	}
}

func TestSchemaPrinter_printDeprecation(t *testing.T) {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"deprecatedQuery": &graphql.Field{
					Type:              graphql.String,
					DeprecationReason: "some reason",
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `type Query {
  deprecatedQuery: String @deprecated(reason: "some reason")
}`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{})
	if actual != expected {
		t.Fatalf(`Unexpected deprecation message. expected = %q, actual = %q`, expected, actual)
	}
}

func TestSchemaPrinter_printAppliedDirectives(t *testing.T) {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"queryWithDirective": &graphql.Field{
					Type: graphql.String,
					AppliedDirectives: []*graphql.AppliedDirective{
						{
							Name: "foo",
							Args: []*graphql.AppliedDirectiveArgument{
								{
									Name:  "name",
									Value: "value",
								},
							},
						},
						{
							Name: "bar",
						},
					},
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `type Query {
  queryWithDirective: String @bar @foo(name: "value")
}`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{})
	if actual != expected {
		t.Fatalf(`Unexpected applied directive. expected = %q, actual = %q`, expected, actual)
	}
}

func TestSchemaPrinter_printEnumDefinition(t *testing.T) {
	enumType := graphql.NewEnum(graphql.EnumConfig{
		Name:        "NUMBER",
		Description: "Number enum",
		Values: graphql.EnumValueConfigMap{
			"ONE": {
				Value:             "ONE",
				DeprecationReason: "use TWO",
			},
			"TWO": {
				Value:       "TWO",
				Description: "Replaces ONE",
			},
			"THREE": {
				Value: "THREE",
				AppliedDirectives: []*graphql.AppliedDirective{
					{
						Name: "bar",
					},
				},
			},
		},
		AppliedDirectives: []*graphql.AppliedDirective{
			{
				Name: "foo",
			},
		},
	})
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"enumQuery": &graphql.Field{
					Type: enumType,
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `"Number enum"
enum NUMBER @foo {
  ONE @deprecated(reason: "use TWO")
  THREE @bar
  "Replaces ONE"
  TWO
}

type Query {
  enumQuery: NUMBER
}`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{})
	if actual != expected {
		t.Fatalf(`Unexpected enum definition. expected = %q, actual = %q`, expected, actual)
	}
}

func TestSchemaPrinter_printInputObjectDefinition(t *testing.T) {
	fooInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        "Foo",
		Description: "Input object",
		Fields: graphql.InputObjectConfigFieldMap{
			"bar": {
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "bar field",
				AppliedDirectives: []*graphql.AppliedDirective{
					{
						Name: "customField",
					},
				},
			},
		},
		AppliedDirectives: []*graphql.AppliedDirective{
			{
				Name: "custom",
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"queryWithArgs": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"name": {
							Type: graphql.NewNonNull(graphql.String),
						},
						"foo": {
							Type: fooInput,
						},
					},
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `"Input object"
input Foo @custom {
  "bar field"
  bar: Int! @customField
}

type Query {
  queryWithArgs(name: String!, foo: Foo): String
}`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{})
	if actual != expected {
		t.Fatalf(`Unexpected input object definition. expected = %q, actual = %q`, expected, actual)
	}
}

func TestSchemaPrinter_printInterfaceDefinition(t *testing.T) {
	myInterface := graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "MyInterface",
		Description: "Interface description",
		Fields: graphql.Fields{
			"foo": &graphql.Field{
				Type: graphql.String,
			},
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object { return nil },
		AppliedDirectives: []*graphql.AppliedDirective{
			{
				Name: "foo",
			},
		},
	})

	implementation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Implementation",
		Fields: graphql.Fields{
			"foo": &graphql.Field{
				Type: graphql.String,
			},
			"bar": &graphql.Field{
				Type: graphql.Int,
			},
		},
		Interfaces: []*graphql.Interface{
			myInterface,
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"interfaceQuery": &graphql.Field{
					Type: myInterface,
				},
			},
		}),
		Types: []graphql.Type{
			implementation,
		},
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `"Interface description"
interface MyInterface @foo {
  foo: String
}

type Implementation implements MyInterface {
  bar: Int
  foo: String
}

type Query {
  interfaceQuery: MyInterface
}`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{})
	if actual != expected {
		t.Fatalf(`Unexpected interface definition. expected = %q, actual = %q`, expected, actual)
	}
}

func TestSchemaPrinter_printObjectDefinition(t *testing.T) {
	bar := graphql.NewObject(graphql.ObjectConfig{
		Name: "Bar",
		Fields: graphql.Fields{
			"baz": &graphql.Field{
				Type: graphql.Int,
				AppliedDirectives: []*graphql.AppliedDirective{
					{
						Name: "customField",
					},
				},
			},
		},
		AppliedDirectives: []*graphql.AppliedDirective{
			{
				Name: "custom",
			},
		},
	})

	foo := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Foo",
		Description: "Foo object",
		Fields: graphql.Fields{
			"foo": &graphql.Field{
				Type:        graphql.String,
				Description: "foo field",
			},
			"bar": &graphql.Field{
				Type: bar,
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"objectQuery": &graphql.Field{
					Type: graphql.NewNonNull(foo),
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `type Bar @custom {
  baz: Int @customField
}

"Foo object"
type Foo {
  bar: Bar
  "foo field"
  foo: String
}

type Query {
  objectQuery: Foo!
}`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{})
	if actual != expected {
		t.Fatalf(`Unexpected object definition. expected = %q, actual = %q`, expected, actual)
	}
}

func TestSchemaPrinter_printUnionDefinition(t *testing.T) {
	bar := graphql.NewObject(graphql.ObjectConfig{
		Name: "Bar",
		Fields: graphql.Fields{
			"bar": &graphql.Field{
				Type: graphql.Int,
			},
		},
	})

	foo := graphql.NewObject(graphql.ObjectConfig{
		Name: "Foo",
		Fields: graphql.Fields{
			"foo": &graphql.Field{
				Type: graphql.String,
			},
		},
	})
	myUnion := graphql.NewUnion(graphql.UnionConfig{
		Name:        "MyUnion",
		Description: "Union description",
		Types: []*graphql.Object{
			foo,
			bar,
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			return nil
		},
		AppliedDirectives: []*graphql.AppliedDirective{
			{
				Name: "custom",
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"unionQuery": &graphql.Field{
					Type: myUnion,
				},
			},
		}),
		Types: []graphql.Type{
			bar,
			foo,
		},
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `type Bar {
  bar: Int
}

type Foo {
  foo: String
}

type Query {
  unionQuery: MyUnion
}

"Union description"
union MyUnion @custom = Bar | Foo`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{})
	if actual != expected {
		t.Fatalf(`Unexpected union definition. expected = %q, actual = %q`, expected, actual)
	}
}

func TestSchemaPrinter_printScalarDefinition(t *testing.T) {
	customScalar := graphql.NewScalar(graphql.ScalarConfig{
		Name:        "Custom",
		Description: "Custom scalar",
		Serialize: func(value interface{}) interface{} {
			return nil
		},
		AppliedDirectives: []*graphql.AppliedDirective{
			{
				Name: "foo",
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"booleanQuery": &graphql.Field{
					Type: graphql.NewNonNull(graphql.Boolean),
				},
				"customScalarQuery": &graphql.Field{
					Type: customScalar,
				},
				"floatQuery": &graphql.Field{
					Type: graphql.Float,
				},
				"idQuery": &graphql.Field{
					Type: graphql.ID,
				},
				"intQuery": &graphql.Field{
					Type: graphql.Int,
				},
				"listNonNullQuery": &graphql.Field{
					Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.Int))),
				},
				"listNullableElementQuery": &graphql.Field{
					Type: graphql.NewNonNull(graphql.NewList(graphql.Int)),
				},
				"listQuery": &graphql.Field{
					Type: graphql.NewList(graphql.Int),
				},
				"stringQuery": &graphql.Field{
					Type: graphql.String,
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("Unable to construct test schema, reason: %q", err.Error())
	}

	expected := `type Query {
  booleanQuery: Boolean!
  customScalarQuery: Custom
  floatQuery: Float
  idQuery: ID
  intQuery: Int
  listNonNullQuery: [Int!]!
  listNullableElementQuery: [Int]!
  listQuery: [Int]
  stringQuery: String
}

"Custom scalar"
scalar Custom @foo`

	actual := federation.PrintSchema(schema, federation.PrinterOptions{})
	if actual != expected {
		t.Fatalf(`Unexpected scalar definition. expected = %q, actual = %q`, expected, actual)
	}
}
