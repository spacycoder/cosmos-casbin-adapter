package cosmosadapter

import (
	"errors"
	"log"
	"strconv"

	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	"github.com/spacycoder/test/cosmos"
)

// CasbinRule represents a rule in Casbin.
type CasbinRule struct {
	ID    string `json:"id"`
	PType string `json:"pType"`
	V0    string `json:"v0"`
	V1    string `json:"v1"`
	V2    string `json:"v2"`
	V3    string `json:"v3"`
	V4    string `json:"v4"`
	V5    string `json:"v5"`
}

// adapter represents the CosmosDB adapter for policy storage.
type adapter struct {
	collectionName string
	databaseName   string
	collection     *cosmos.Collection
	db             *cosmos.Database
	client         *cosmos.Client
	filtered       bool
}

// NewAdapter is the constructor for Adapter.
// no options are given the database name is "casbin" and the collection is named casbin_rule
// if the database or collection is not found it is automatically created.
// the database can be changed by using the Database(db string) option.
// the collection can be changed by using the Collection(coll string) option.
// see README for example
func NewAdapter(connectionString string, options ...Option) persist.Adapter {
	client, err := cosmos.New(connectionString)
	if err != nil {
		log.Fatalf("Creating new cosmos client caused error: %s", err.Error())
	}
	a := &adapter{collectionName: "casbin_rule", databaseName: "casbin", client: client}

	for _, option := range options {
		option(a)
	}

	db := client.Database(a.databaseName)
	a.createDatabaseIfNotExist(db)
	collection := db.Collection(a.collectionName)
	a.createCollectionIfNotExist(collection)
	a.db = db
	a.collection = collection
	a.filtered = false
	return a
}

func (a *adapter) createDatabaseIfNotExist(db *cosmos.Database) {
	_, err := db.Read()
	if err != nil {
		if err, ok := err.(*cosmos.Error); ok {
			if err.NotFound() {
				a.client.Databases().Create(a.databaseName)
				if err != nil {
					log.Fatalf("Creating cosmos database caused error: %s", err.Error())
				}
			} else {
				log.Fatalf("Reading cosmos database caused error: %s", err.Error())
			}
		} else {
			log.Fatalf("Reading cosmos database caused error: %s", err.Error())
		}
	}
}

func (a *adapter) createCollectionIfNotExist(collection *cosmos.Collection) {
	_, err := collection.Read()
	if err != nil {
		if err, ok := err.(*cosmos.Error); ok {
			if err.NotFound() {
				collDef := &cosmos.CollectionDefinition{Resource: cosmos.Resource{ID: a.collectionName}, PartitionKey: cosmos.PartitionKeyDefinition{Paths: []string{"/pType"}, Kind: "Hash"}}
				_, err := a.db.Collections().Create(collDef)
				if err != nil {
					log.Fatalf("Creating cosmos collection caused error: %s", err.Error())
				}
			} else {
				log.Fatalf("Reading cosmos collection caused error: %s", err.Error())
			}
		} else {
			log.Fatalf("Reading cosmos collection caused error: %s", err.Error())
		}
	}
}

// NewFilteredAdapter is the constructor for FilteredAdapter.
// Casbin will not automatically call LoadPolicy() for a filtered adapter.
func NewFilteredAdapter(url string, options ...Option) persist.FilteredAdapter {
	a := NewAdapter(url, options...).(*adapter)
	a.filtered = true
	return a
}

func (a *adapter) dropCollection() error {
	_, err := a.collection.Delete()
	if err != nil {
		return err
	}
	_, err = a.db.Collections().Create(&cosmos.CollectionDefinition{Resource: cosmos.Resource{ID: a.collectionName}, PartitionKey: cosmos.PartitionKeyDefinition{Paths: []string{"/pType"}, Kind: "Hash"}})
	return err
}

func loadPolicyLine(line CasbinRule, model model.Model) {
	key := line.PType
	sec := key[:1]

	tokens := []string{}
	if line.V0 != "" {
		tokens = append(tokens, line.V0)
	} else {
		goto LineEnd
	}

	if line.V1 != "" {
		tokens = append(tokens, line.V1)
	} else {
		goto LineEnd
	}

	if line.V2 != "" {
		tokens = append(tokens, line.V2)
	} else {
		goto LineEnd
	}

	if line.V3 != "" {
		tokens = append(tokens, line.V3)
	} else {
		goto LineEnd
	}

	if line.V4 != "" {
		tokens = append(tokens, line.V4)
	} else {
		goto LineEnd
	}

	if line.V5 != "" {
		tokens = append(tokens, line.V5)
	} else {
		goto LineEnd
	}

LineEnd:
	model[sec][key].Policy = append(model[sec][key].Policy, tokens)
}

// LoadPolicy loads policy from database.
func (a *adapter) LoadPolicy(model model.Model) error {
	return a.LoadFilteredPolicy(model, nil)
}

