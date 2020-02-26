package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

var debug bool
var progress int

func init() {
	flag.BoolVar(&debug, "d", false, "Debugging output")
	flag.IntVar(&progress, "p", 100, "Report progress every p files")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s filter [args]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(),
`
Runs the given program as a filter on the input. Standard input and output are
expected to be base 64 encoded, one document or record per line. The program is
run for each line, and the output is then re-encoded.

Example:

    $ < test b64map cat > test.cat
    2020/02/24 12:15:29 b64map.go:180: processed 2 documents
    $ diff test test.cat
    $
`)
	}
}

func readDocs(r io.ReadCloser) (ch chan []byte) {
	ch = make(chan []byte)
	go func() {
		buf := bufio.NewReader(r)

		line := make([]byte, 0, 1024)
		for {
			chunk, pfx, err := buf.ReadLine()
			// we got some bytes, accumulate
			if len(chunk) > 0 {
				line = append(line, chunk...)
			}
			// we're done
			if err != nil {
				if debug {
					log.Printf("readDocs: finished (%v)", err)
				}
				if err == io.EOF {
					if len(line) > 0 {
						b := make([]byte, base64.StdEncoding.DecodedLen(len(line)))
						n, err := base64.StdEncoding.Decode(b, line)
						if err != nil {
							log.Fatalf("readDocs: error decoding line (%v)", err)
						}
						ch <- b[:n]
					}
				} else {
					log.Fatalf("error reading line: %v", err)
				}
				close(ch)
				return
			}
			// if we have a complete line, send it
			if !pfx {
				b := make([]byte, base64.StdEncoding.DecodedLen(len(line)))
				n, err := base64.StdEncoding.Decode(b, line)
				if err != nil {
					log.Fatalf("readDocs: error decoding line (%v)", err)
				}
				ch <- b[:n]
				line = make([]byte, 0, 1024)
			}
		}
	}()
	return ch
}

func writeDoc(doc []byte, w io.Writer, prog string, args ...string) (err error) {
	cmd := exec.Command(prog, args...)
	cmdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("error getting command input: %v", err)
	}
	cmdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("error getting command output: %v", err)
	}
	cmderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("error getting command standard error: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatalf("error starting command: %v", err)
	}

	n, err := cmdin.Write(doc)
	if err != nil {
		log.Printf("error writing to command: %v", err)
		return
	}
	if n != len(doc) {
		log.Printf("did not write the expected number of bytes to command %d != %d", n, len(doc))
	}

	if err := cmdin.Close(); err != nil {
		log.Printf("error closing output pipe: %v", err)
		return err
	}

	doc, err = ioutil.ReadAll(cmdout)
	if err != nil {
		log.Printf("error reading from command: %v", err)
		return
	}

	stderr, err := ioutil.ReadAll(cmderr)
	if err != nil {
		log.Printf("error reading standard error: %v", err)
		return
	}
	if _, err := os.Stderr.Write(stderr); err != nil {
		log.Printf("error writing standard error: %v", err)
	}

	if err = cmd.Wait(); err != nil {
		log.Printf("error waiting for command: %v", err)
		return
	}

	// encode and output
	elen := base64.StdEncoding.EncodedLen(len(doc))
	b := make([]byte, elen, elen+1)
	base64.StdEncoding.Encode(b, doc)
	b = append(b, '\n')
	_, err = w.Write(b)
	if err != nil {
		log.Printf("writeDocs: error writing %v lines %v", n, err)
		return
	}

	return
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(-1)
	}

	args := flag.Args()
	ndocs := 0
	start := time.Now()
	for doc := range readDocs(os.Stdin) {
		ndocs += 1
		if err := writeDoc(doc, os.Stdout, args[0], args[1:]...); err != nil {
			log.Fatalf("fatal: %v", err)
		}
		if progress > 0 && ndocs % progress == 0 {
			now := time.Now()
			log.Printf("written %d docs in %s", ndocs, now.Sub(start).String())
		}
	}

	now := time.Now()
	log.Printf("processed %v documents in %s", ndocs, now.Sub(start).String())
}

