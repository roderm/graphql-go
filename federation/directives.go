package federation

import (
	"github.com/graphql-go/graphql"
)

//
// directive definitions
//

// directive @composeDirective(name: String!) repeatable on SCHEMA
var ComposeDirectiveDefinition = &graphql.Directive{
	Name: "composeDirective",
	Args: []*graphql.Argument{
		{
			PrivateName: "name",
			Type:        graphql.NewNonNull(graphql.String),
		},
	},
	Locations: []string{
		graphql.DirectiveLocationSchema,
	},
	Repeatable: true,
}

// directive @contact(name: String!, url: String, description: String) on SCHEMA
var ContactDirectiveDefinition = &graphql.Directive{
	Name:        "contact",
	Description: "Provides contact information of the owner responsible for this subgraph schema.",
	Args: []*graphql.Argument{
		{
			PrivateName:        "name",
			Type:               graphql.NewNonNull(graphql.String),
			PrivateDescription: "Contact title of the subgraph owner",
		},
		{
			PrivateName:        "url",
			Type:               graphql.String,
			PrivateDescription: "URL where the subgraph's owner can be reached",
		},
		{
			PrivateName:        "description",
			Type:               graphql.String,
			PrivateDescription: "Other relevant notes can be included here; supports markdown links",
		},
	},
	Locations: []string{
		graphql.DirectiveLocationSchema,
	},
}

// directive @external on FIELD_DEFINITION
var ExternalDirectiveDefinition = &graphql.Directive{
	Name:        "external",
	Description: "Marks target field as external meaning it will be resolved by federated schema",
	Locations: []string{
		graphql.DirectiveLocationFieldDefinition,
	},
}

// directive @inaccessible on FIELD_DEFINITION | OBJECT | INTERFACE | UNION | ENUM | ENUM_VALUE | SCALAR | INPUT_OBJECT | INPUT_FIELD_DEFINITION | ARGUMENT_DEFINITION
var InaccessibleDirectiveDefinition = &graphql.Directive{
	Name:        "inaccessible",
	Description: "Marks location within schema as inaccessible from the GraphQL Gateway",
	Locations: []string{
		graphql.DirectiveLocationFieldDefinition,
		graphql.DirectiveLocationObject,
		graphql.DirectiveLocationInterface,
		graphql.DirectiveLocationUnion,
		graphql.DirectiveLocationEnum,
		graphql.DirectiveLocationEnumValue,
		graphql.DirectiveLocationScalar,
		graphql.DirectiveLocationInputObject,
		graphql.DirectiveLocationInputFieldDefinition,
		graphql.DirectiveLocationArgumentDefinition,
	},
}

// directive @key(fields: FieldSet!, resolvable: Boolean = true) repeatable on OBJECT | INTERFACE
var KeyDirectiveDefinition = &graphql.Directive{
	Name:        "key",
	Description: "Space separated list of primary keys needed to access federated object",
	Args: []*graphql.Argument{
		{
			PrivateName: "fields",
			Type:        graphql.NewNonNull(_FieldSetType),
		},
		{
			PrivateName:  "resolvable",
			Type:         graphql.Boolean,
			DefaultValue: true,
		},
	},
	Locations: []string{
		graphql.DirectiveLocationObject,
		graphql.DirectiveLocationInterface,
	},
	Repeatable: true,
}

// directive @link(url: String, import: [Import]) repeatable on SCHEMA
var LinkDirectiveDefinition = &graphql.Directive{
	Name: "link",
	Args: []*graphql.Argument{
		{
			PrivateName: "url",
			Type:        graphql.NewNonNull(graphql.String),
		},
		{
			PrivateName: "import",
			Type:        graphql.NewList(graphql.String),
		},
	},
	Locations: []string{
		graphql.DirectiveLocationSchema,
	},
	Repeatable: true,
}

// directive @override(from: String!) on FIELD_DEFINITION
var OverrideDirectiveDefinition = &graphql.Directive{
	Name:        "override",
	Description: "Overrides fields resolution logic from other subgraph. Used for migrating fields from one subgraph to another.",
	Args: []*graphql.Argument{
		{
			PrivateName: "from",
			Type:        graphql.NewNonNull(graphql.String),
		},
	},
	Locations: []string{
		graphql.DirectiveLocationFieldDefinition,
	},
}

// directive @provides(fields: FieldSet!) on FIELD_DEFINITION
var ProvidesDirectiveDefinition = &graphql.Directive{
	Name:        "provides",
	Description: "Specifies locally selectable fields on a given entity",
	Args: []*graphql.Argument{
		{
			PrivateName: "fields",
			Type:        graphql.NewNonNull(_FieldSetType),
		},
	},
	Locations: []string{
		graphql.DirectiveLocationFieldDefinition,
	},
}

// directive @requires(fields: FieldSet!) on FIELD_DEFINITION
var RequiresDirectiveDefinition = &graphql.Directive{
	Name:        "requires",
	Description: "Specifies external federated fields required for computing this field value",
	Args: []*graphql.Argument{
		{
			PrivateName: "fields",
			Type:        graphql.NewNonNull(_FieldSetType),
		},
	},
	Locations: []string{
		graphql.DirectiveLocationFieldDefinition,
	},
}