// LoadFilteredPolicy loads matching policy lines from database. If not nil,
// the filter must be a valid MongoDB selector.
func (a *adapter) LoadFilteredPolicy(model model.Model, filter interface{}) error {
	lines := []CasbinRule{}
	if filter == nil {
		a.filtered = false
		res, err := a.collection.Documents().ReadAll(&lines, cosmos.CrossPartition())
		if err != nil {
			return err
		}
		tokenString := res.Continuation()
		for tokenString != "" {
			newLines := []CasbinRule{}
			res, err := a.collection.Documents().ReadAll(&newLines, cosmos.CrossPartition(), cosmos.Continuation(tokenString))
			if err != nil {
				return err
			}
			tokenString = res.Continuation()
			lines = append(lines, newLines...)
		}
	} else {
		querySpec := filter.(cosmos.SqlQuerySpec)
		a.filtered = true
		res, err := a.collection.Documents().Query(&querySpec, &lines, cosmos.CrossPartition())
		if err != nil {
			return err
		}
		tokenString := res.Continuation()
		for tokenString != "" {
			newLines := []CasbinRule{}
			res, err := a.collection.Documents().Query(&querySpec, &newLines, cosmos.CrossPartition(), cosmos.Continuation(tokenString))
			if err != nil {
				return err
			}
			tokenString = res.Continuation()
			lines = append(lines, newLines...)
		}
	}

	for _, line := range lines {
		loadPolicyLine(line, model)
	}

	return nil

}

// IsFiltered returns true if the loaded policy has been filtered.
func (a *adapter) IsFiltered() bool {
	return a.filtered
}

func savePolicyLine(ptype string, rule []string) CasbinRule {
	line := CasbinRule{
		PType: ptype,
	}

	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}

	return line
}

// SavePolicy saves policy to database.
func (a *adapter) SavePolicy(model model.Model) error {
	if a.filtered {
		return errors.New("cannot save a filtered policy")
	}
	if err := a.dropCollection(); err != nil {
		return err
	}

	var lines []CasbinRule

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			lines = append(lines, line)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			lines = append(lines, line)
		}
	}

	for _, line := range lines {
		_, err := a.collection.Documents().Create(&line, cosmos.PartitionKey(line.PType))
		if err != nil {
			return err
		}
	}
	return nil
}

// AddPolicy adds a policy rule to the storage.
func (a *adapter) AddPolicy(sec string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)
	_, err := a.collection.Documents().Create(&line, cosmos.PartitionKey(line.PType))
	return err
}

// RemovePolicy removes a policy rule from the storage.
func (a *adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	query := "SELECT * FROM root WHERE root.pType = @pType"
	parameters := []cosmos.QueryParam{{Name: "@pType", Value: ptype}}
	for i, value := range rule {
		indexString := strconv.Itoa(i)
		query += " AND root.v" + indexString + " = @v" + indexString
		parameters = append(parameters, cosmos.QueryParam{Name: "@v" + indexString, Value: value})
	}

	querySpec := cosmos.SqlQuerySpec{Parameters: parameters, Query: query}
	var policies []CasbinRule
	_, err := a.collection.Documents().Query(&querySpec, &policies, cosmos.PartitionKey(ptype))
	if err != nil {
		return err
	}

	for _, policy := range policies {
		_, err := a.collection.Document(policy.ID).Delete(cosmos.PartitionKey(policy.PType))
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
// @TODO IMPLEMENT
func (a *adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	selector := make(map[string]interface{})

	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		if fieldValues[0-fieldIndex] != "" {
			selector["v0"] = fieldValues[0-fieldIndex]
		}
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		if fieldValues[1-fieldIndex] != "" {
			selector["v1"] = fieldValues[1-fieldIndex]
		}
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		if fieldValues[2-fieldIndex] != "" {
			selector["v2"] = fieldValues[2-fieldIndex]
		}
	}
	if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) {
		if fieldValues[3-fieldIndex] != "" {
			selector["v3"] = fieldValues[3-fieldIndex]
		}
	}
	if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) {
		if fieldValues[4-fieldIndex] != "" {
			selector["v4"] = fieldValues[4-fieldIndex]
		}
	}
	if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) {
		if fieldValues[5-fieldIndex] != "" {
			selector["v5"] = fieldValues[5-fieldIndex]
		}
	}

	query := "SELECT * FROM root WHERE root.pType = @pType"
	parameters := []cosmos.QueryParam{{Name: "@pType", Value: ptype}}
	for key, value := range selector {
		query += " AND root." + key + " = @" + key
		parameters = append(parameters, cosmos.QueryParam{Name: "@" + key, Value: value})
	}

	querySpec := cosmos.SqlQuerySpec{Parameters: parameters, Query: query}
	var policies []CasbinRule
	_, err := a.collection.Documents().Query(&querySpec, &policies, cosmos.PartitionKey(ptype))
	if err != nil {
		return err
	}

	for _, policy := range policies {
		_, err := a.collection.Document(policy.ID).Delete(cosmos.PartitionKey(policy.PType))
		if err != nil {
			return err
		}
	}

	return nil
}

type Option func(*adapter)

func Database(db string) Option {
	return func(a *adapter) {
		a.databaseName = db
	}
}

func Collection(coll string) Option {
	return func(a *adapter) {
		a.collectionName = coll
	}
}
