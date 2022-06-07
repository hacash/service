package main

import (
	"fmt"
	"github.com/hacash/core/sys"
	"os"
	"os/signal"
)

/*
export GOPATH=/media/yangjie/500GB/hacash/go
go build -o test/ranking1 github.com/hacash/service/ranking && ./test/ranking1 rankingt1.config.ini
*/

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	target_ini_file := "./ranking.config.ini"
	if len(os.Args) >= 2 {
		target_ini_file = os.Args[1]
	}

	target_ini_file = sys.AbsDir(target_ini_file)

	// start-up
	hinicnf, err := sys.LoadInicnf(target_ini_file)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	rank := NewRanking(hinicnf)
	rank.Start()

	s := <-c
	fmt.Println("Got signal:", s)
}
