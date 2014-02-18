package neo4j_test

import (
    "testing"
    "github.com/jakewins/neo4j"
)

func TestShouldCreateHttpDriver(t *testing.T) {
    driver, err := neo4j.NewDriver("http://example.com")
    if driver == nil {
        t.Error("Expected a driver back.")
    }
    if err != nil {
        t.Error("Did not expect an error.")
    }
}

func TestGoodErrorOnUnknownConnectionScheme(t *testing.T) {
    driver, err := neo4j.NewDriver("nonsense://example.com")
    if driver != nil {
        t.Error("Did not expect a driver instance.")
    }
    if err == nil {
        t.Error("Expected an error result")
    }
    if err.Error() != "neo4j: Unknown connection scheme, nonsense." {
        t.Errorf("Expected error to be 'neo4j: Unknown connection scheme, nonsense.', but it was '%s'", err.Error())
    }
}