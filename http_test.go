package neo4j_test

import (
    "testing"
    "github.com/jakewins/neo4j"
)

// NOTE: These are integration tests. They will fail unless you have a neo 
// instance running at localhost:7474.


func ExampleUsage() {
    // This should be a single instance per application
    drive, _ := neo4j.NewDriver("http://localhost:7474")

    // This can, if you want, be pooled and re-used, although we use TCP 
    // connection pooling by default even if you don't reuse sessions. That 
    // is not guaranteed in future versions though.
    sess,_ := drive.NewSession()

    // This denotes an atomic transaction against Neo4j
    tx,_ := sess.NewTransaction()
 
    // Run a statement without parameters like this
    res,_ := tx.Execute("CREATE (n) RETURN id(n)")

    // Loop through the result like below. Next forwards the result cursor, 
    // and starts at -1. It returns true if there was another row in the result.
    for res.Next() {
        // Once we're on a row, we can read the columns in it
        id := res.GetInt("id(n)")
    }

    // And it's considered good form to close the result when you're done.
    res.Close()

    // However, you will generally not want raw cypher statement, you'll want
    // to use parameters. This allows re-use of query plans, and gets rid of 
    // issues of SQL-injection style vulnerabilities. {id} is the name of a
    // parameter in thie example
    res,_ = tx.ExecuteWithParams("START n=node({id}) RETURN id(n)", map[string]interface{}{ "id":id })
    res.Close()

    // Finally, we commit our changes (or tx.Rollback() if we're unhappy with them)
    tx.Commit()
}

func TestBegin__Execute__Commit(t *testing.T) {

    // Given
    drive, _ := neo4j.NewDriver("http://localhost:7474")
    sess,_ := drive.NewSession()
    tx,_ := sess.NewTransaction()
 
    res,_ := tx.Execute("CREATE (n) RETURN id(n)")

    if res.Next() != true {
        t.Error("Expected one result back.") 
    }

    id := res.GetInt("id(n)")
    if id < 0 {
        t.Error("Expected an id >= 0 back from the server.")
    } 
 
    if res.Next() != false {
        t.Error("Expected only one result.") 
    }

    // When
    tx.Commit()

    // Then
    tx,_ = sess.NewTransaction()
    res,_ = tx.ExecuteWithParams("START n=node({id}) RETURN id(n)", map[string]interface{}{ "id":id })
    res.Next()
    idAgain := res.GetInt("id(n)")
    if id != idAgain {
        t.Error("Expected the transaction to have commmitted.")
    } 
}

func TestBegin__Execute__Rollback(t *testing.T) {

    // Given
    drive, _ := neo4j.NewDriver("http://localhost:7474")
    sess,_ := drive.NewSession()
    tx,_ := sess.NewTransaction()
 
    _, err := tx.Execute("CREATE (n:ShouldNotBeCommitted)")

    if err != nil {
        t.Error(err)
    }

    // When
    tx.Rollback()

    // Then
    tx,_ = sess.NewTransaction()
    res,_ := tx.Execute("MATCH (n:ShouldNotBeCommitted) RETURN n")
    
    if res.Next() != false {
        t.Error("Expected node to not get committed")
    } 
}

func TestCellTypes(t *testing.T) {

    // Given
    drive, _ := neo4j.NewDriver("http://localhost:7474")
    sess,_ := drive.NewSession()
    tx,_ := sess.NewTransaction()

    // When
    res, _ := tx.Execute("CREATE (n:ShouldNotBeCommitted) RETURN 1.2 as float, 1 as int, false as bool, \"hello\" as string, {k:1} as map, [1,2.2] as array")
    res.Next()

    // Then array should be correct (TODO: Is there something like hamcrest matchers for go? This array assertion is silly.)
    if int(res.GetArray("array")[0].(float64)) != 1 {
        t.Error("Expected entry 0 to be 1")
    }
    if res.GetArray("array")[1].(float64) != 2.2 {
        t.Error("Expected entry 1 to be 2.2")
    }

    // Then map should be correct
    if res.GetMap("map")["k"].(float64) != 1.0 {
        t.Error("Expected map to be k->1")
    }

    // Then primitives should be correct
    if res.GetFloat("float") != 1.2 {
        t.Error("Expected float to be 1.2")
    }
    if res.GetInt("int") != 1 {
        t.Error("Expected int to be 1")
    }
    if res.GetBool("bool") != false {
        t.Error("Expected bool to be false")
    }
    if res.GetString("string") != "hello" {
        t.Error("Expected string to be 'hello'")
    }

}