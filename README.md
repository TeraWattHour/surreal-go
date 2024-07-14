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
    db, err := surreal.Connect("ws://localhost:8000/rpc", &surreal.Options{
        Verbose: true,
        WebSocketOptions: surreal.WebSocketOptions{
            OnDropCallback: func(reason error) {
                fmt.Println("dropped connection", reason)
            },
        },
    })
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    if err := db.Use("ns-test", "db-test"); err != nil {
        panic(err)
    }
	
    var user map[string]any
    if err := db.Select("users:eqxomgmyq9z4lnl1gp65", &user); err != nil {
        panic(err)
    }
    
    fmt.Println(user)
}
```


