## Unofficial Go driver for SurrealDB

### Install 

```bash
go get github.com/terawatthour/surreal-go
```

### Why? There's already an official Go driver for SurrealDB.

I found the code dodgy and hard to work with. I wanted to write my own
driver that was easier to work with, especially for my own use. Most of the 
code is based on the official driver, but I've made some changes to make it 
more appealing to me.

### Usage

```go
package main

import (
    "fmt"
    "math/rand"
    "github.com/terawatthour/surreal-go"
)
    
func main() {
    // establish a connection to the SurrealDB server
    db, _ := surreal.Connect("ws://localhost:8000/rpc", &surreal.Options{
        Verbose: true,
        WebSocketOptions: surreal.WebSocketOptions{
            // the websocket driver doesn't handle reconnections, but you can do it yourself  
            OnDropCallback: func(reason error) {
                fmt.Println("dropped connection", reason)
            },
        },
    })
    defer db.Close()
    
    // set the desired namespace and database
    _ = db.Use("ns-test", "db-test")
	
    // select desired data into either a map, struct, or a slice 
    var user map[string]any
    _ = db.Select("users:eqxomgmyq9z4lnl1gp65", &user)
}
```