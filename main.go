package main

import (
	"embed"
	"fmt"
	"log"
	"os"

	arg "github.com/alexflint/go-arg"
	logalize "github.com/deponian/logalize/src"

	"github.com/aaaton/golem/v4"
	"github.com/aaaton/golem/v4/dicts/en"
)

var (
	version     string = "0.1.0"
	releaseDate string = "2024-05-12"
)

//go:embed builtins/*
var builtins embed.FS

func main() {
	logalize.SetGlobals(version, releaseDate)

	// parse options
	options := logalize.Options{}
	parser, err := logalize.ParseOptions(os.Args[1:], &options)
	switch {
	case err == arg.ErrHelp:
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	case err == arg.ErrVersion:
		fmt.Fprintln(os.Stdout, options.Version())
		os.Exit(0)
	case err != nil:
		parser.Fail(err.Error())
	}

	// build config
	lemmatizer, err := golem.New(en.New())
	if err != nil {
		log.Fatal(err)
	}
	config, err := logalize.InitConfig(options, builtins)
	if err != nil {
		log.Fatal(err)
	}

	// run the app
	err = logalize.Run(os.Stdin, os.Stdout, config, builtins, lemmatizer)
	if err != nil {
		log.Fatal(err)
	}
}
