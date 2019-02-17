CosmosDB Adapter
====

CosmosDB Adapter is the [CosmosDB]adapter for [Casbin](https://github.com/casbin/casbin). With this library, Casbin can load policy from MongoDB or save policy to it.

## Installation

    go get github.com/SpacyCoder/cosmos-adapter

## Simple Example

```go
package main

import (
	"github.com/casbin/casbin"
	"github.com/spacycoder/cosmos-adapter"
)

func main() {
	// Initialize a MongoDB adapter and use it in a Casbin enforcer:
	// The adapter will use the database named "casbin".
	// If it doesn't exist, the adapter will create it automatically.
	a := mongodbadapter.NewAdapter("127.0.0.1:27017") // Your MongoDB URL. 
	
	// Or you can use an existing DB "abc" like this:
	// The adapter will use the table named "casbin_rule".
	// If it doesn't exist, the adapter will create it automatically.
	// a := mongodbadapter.NewAdapter("127.0.0.1:27017/abc")
	
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
import "github.com/spacycoder/test/cosmos"

// This adapter also implements the FilteredAdapter interface. This allows for
// efficent, scalable enforcement of very large policies:
filter := cosmos.SqlQuerySpec{Query: "SELECT * FROM root WHERE root.v0 = @v0", Parameters: []cosmos.QueryParam{{Name: "@v0", Value: "bob"}}}
e.LoadFilteredPolicy(filter)

// The loaded policy is now a subset of the policy in storage, containing only
// the policy lines that match the provided filter. This filter should be a
// valid MongoDB selector using BSON. A filtered policy cannot be saved.
```

## Getting Help

- [Casbin](https://github.com/casbin/casbin)

## License

This project is under Apache 2.0 License. See the [LICENSE](LICENSE) file for the full license text.
