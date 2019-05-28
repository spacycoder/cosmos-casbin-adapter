# CosmosDB Adapter

CosmosDB Adapter is the cosmosDB adapter for [Casbin](https://github.com/casbin/casbin). With this library, Casbin can load policy from CosmosDB or save policy to it.

## Installation

    go get github.com/spacycoder/cosmos-adapter

## Simple Example

```go
package main

import (
	"github.com/casbin/casbin"
	cosmosadapter "github.com/spacycoder/cosmos-casbin-adapter"
)

func main() {
	// Initialize a CosmosDB adapter and use it in a Casbin enforcer:
	// The first argument is the cosmos connection string.
	// The second argument takes options as its input.
	// if not option is given the default database name is "casbin" and the default collection name is "casbin_rule"
	// The adapter will try to create the database and collection if it does not find them.
	a := cosmosadapter.NewAdapter("connstring")
	e := casbin.NewEnforcer("examples/rbac_model.conf", a)

	// Load the policy from DB.
	e.LoadPolicy()

	// Check the permission.
	e.Enforce("alice", "data1", "read")

	// Modify the policy.
	// e.AddPolicy(...)
	// e.RemovePolicy(...)

	// Save the policy back to DB.
	e.SavePolicy()
}
```

## With options

```go
package main

import (
	"github.com/casbin/casbin"
	cosmosadapter "github.com/spacycoder/cosmos-casbin-adapter"
)

func main() {
	// Initialize a CosmosDB adapter and use it in a Casbin enforcer:
	// The adapter will try to create the database and collection if it does not find them.
	a := cosmosadapter.NewAdapter("connstring", cosmosadapter.Database("mycasbindb"), cosmosadapter.Collection("mycasbincollection"))
	e := casbin.NewEnforcer("examples/rbac_model.conf", a)

	// Load the policy from DB.
	e.LoadPolicy()

	// Check the permission.
	e.Enforce("alice", "data1", "read")

	// Modify the policy.
	// e.AddPolicy(...)
	// e.RemovePolicy(...)

	// Save the policy back to DB.
	e.SavePolicy()
}
```

## Filtered Policies

```go
import "github.com/spacycoder/cosmosdb-go-sdk/cosmos"

// This adapter also implements the FilteredAdapter interface. This allows for
// efficent, scalable enforcement of very large policies:

filter := cosmos.Q{Query: "SELECT * FROM root WHERE root.v0 = @v0", Parameters: []cosmos.P{{Name: "@v0", Value: "bob"}}}
e.LoadFilteredPolicy(filter)

// The loaded policy is now a subset of the policy in storage, containing only
// the policy lines that match the provided filter.
```

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
