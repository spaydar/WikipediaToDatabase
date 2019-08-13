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
	err = queryAndCache(db, service, infile, "Cities3", "urls", "errors")
	if err != nil {
		t.Error("Error in queryAndCache:", err)
		return
	}
}

func Test3Cities1Error(t *testing.T) {
	db, err := startDatabase()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	service, err := getService()
	if err != nil {
		panic(err)
	}
	infile, err := os.Open("3cities1error.txt")
	if err != nil {
		panic(err)
	}
	defer infile.Close()
	err = queryAndCache(db, service, infile, "3Cities1Error", "urls1", "errors1")
	if err != nil {
		t.Error("Error in queryAndCache:", err)
		return
	}
}

