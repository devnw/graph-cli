package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/benjivesterby/graph"

	"github.com/pkg/errors"
)

const fileExtension = ".gl"
const linedelim = '\n'
const directed = "directed"
const weighted = "weighted"

func main() {
	var err error

	ctx := context.Background()

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

				var input *graph.Graphy
				if input, err = buildGraph(line); err == nil {

					index := 1
					eof := false
					for err == nil && !eof {
						if line, err = reader.ReadString(linedelim); err == nil || err == io.EOF {

							if err == io.EOF {
								eof = true
							}

							// Build the graph using the line from the file
							err = buildNode(input, line)
						}

						index++
					}

					if err == nil || err == io.EOF {
						// TODO: print out the graph here
						fmt.Println(input.String(context.Background()))

						// TODO: Execute Primm's Algorithm Here
						var output *graph.Graphy
						if output, err = prim(ctx, input, "r"); err == nil {
							if output != nil {

								fmt.Println("PRIM Graph")
								fmt.Println(output.String(context.Background()))
							}
						} else {
							// TODO:
						}
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

func prim(ctx context.Context, input *graph.Graphy, r interface{}) (output *graph.Graphy, err error) {

	nodes := input.Nodes(ctx)
	heap := &graph.Heap{}

	// Add all nodes to the heap
	func() {

		for {
			select {
			case <-ctx.Done():
				return
			case n, ok := <-nodes:
				if ok {

					// Set the root node cost to 0
					if n.Value == r {
						n.Cost = 0
					} else {

						// Set infinite value for the float on each node that is not a root node
						n.Cost = float64(^uint(0) >> 1)
					}

					heap.Insert(n)

				} else {
					return
				}
			}

		}
	}()

	output = &graph.Graphy{
		Directional: input.Directional,
		Weighted:    input.Weighted,
	}

	fmt.Println("Adding Root Node to MST")
	var root *graph.Node
	if root, err = heap.ExtractMin(); err == nil {

		output.Node(root.Value)
		processNodeEdges(ctx, root, heap)

		for heap.Size() > 0 {
			// Pop the root node off the heap
			var min *graph.Node
			if min, err = heap.ExtractMin(); err == nil {

				fmt.Printf("Adding %v Node to MST\n", min.Value)

				// using the attachment cost of the edges for the root node update the costs of the
				// nodes still in the heap

				// Process the existing edges from the original graph
				processNodeEdges(ctx, min, heap)

				// Strip the existing edges from this node
				var newNode *graph.Node
				if newNode, err = output.Node(min.Value); err == nil {

					var parent *graph.Node
					if parent, err = output.Node(min.Parent.Value); err == nil {
						output.AddEdge(parent, newNode, nil, min.Cost)
					}
				}

				heap.Print()
				fmt.Println()
			}
		}
	}

	fmt.Println("Input Graph")
	fmt.Println(input.String(ctx))

	fmt.Println("Ouput Graph / MST")
	fmt.Println(output.String(ctx))

	// fmt.Println("Heap")
	// heap.Print()

	return output, err
}

func processNodeEdges(ctx context.Context, node *graph.Node, heap *graph.Heap) {
	edges := node.Edges(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-edges:
			if ok {
				child := e.Child()
				if child == node {
					child = e.Parent()
				}

				if child != node.Parent {

					if wedge, ok := e.(graph.WeightedEdge); ok {
						heap.ChangeCost(child.Value, node, wedge.Weight())
					} else {
						// TODO:
						fmt.Printf("invalid weighted edge %v\n", child.Value)
					}

				}
			} else {
				return
			}
		}
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
	line = clean(line)

	if len(line) > 0 {
		values := strings.Split(line, "=")
		if len(values) <= 3 {
			if len(values) > 1 {
				var parent *graph.Node
				var child *graph.Node

				if len(values[0]) > 0 {
					if parent, err = graphy.Node(values[0]); err == nil {

						if len(values[1]) > 0 {
							if child, err = graphy.Node(values[1]); err == nil {

								var weight float64
								if len(values) == 3 {
									weight, err = strconv.ParseFloat(values[2], 64)
								}

								if err == nil {
									// Add the edge
									err = graphy.AddEdge(parent, child, nil, weight)
								}
							} else {
								err = errors.New("error while loading child node")
							}
						}
					} else {
						err = errors.New("error while loading parent node")
					}
				}
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
