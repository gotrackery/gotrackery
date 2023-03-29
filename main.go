/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/gotrackery/gotrackery/cmd"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	cmd.Execute()
}
