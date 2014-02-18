package neo4j

import (
    "fmt"
    "errors"
    "net/url"
)

type Driver interface {
 
    NewSession() (Session, error)

}

type Session interface {

    NewTransaction() (Transaction, error)

}

type Transaction interface {

   Execute(string) (Result, error)
   ExecuteWithParams(string, map[string]interface{}) (Result, error)

   Commit() (error)
   Rollback() (error)

}

type Result interface {

    Next() bool
    Close()

    GetString(column string) (string)
    GetInt(column string) (int64)
    GetFloat(column string) (float64)
    GetBool(column string) (bool)
    
    GetMap(column string) (map[string]interface{})
    GetArray(column string) ([]interface{})
}

func NewDriver( connectionStr string ) (Driver, error) {
    url, err := url.Parse(connectionStr)
    if(err != nil) {
        return nil, errors.New("neo4j: Invalid connection string.")
    }
    
    switch url.Scheme {
    case "http":
        fallthrough
    case "https":
        return NewHttpDriver( url )
    default:
        return nil, fmt.Errorf("neo4j: Unknown connection scheme, %s.", url.Scheme)
    }
}