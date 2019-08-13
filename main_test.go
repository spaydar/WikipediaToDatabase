package main

import (
	"testing"
)

func TestOne(t *testing.T) {
	db, err := startDatabase()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = queryAndCache(db,"cities3.txt")
	if err != nil {
		t.Error("Error in main function:", err)
		return
	}
}