// directive @shareable on FIELD_DEFINITION | OBJECT
var ShareableDirectiveDefinition = &graphql.Directive{
	Name:        "shareable",
	Description: "Indicates that given object and/or field can be resolved by multiple subgraphs",
	Locations: []string{
		graphql.DirectiveLocationFieldDefinition,
		graphql.DirectiveLocationObject,
	},
}

// directive @tag(name: String!) repeatable on FIELD_DEFINITION | OBJECT | INTERFACE | UNION | ARGUMENT_DEFINITION | SCALAR | ENUM | ENUM_VALUE | INPUT_OBJECT | INPUT_FIELD_DEFINITION
var TagDirectiveDefinition = &graphql.Directive{
	Name:        "tag",
	Description: "Allows users to annotate fields and types with additional metadata information",
	Args: []*graphql.Argument{
		{
			PrivateName: "name",
			Type:        graphql.NewNonNull(graphql.String),
		},
	},
	Locations: []string{
		graphql.DirectiveLocationFieldDefinition,
		graphql.DirectiveLocationObject,
		graphql.DirectiveLocationInterface,
		graphql.DirectiveLocationUnion,
		graphql.DirectiveLocationArgumentDefinition,
		graphql.DirectiveLocationScalar,
		graphql.DirectiveLocationEnum,
		graphql.DirectiveLocationEnumValue,
		graphql.DirectiveLocationInputObject,
		graphql.DirectiveLocationInputFieldDefinition,
	},
	Repeatable: true,
}

//
// applied directives
//

// schema @link(url: "my spec url", import: ["@myDirective"]) @composeDirective(name: "@myDirective")
func ComposeDirectiveAppliedDirective(name string) *graphql.AppliedDirective {
	return &graphql.AppliedDirective{
		Name: "composeDirective",
		Args: []*graphql.AppliedDirectiveArgument{
			{
				Name:  "name",
				Value: name,
			},
		},
	}
}

// @contact(name: "my team name", url: "slack url", description: "additional contact info")
func ContactAppliedDirective(name string, url string, description string) *graphql.AppliedDirective {
	return &graphql.AppliedDirective{
		Name: "contact",
		Args: []*graphql.AppliedDirectiveArgument{
			{
				Name:  "name",
				Value: name,
			},
			{
				Name:  "url",
				Value: url,
			},
			{
				Name:  "description",
				Value: description,
			},
		},
	}
}

// @external
var ExternalAppliedDirective = &graphql.AppliedDirective{
	Name: "external",
}

//	type Foo {
//	  id: String!
//	  secret: String! @inaccessible
//	}
var InaccessibleAppliedDirective = &graphql.AppliedDirective{
	Name: "inaccessible",
}

// @key(fields: "foo bar")
func KeyAppliedDirective(fieldSet string, resolvable bool) *graphql.AppliedDirective {
	return &graphql.AppliedDirective{
		Name: "key",
		Args: []*graphql.AppliedDirectiveArgument{
			{
				Name:  "fields",
				Value: fieldSet,
			},
			{
				Name:  "resolvable",
				Value: resolvable,
			},
		},
	}
}

// @link(url: "spec url", imports: ["foo"])
func LinkAppliedDirective(url string, imports []string) *graphql.AppliedDirective {
	return &graphql.AppliedDirective{
		Name: "link",
		Args: []*graphql.AppliedDirectiveArgument{
			{
				Name:  "url",
				Value: url,
			},
			{
				Name:  "import",
				Value: imports,
			},
		},
	}
}

// override(from: "subgraphA")
func OverrideAppliedDirective(from string) *graphql.AppliedDirective {
	return &graphql.AppliedDirective{
		Name: "override",
		Args: []*graphql.AppliedDirectiveArgument{
			{
				Name:  "from",
				Value: from,
			},
		},
	}
}

// @provides(fields: "foo")
func ProvidesAppliedDirective(fieldSet string) *graphql.AppliedDirective {
	return &graphql.AppliedDirective{
		Name: "provides",
		Args: []*graphql.AppliedDirectiveArgument{
			{
				Name:  "fields",
				Value: fieldSet,
			},
		},
	}
}

// @requires(fields: "foo")
func RequiresAppliedDirective(fieldSet string) *graphql.AppliedDirective {
	return &graphql.AppliedDirective{
		Name: "requires",
		Args: []*graphql.AppliedDirectiveArgument{
			{
				Name:  "fields",
				Value: fieldSet,
			},
		},
	}
}

// @shareable
var ShareableAppliedDirective = &graphql.AppliedDirective{
	Name: "shareable",
}

// @tag(name: "foo")
func TagAppliedDirective(value string) *graphql.AppliedDirective {
	return &graphql.AppliedDirective{
		Name: "tag",
		Args: []*graphql.AppliedDirectiveArgument{
			{
				Name:  "name",
				Value: value,
			},
		},
	}
}
