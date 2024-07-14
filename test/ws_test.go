package test

import (
	"fmt"
	"github.com/terawatthour/surreal-go"
	"math/rand"
	"testing"
)

type User struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func TestEstablishConnection(t *testing.T) {
	db, err := surreal.Connect("ws://localhost:8000/rpc", &surreal.Options{
		Verbose: true,
		WebSocketOptions: surreal.WebSocketOptions{
			OnDropCallback: func(reason error) {
				fmt.Println("dropped connection", reason)
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	defer db.Close()

	if err := db.Use("test", "test"); err != nil {
		t.Fatalf("unexpected Use error: %s", err)
	}

	if err := db.SignIn(surreal.Map{
		"user": "test",
		"pass": "test",
	}); err != nil {
		t.Fatalf("unexpected SignIn error: %s", err)
	}

	//var created []User
	//_ = db.Create("users", User{
	//	Name: "created-1",
	//}, &created)

	var array []int
	var statusCode int
	if err := db.Query(`return $test; return $test2;`, surreal.Map{
		"test":  []int{1, 2, 3},
		"test2": rand.Intn(213),
	}, &array, &statusCode); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	var users []map[string]any

	if err := db.Select("users", &users); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	fmt.Println(users)
}
