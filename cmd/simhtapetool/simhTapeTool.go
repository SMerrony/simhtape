// simhtapeTool is a utility for manipulating SimH-encoded images of tapes for AOS/VS
// systems using the simhtape package.

// Copyright (C) 2018,2019  Steve Merrony

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/SMerrony/simhtape/pkg/simhtape"
)

var (
	// verbs
	createFlag = flag.String("create", "", "Create a new SimH Tape Image file")
	dumpFlag   = flag.String("dump", "", "Dump all files in image as blobs in current directory")
	scanFlag   = flag.String("scan", "", "Scan a SimH Tape Image file for correctness")

	// flags/args
	csvFlag            = flag.Bool("csv", false, "Generate CSV-format data from scan")
	fromDefinitionFlag = flag.String("fromDefinition", "", "Use a definition file")
	vFlag              = flag.Bool("v", false, "Be more verbose")
)

func main() {
	flag.Parse()

	switch {
	case *scanFlag != "":
		fmt.Printf("Scanning tape file : %s", *scanFlag)
		fmt.Printf("%s\n", simhtape.ScanImage(*scanFlag, *csvFlag))
	case *createFlag != "":
		if *fromDefinitionFlag == "" {
			log.Fatal("ERROR: Must specify --fromASCIIOctal or --fromDefinition to create new image")
		}
		if *fromDefinitionFlag != "" {
			createImageFromDefinition()
		}
	case *dumpFlag != "":
		fmt.Println("Dumping files...")
		simhtape.DumpFiles(*dumpFlag)
		fmt.Println("...finished.")
	default:
		log.Fatalln("ERROR: Must specify an action - create, dump, or scan.  Use -h for help.")
	}
}

func createImageFromDefinition() {
	defCSVfile, err := os.Open(*fromDefinitionFlag)
	if err != nil {
		log.Fatalf("ERROR: Could not access CSV Definition file %s", *fromDefinitionFlag)
	}
	defer defCSVfile.Close()
	csvReader := csv.NewReader(defCSVfile)
	imgFile, err := os.Create(*createFlag)
	if err != nil {
		log.Fatalf("ERROR: Could not create new image file %s", *createFlag)
	}
	for {
		// read a line from the CSV definition file
		defRec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("ERROR: Could not parse CSV definition file")
		}
		// 1st field of defRec is the src file name, 2nd field is the block size
		thisSrcFile, err := os.Open(defRec[0])
		if err != nil {
			log.Fatalf("ERROR: Could not open input file %s", defRec[0])
		}
		thisBlkSize, err := strconv.Atoi(defRec[1])
		if err != nil {
			log.Fatalf("ERROR: Could not parse block size for input file %s", defRec[0])
		}
		switch thisBlkSize {
		case 2048, 4096, 8192, 16384:
			fmt.Printf("\nAdding file: %s with block size: %d ", defRec[0], thisBlkSize)
			block := make([]byte, thisBlkSize)
			for {
				bytesRead, err := thisSrcFile.Read(block)
				if err != nil && err != io.EOF {
					log.Fatal(err)
				}
				if bytesRead > 0 {
					simhtape.WriteMetaData(imgFile, uint32(bytesRead)) // block header
					if *vFlag {
						fmt.Printf(" Wrote Header value: %d...", uint32(bytesRead))
					}
					ok := simhtape.WriteRecordData(imgFile, block[0:bytesRead]) // block
					if !ok {
						log.Fatal("ERROR: Error writing image file")
					}
					fmt.Printf(".")
					simhtape.WriteMetaData(imgFile, uint32(bytesRead)) // block trailer
					if *vFlag {
						fmt.Printf(" Wrote Trailer value: %d...", uint32(bytesRead))
					}
				}
				if bytesRead == 0 || err == io.EOF { // End of this file
					thisSrcFile.Close()
					simhtape.WriteMetaData(imgFile, simhtape.SimhMtrTmk)
					if *vFlag {
						fmt.Printf(" EOF: Wrote Tape Mark value: %d...", simhtape.SimhMtrTmk)
					}
					break
				}
			} // loop round for next block
		default:
			log.Fatalf("ERROR: Unsupported block size %d for input file %s", thisBlkSize, defRec[0])
		}
	}
	// // old EOT was 3 zero headers...
	// WriteMetaData(imgFile, 0)
	// WriteMetaData(imgFile, 0)

	simhtape.WriteMetaData(imgFile, simhtape.SimhMtrEom)
	if *vFlag {
		fmt.Printf(" EOM: Wrote Tape Mark value: %d...", simhtape.SimhMtrEom)
	}
	imgFile.Close()
	fmt.Printf("\nDone\n")
}
