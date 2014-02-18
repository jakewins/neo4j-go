package neo4j

import (
    "encoding/json"
    "net/url"
    "net/http"
    "io/ioutil"
    "strings"
    "fmt"
)




func NewHttpDriver(url *url.URL) (Driver, error) {
    tr := &http.Transport{MaxIdleConnsPerHost: 6}
    client := &http.Client{Transport: tr}
    return &HttpDriver{ url.String(), client }, nil
}


type HttpDriver struct /* implements Driver */ {

    url string
    client *http.Client

}

type HttpSession struct {
    url string
    client *http.Client
}

type HttpTransaction struct {
    url string
    client *http.Client
    started bool
}

type HttpResult struct {
    rows []interface{}
    columns map[string]int
    errors []interface{}
    cursor int
}

//
// Driver
//

func (driver *HttpDriver) NewSession() (Session, error) {
    return &HttpSession{ driver.url, driver.client }, nil
}

// 
// Session
//

func (session *HttpSession) NewTransaction() (Transaction, error) {
    return &HttpTransaction{ session.url + "/db/data/transaction" , session.client, false }, nil
}

//
// Transaction
//


func (tx *HttpTransaction) Execute(statement string) (Result, error) {
    res, err := tx.ExecuteWithParams( statement, make(map[string]interface{}) )
    return res, err
}

func (tx *HttpTransaction) ExecuteWithParams(statement string, params map[string]interface{}) (Result, error) {
    // NOTE: This is messy on purpose. I want to get an idea of how to do this in Go
    // before I refactor it. Thus, mess in here temporarily.

    data, err := json.Marshal(map[string]interface{}{
        "statements": []map[string]interface{}{
            {"statement" : statement, "parameters" : params},
        },
    })

    req, _ := http.NewRequest("POST", tx.url, strings.NewReader(string(data)))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Connection", "Keep-Alive")
    
    resp, err := tx.client.Do(req)
    if err != nil {
        return nil, err
    }

    // Close the request when we exit this function
    defer resp.Body.Close()

    // Pull out the transaction location header, if one was given
    location, err := resp.Location()
    if location != nil {
        tx.started = true
        tx.url = location.String()
    }

    // Pull out the request body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // Read the JSON response neo gave us
    result := make(map[string]interface{})
    err = json.Unmarshal(body, &result)
    if err != nil {
        return nil, err
    }

    // Verify there were no errors on the server side
    errors := result["errors"].([]interface{})
    for _, neoError := range errors {
        // Obv. do error mapping here
        fmt.Println(neoError)
        return nil, nil
    }

    // Pull the result out of the response and give it to the user
    results := result["results"].([]interface{})
    firstResult := results[0].(map[string]interface{})
    columnArray := firstResult["columns"].([]interface{})
    rows := firstResult["data"].([]interface{})

    // Build a map of column name -> id
    columns := make(map[string]int)
    for index,key := range columnArray {
        columns[key.(string)] = index
    }

    return &HttpResult{rows, 
                       columns,
                       errors,
                       -1}, nil
}

func (tx *HttpTransaction) Commit() (error) {

    data, err := json.Marshal(map[string]interface{}{
        "statements": []interface{}{},
    })

    req, _ := http.NewRequest("POST", tx.url + "/commit", strings.NewReader(string(data)))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Connection", "Keep-Alive")
    
    resp, err := tx.client.Do(req)
    if err != nil {
        return err
    }

    // Close the request when we exit this function
    defer resp.Body.Close()

    // Pull out the request body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }

    // Read the JSON response neo gave us
    result := make(map[string]interface{})
    err = json.Unmarshal(body, &result)
    if err != nil {
        return err
    }

    // Verify there were no errors on the server side
    errors := result["errors"].([]interface{})
    for _, neoError := range errors {
        // Obv. do error mapping here
        fmt.Println(neoError)
        return nil
    }

    return nil
}

func (tx *HttpTransaction) Rollback() (error) {
    if tx.started {
        req, _ := http.NewRequest("DELETE", tx.url, nil)
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Connection", "Keep-Alive")
        _, err := tx.client.Do(req)
        return err
    }
    return nil
}

//
// Result
//

func (res *HttpResult) Next() bool {
    res.cursor++
    return res.cursor < len(res.rows)
}

func (res *HttpResult) Close(){
    // Saved for the future, to allow good knowledge of what the user is up to
    // to tell if we can stop streaming data from the server and so on.
}

func (res *HttpResult) GetString(column string) (string){
    return res.getRaw(column).(string)
}

func (res *HttpResult) GetInt(column string) (int64){
    return int64(res.getRaw(column).(float64))
}

func (res *HttpResult) GetFloat(column string) (float64){
    return res.getRaw(column).(float64)
}

func (res *HttpResult) GetBool(column string) (bool){
    return res.getRaw(column).(bool)
}

func (res *HttpResult) GetMap(column string) (map[string]interface{}){
    return res.getRaw(column).(map[string]interface{})
}

func (res *HttpResult) GetArray(column string) ([]interface{}){
    return res.getRaw(column).([]interface{})
}

func (res *HttpResult) getRaw(column string) (interface{}){
    rowData := res.rows[res.cursor].(map[string]interface{})
    cell := rowData["row"].([]interface{})[res.columns[column]]
    return cell
}


