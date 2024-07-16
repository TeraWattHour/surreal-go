package test

import (
	"github.com/terawatthour/surreal-go"
	"testing"
	"time"
)

func connect() *surreal.DB {
	db, err := surreal.Connect("ws://localhost:8000/rpc", nil)
	if err != nil {
		panic(err)
	}

	if err := db.Use("test", "test"); err != nil {
		panic(err)
	}

	_ = db.SignIn(surreal.AuthArgs{
		Namespace: "test",
		Database:  "test",
		Other: surreal.Map{
			"user": "test",
			"pass": "test",
		},
	})

	return db
}

type Article struct {
	ID        string    `json:"id,omitempty"`
	Title     string    `json:"title,omitempty"`
	Content   string    `json:"content,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

func TestInsert(t *testing.T) {
	db := connect()
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	toInsert := Article{
		Title:     "Hello, World!",
		Content:   "This is the first article",
		CreatedAt: time.Now(),
	}

	var inserted Article
	err := db.Insert("article", toInsert, &inserted)
	if err != nil {
		t.Fatal(err)
	}

	if inserted.ID == "" || inserted.Title != toInsert.Title || inserted.Content != toInsert.Content || inserted.CreatedAt.Unix() != toInsert.CreatedAt.Unix() {
		t.Fatalf("Inserted article does not match: %+v %+v", inserted, toInsert)
	}

	batchInsert := []Article{
		{
			Title:     "Hello, World!1",
			Content:   "This is the first article",
			CreatedAt: time.Now(),
		}, {
			Title:     "Hello, World!2",
			Content:   "This is the first article",
			CreatedAt: time.Now(),
		}, {
			Title:     "Hello, World!3",
			Content:   "This is the first article",
			CreatedAt: time.Now(),
		},
	}

	var batchInserted []Article
	err = db.Insert("article", batchInsert, &batchInserted)
	if err != nil {
		t.Fatal(err)
	}

	if len(batchInserted) != len(batchInsert) {
		t.Fatalf("Batch insert failed: %+v %+v", batchInserted, batchInsert)
	}

	for i, inserted := range batchInserted {
		if inserted.ID == "" || inserted.Title != batchInsert[i].Title || inserted.Content != batchInsert[i].Content || inserted.CreatedAt.Unix() != batchInsert[i].CreatedAt.Unix() {
			t.Fatalf("Batch insert failed: %+v %+v", inserted, batchInsert[i])
		}
	}
}

func TestMerge(t *testing.T) {
	db := connect()
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	now := time.Now()

	var merged Article
	err := db.Merge("article", surreal.Map{
		"updatedAt": now,
	}, &merged)
	if err != nil {
		t.Fatal(err)
	}

	if merged.ID == "" || merged.UpdatedAt.Unix() != now.Unix() {
		t.Fatalf("Merged article does not match: %+v %+v", merged.UpdatedAt, now)
	}

	var nonExistent []Article
	err = db.Merge("article:xdd", surreal.Map{
		"updatedAt": now,
	}, &nonExistent)
	if err != nil {
		t.Fatal(err)
	}
	if len(nonExistent) != 1 {
		t.Fatalf("Non-existent merge failed: %+v", nonExistent)
	}

	var nonExistentSingle Article
	err = db.Merge("article:xdd", surreal.Map{
		"updatedAt": now,
	}, &nonExistentSingle)
	if err != nil {
		t.Fatal(err)
	}
	if nonExistentSingle.UpdatedAt.Unix() != now.Unix() {
		t.Fatalf("Non-existent single merge failed: %+v", nonExistent)
	}
}
