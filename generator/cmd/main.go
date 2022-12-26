package main

import (
	"flag"
	"log"
	"os"

	"github.com/kislerdm/neon-sdk-go/generator"
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

	if err := generator.Run(
		generator.Config{
			OpenAPIReader: nil,
			PathOutput:    outputDir,
		},
	); err != nil {
		log.Fatalln(err)
	}
}
