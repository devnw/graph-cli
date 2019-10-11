package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const fileExtension = ".graph"

func main() {
	var err error

	filePath := flag.String("file", "", "The path to the file to load the graph from")
	flag.Parse()

	var valid bool
	if valid, err = validFile(*filePath); valid && err == nil {

		var extension = filepath.Ext(*filePath)
		if fileExtension == extension {

			log.Printf("[%s] is a valid file", *filePath)
			// TODO: Parse arguments

			// TODO: Load File

			// TODO: Parse file

			// TODO: Build graph
		} else {
			log.Fatalf("[%s] is an invalid file extension, expected [%s]", extension, fileExtension)
		}
	} else {

		var errorText string
		if err != nil {
			errorText = err.Error()
		}

		log.Fatalf("error occurred while validating file: [%s]", errorText)
	}
}

func validFile(path string) (valid bool, err error) {
	var file os.FileInfo

	if len(path) > 0 {
		if file, err = os.Stat(path); err == nil {
			if file != nil {
				if !file.IsDir() {
					valid = true
				} else {
					err = errors.Errorf("path [%s] is a directory", path)
				}
			} else {
				err = errors.New("unable to load file stats, file returned nil")
			}
		}
	} else {
		err = errors.New("file path cannot be empty")
	}

	return valid, err
}
