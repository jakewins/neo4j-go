# Neo4j Go

An experimental Neo4j driver for the Go language. It sports a transactional API, and supports Cypher only.

### IMPORTANT: Relation to Neo4j project

While I work for Neo Technology, this is a personal project. While I am happy for you to use and
benefit from it, the project is maintained by a single developer on weekends, and it is still 
experimental.

As it is a personal project, Neo Technology commercial support does not apply to this code base,
nor does the backwards compatibility or stabilty guarantees of the Neo4j project.

## Usage

    import (
        "github.com/jakewins/neo4j-go"
    )

    // You'll generally want a single driver instance per Neo4j database used
    // by your application.
    drive, _ := neo4j.NewDriver("http://localhost:7474")

    // The driver allows you to create database Sessions
    // Sessions can, if you want, be pooled and re-used, although we use TCP 
    // connection pooling by default even if you don't reuse sessions. That 
    // is not guaranteed in future versions though.
    sess,_ := drive.NewSession()

    // Sessions allow you to create transactions
    tx,_ := sess.NewTransaction()
 
    // Run a Cypher statement like this
    res,_ := tx.Execute("CREATE (n) RETURN id(n)")

    // Loop through the result like below. Next() forwards the result cursor, 
    // which starts at -1. It returns true if there was another row in the result.
    var id int64
    for res.Next() {
        // Once we're on a row, we can read the columns in it
        id = res.GetInt("id(n)")
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

For more details, see driver.go for interfaces, and http_test.go for integration tests.

## Contributing

Contributing to this project is super welcome! To make sure your contributions get merged in, please
contact me ahead of time if you are making anything but trivial changes, to ensure your changes are in
line with the plans for the driver.

You can reach me at jake [at] neotechnology.com

## License

http://www.apache.org/licenses/LICENSE-2.0.html