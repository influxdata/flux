package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/querytest"

	_ "github.com/influxdata/flux/stdlib" // Import the Flux standard library

	"golang.org/x/text/unicode/norm"
)

func init() {
	flux.FinalizeBuiltIns()
}

func normalizeString(s string) []byte {
	result := norm.NFC.String(strings.TrimSpace(s))
	re := regexp.MustCompile(`\r?\n`)
	return []byte(re.ReplaceAllString(result, "\r\n"))
}

func printUsage() {
	fmt.Println("usage: prepcsvtests /path/to/testfiles [testname]")
}

func main() {
	fnames := make([]string, 0)
	path := ""
	var err error
	args := os.Args[1:]
	embed := false
	if args[0] == "embed" {
		embed = true
		args = args[1:]
	}
	if len(args) == 2 {
		path = args[1]
		fnames = append(fnames, filepath.Join(path, os.Args[2])+".flux")
	} else if len(args) == 1 {
		path = args[0]
		fnames, err = filepath.Glob(filepath.Join(path, "*.flux"))
		if err != nil {
			return
		}

		if len(fnames) == 0 {
			fmt.Printf("could not find any .flux files in directory \"%s\"", path)
			return
		}
	} else {
		printUsage()
		return
	}

	if embed {
		embedCSV(fnames)
	} else {

		generateOutput(fnames)
	}
}

func generateOutput(fnames []string) {
	for _, fname := range fnames {
		ext := ".flux"
		testName := fname[0 : len(fname)-len(ext)]
		incsv := testName + ".in.csv"
		indata, err := ioutil.ReadFile(incsv)
		if err != nil {
			fmt.Printf("could not open file %s", incsv)
			return
		}

		fmt.Printf("Generating output for test case %s\n", testName)

		indata = normalizeString(string(indata))
		fmt.Println("Writing input data to file")
		ioutil.WriteFile(incsv, indata, 0644)

		querytext, err := ioutil.ReadFile(fname)
		if err != nil {
			fmt.Printf("error reading query text: %s", err)
			return
		}

		pqs := querytest.NewQuerier()
		c := lang.FluxCompiler{
			Query: string(querytext),
		}
		d := csv.DefaultDialect()

		var buf bytes.Buffer
		_, err = pqs.Query(context.Background(), &buf, c, d)
		if err != nil {
			fmt.Printf("error: %s", err)
			return
		}

		fmt.Printf("FLUX:\n %s\n\n", querytext)
		fmt.Printf("CHECK RESULT:\n%s\n____________________________________________________________", buf.String())

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Results ok (y/n)?: ")
		text, _ := reader.ReadString('\n')
		if text == "y\n" {
			fmt.Printf("writing output file: %s", testName+".out.csv")
			ioutil.WriteFile(testName+".out.csv", buf.Bytes(), 0644)
		}
	}
}

func embedCSV(fnames []string) {

	for _, fname := range fnames {
		ext := ".flux"
		testName := fname[0 : len(fname)-len(ext)]
		incsv := testName + ".in.csv"
		indata, err := ioutil.ReadFile(incsv)
		if err != nil {
			fmt.Printf("could not open file %s", incsv)
			return
		}

		inDataStr := string(normalizeString(string(indata)))

		outcsv := testName + ".out.csv"
		outdata, err := ioutil.ReadFile(outcsv)
		if err != nil {
			fmt.Printf("could not open file %s", outcsv)
			return
		}

		outDataStr := string(normalizeString(string(outdata)))

		querytext, err := ioutil.ReadFile(fname)
		if err != nil {
			fmt.Printf("error reading query text: %s", err)
			return
		}

		querystr := strings.Replace(string(querytext), `"`+incsv+`"`, "inData", -1)
		querystr = strings.Replace(querystr, `"`+outcsv+`"`, "outData", -1)

		if querystr == string(querytext) {
			fmt.Printf("file %s does not reference corresponding data files.\n", fname)
			continue
		}

		newQueryText := "import \"testing\"\n\ninData = \n\"\n" + inDataStr + "\n\"\noutData = \n\"" + outDataStr + "\n\"\n\n" + querystr
		fmt.Println(newQueryText)
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Embed CSV in %s (y/n)?: ", fname)
		text, _ := reader.ReadString('\n')
		if text == "y\n" {
			fmt.Printf("writing output file: %s\n", testName+".out.csv")
			ioutil.WriteFile(testName+".flux", []byte(newQueryText), 0644)
		}
	}
}
