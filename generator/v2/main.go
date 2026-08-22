package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"generator/generator"
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

	outputFilesCreator, err := newFilesCreator(outputDir)
	if err != nil {
		log.Fatalln(err)
	}

	if err := generator.Run(openAPISpec, outputFilesCreator); err != nil {
		log.Fatalln(err)
	}
}

func newFilesCreator(dirName string) (generator.CreatorFn, error) {
	_, err := os.ReadDir(dirName)
	if err != nil {
		return nil, err
	}
	return func(fileName string) (io.WriteCloser, error) {
		subDirName, _ := path.Split(fileName)
		if subDirName != "" {
			if err := os.MkdirAll(subDirName, 0750); err != nil {
				return nil, fmt.Errorf("could not create subdir %s: %w", subDirName, err)
			}
		}
		return os.Open(fileName)
	}, nil
}
