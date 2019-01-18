package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/csv"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/querytest"

	_ "github.com/influxdata/flux/stdlib" // Import the Flux standard library

	"golang.org/x/text/unicode/norm"
)

// TODO(cwolff): This utility is somewhat broken, and could use some improvements.
//   In order to use it, it's necessary to make assertEquals (the Flux function)
//   echo only its "got" tables to its output (and not produce an error when
//   want and got are unequal).

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
		path = args[0]
		fnames = append(fnames, filepath.Join(path, args[1])+".flux")
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

		if err := generateOutput(fnames); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating output: %v\n", err)
			os.Exit(1)
		}
	}
}

func generateOutput(fnames []string) error {
	l := len(fnames)
	for i, fname := range fnames {
		fmt.Printf("\n\n**** (%v/%v) Generating output for %s ****\n\n", i + 1, l, fname)
		if err := doFileOutput(fname); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating output for %v: %v\n", fname, err)
			fmt.Fprintf(os.Stderr, "Continuing\n")
			fmt.Printf("Press <return> to continue\n")
			bufio.NewReader(os.Stdin).ReadString('\n')
		}
	}

	return nil
}

func doFileOutput(fname string) error {
	ff := func(r rune) bool {
		if r == '\r' || r == '\n' {
			return true
		}
		return false
	}

	querytext, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	// Print all of the test below inData and outData...
	lines := strings.FieldsFunc(string(querytext), ff)
	lastQuote := 0
	for i, line := range lines {
		if line == "\"" {
			lastQuote = i
		}
	}
	belowQuote := strings.Join(lines[lastQuote+1:], "\n") + "\n"
	fmt.Printf("Contents of script below inData/outData:\n________\n%v________\n\n",
		belowQuote)

	pqs := querytest.NewQuerier()
	c := lang.FluxCompiler{
		Query: string(querytext),
	}
	d := csv.DefaultDialect()

	var buf bytes.Buffer
	_, err = pqs.Query(context.Background(), &buf, c, d)
	if err != nil {
		return err
	}

	gotOutData := buf.String()
	wantOutData, err := getOutData(fname)
	if err != nil {
		return err
	}

	wantOutData = strings.Trim(wantOutData, "\n\r")
	gotOutData = strings.Trim(gotOutData, "\n\r")
	ws := strings.FieldsFunc(wantOutData, ff)
	gs := strings.FieldsFunc(gotOutData, ff)
	if cmp.Equal(ws, gs) {
		fmt.Printf("want/got are equal; nothing to do.\n")
		return nil
	}

	fmt.Printf("want/got not equal; -want/+got:\n%v\n\n", cmp.Diff(ws, gs))

	fmt.Print("Update outData in .flux script? (y/n):\n")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.Trim(text, "\n\r")
	if text != "y" {
		// TODO(cwolff): would be nice to give the user the opportunity to update the
		//   script and rerun it here.
		fmt.Printf("Ok, try again after you've fixed it.")
		return nil
	}

	newOutData := strings.Join(gs, "\n")
	if err := replaceOutData(fname, newOutData); err != nil {
		return err
	}

	return nil
}

func getOutData(fluxFile string) (string, error) {

	file, err := os.Open(fluxFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var outData []string
	inOutData := false

	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()
		if !inOutData {
			if line == "outData = \"" {
				inOutData = true
			}

		} else {
			if line != "\"" {
				outData = append(outData, line)
			} else {
				break
			}
		}
	}

	return strings.Join(outData, "\n"), nil
}

func replaceOutData(fname, newOutData string) error {
	file, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	var outFile []string
	inOutData := false
	replaced := false

	scanner := bufio.NewScanner(file)
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()
		if !inOutData {
			outFile = append(outFile, line)
			if line == "outData = \"" {
				if replaced {
					return errors.New("multiple outData")
				}
				inOutData = true
			}
		} else if line == "\"" {
			outFile = append(outFile, newOutData, line)
			inOutData = false
			replaced = true
		}
	}

	outStr := strings.Join(outFile, "\n") + "\n"

	fi, err := os.Stat(fname)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(fname, []byte(outStr), fi.Mode()); err != nil {
		return err
	}

	fmt.Printf("Updated outData for %v.\n", fname)
	return nil
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
