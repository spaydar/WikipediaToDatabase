package main

import (
	"os"
	"testing"
)

func TestOne(t *testing.T) {
	db, err := startDatabase()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	service, err := getService()
	if err != nil {
		panic(err)
	}
	infile, err := os.Open("cities3.txt")
	if err != nil {
		panic(err)
	}
	defer infile.Close()
	err = queryAndCache(db, service, infile)
	if err != nil {
		t.Error("Error in queryAndCache:", err)
		return
	}
}

