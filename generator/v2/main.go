package main

import (
	"flag"
	"generator/generator"
	"log"
	"os"
)

func main() {
	var outputDir, inputPath string
	flag.StringVar(&inputPath, "input", "", "path to the input openAPI spec JSON file [required].")
	flag.StringVar(&outputDir, "output", "", "directory to store the output [required].")
	flag.Parse()

	if inputPath == "" || outputDir == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	openAPISpec, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalln("cannot open input file " + inputPath)
	}

	if err := generator.Run(openAPISpec, outputDir); err != nil {
		log.Fatalln(err)
	}
}
