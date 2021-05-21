* Entities like Model, Region, Host and Component have a Scope.
* Each Scope has Variables, which is just a `map[interface{}]interface{}`
* Each Scope can have a parent Scope, which is the Scope of the parent entity.

The leaf nodes of the Variables trees are instances of Variable.

Each Variable has the following fields

* Description
    * type: string
    * Used only for debugging. Needed?
* Default
    * type: interface{}
    * The default value for the variable. Only relevant if not required
* Required bool
    * type: bool
    * Whether a value for the variable must be specified
* Scoped
    * type: bool
    * Indicates whether the variable name should be prefixed with the entity path
* GlobalFallback
    * type: bool
    * Indicated whether the variable should be looked up without the scope prefix if it can't be found with the scope prefix
* Sensitive
    * type: bool
    * controls whether we output the value in dumps
* Binder
    * type: func(v *Variable, i interface{}, path ...string)
    * Allows models a hook to allow processing after the value is set
* Value
    * type: interface{}
    * The variables value

## Proposal

```go
package model

type VariableResolver interface {
	Resolve(entity Entity, name, path string) (interface{}, bool)
}
```

Have VariableResolverChain which aggregates multiple resolvers.

Have resolvers for the following:

1. Scope local cache, for already resolved variables
1. Parent Scope with path
1. Parent Scope without path
1. Binding with path
1. Binding without path
1. Command line override with path
1. Command line override without path
1. Environment variable with path
1. Environment variable without path
1. A constant value (used for providing defaults)

Thoughts:

1. Do we want to define all variables ahead of time? Can we resolve variables on the fly as they are needed and just list variable defaults and secret variables?