package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/benjivesterby/graph"

	"github.com/pkg/errors"
)

const fileExtension = ".graph"
const linedelim = '\n'
const directed = "directed"
const weighted = "weighted"

func main() {
	var err error

	filePath := flag.String("file", "", "The path to the file to load the graph from")
	flag.Parse()

	if len(*filePath) > 0 {
		path := *filePath

		var file *os.File
		var reader *bufio.Reader
		if file, reader, err = loadFile(path); err == nil {
			defer file.Close()

			var line string
			if line, err = reader.ReadString(linedelim); err == nil {

				var graphy *graph.Graphy
				if graphy, err = buildGraph(line); err == nil {

					index := 1
					for err == nil {
						if line, err = reader.ReadString(linedelim); err == nil {

							// Build the graph using the line from the file
							err = buildNode(graphy, line)
						}

						index++
					}

					if err == nil || err == io.EOF {
						// TODO: print out the graph here

					} else {
						log.Fatalf("error at line [%v]: [%s]", index, err.Error())
					}
				} else {
					log.Fatalf("error in file [%s]: [%s]", path, err.Error())
				}
			} else {
				log.Fatalf("error while reading line from file [%s]", err.Error())
			}
		} else {
			log.Fatalf("unable to open file [%s]: [%s]", path, err.Error())
		}
	} else {
		log.Fatalf("--file path argument is required")
	}
}

func loadFile(path string) (file *os.File, reader *bufio.Reader, err error) {
	var valid bool
	if valid, err = validFile(path); valid && err == nil {

		var extension = filepath.Ext(path)
		if fileExtension == extension {

			// Load the file
			if file, err = os.Open(path); err == nil {

				reader = bufio.NewReader(file)
			} else {
				err = errors.Errorf("error while reading from file [%s] : [%s]", path, err.Error())
			}
		} else {
			err = errors.Errorf("[%s] is an invalid file extension, expected [%s]", extension, fileExtension)
		}
	} else {

		var errorText string
		if err != nil {
			errorText = err.Error()
		}

		err = errors.Errorf("error occurred while validating file: [%s]", errorText)
	}

	return file, reader, err
}

func buildGraph(line string) (graphy *graph.Graphy, err error) {

	line = clean(line)

	// Parse file
	if len(line) > 0 {
		// Parse out the header of the file
		values := strings.Split(line, " ")

		// The header should only be two fields
		if len(values) == 2 {
			wedgy := false
			dedgy := false

			if values[0] == directed {
				dedgy = true
			}

			if values[1] == weighted {
				wedgy = true
			}

			graphy = &graph.Graphy{
				Directional: dedgy,
				Weighted:    wedgy,
			}
		} else {
			err = errors.New("file is empty")
		}
	} else {
		err = errors.New("file is empty")
	}

	return graphy, err
}

func buildNode(graphy *graph.Graphy, line string) (err error) {
	clean(line)

	if len(line) > 0 {
		values := strings.Split(line, "=")
		if len(values) <= 3 {
			if graphy.Weighted && len(values) == 3 {
				// Weighted Graph

			} else if !graphy.Weighted && len(values) == 2 {
				// Unweighted Graph

			} else if len(values) == 1 {
				// Node w/out edges
				graphy.AddNode(values[0])
			} else {
				err = errors.Errorf("line [%s] is malformed for the graph type", line)
			}
		} else {
			err = errors.Errorf("line [%s] does not contain enough buckets, expects 2-3", line)
		}
	} else {
		err = errors.New("line is empty")
	}

	return err
}

// Clean up the line text and remove unwanted characters
func clean(value string) string {

	value = strings.ReplaceAll(value, "\n", "")
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.TrimSpace(value)

	return value
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
