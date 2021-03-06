// +build ignore

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"sort"

	"github.com/knq/snaker"
)

const (
	// chromiumSrc is the base chromium source repo location
	chromiumSrc = "https://chromium.googlesource.com/chromium/src"

	// devtoolsHTTPClientCc contains the target_type names.
	devtoolsHTTPClientCc = chromiumSrc + "/+/master/chrome/test/chromedriver/chrome/devtools_http_client.cc?format=TEXT"
)

var (
	flagOut = flag.String("out", "targettype.go", "out file")

	typeAsStringRE = regexp.MustCompile(`type_as_string\s+==\s+"([^"]+)"`)
)

func main() {
	flag.Parse()

	// grab source
	buf, err := grab(devtoolsHTTPClientCc)
	if err != nil {
		log.Fatal(err)
	}

	// find names
	matches := typeAsStringRE.FindAllStringSubmatch(string(buf), -1)
	names := make([]string, len(matches))
	for i, m := range matches {
		names[i] = m[1]
	}
	sort.Strings(names)

	// process names
	var constVals, decodeVals string
	for _, n := range names {
		name := snaker.SnakeToCamelIdentifier(n)
		constVals += fmt.Sprintf("%s TargetType = \"%s\"\n", name, n)
		decodeVals += fmt.Sprintf("case %s:\n*tt=%s\n", name, name)
	}

	err = ioutil.WriteFile(*flagOut, []byte(fmt.Sprintf(targetTypeSrc, constVals, decodeVals)), 0644)
	if err != nil {
		log.Fatal(err)
	}

	err = exec.Command("gofmt", "-w", "-s", *flagOut).Run()
	if err != nil {
		log.Fatal(err)
	}
}

// grab retrieves a file from the chromium source code.
func grab(path string) ([]byte, error) {
	res, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	buf, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return nil, err
	}

	return buf, nil
}

const (
	targetTypeSrc = `package client

// Code generated by gen.go. DO NOT EDIT.

import (
	// "errors"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// TargetType are the types of targets available in Chrome.
type TargetType string

// TargetType values.
const (
%s
)

// String satisfies stringer.
func (tt TargetType) String() string {
	return string(tt)
}

// MarshalEasyJSON satisfies easyjson.Marshaler.
func (tt TargetType) MarshalEasyJSON(out *jwriter.Writer) {
	out.String(string(tt))
}

// MarshalJSON satisfies json.Marshaler.
func (tt TargetType) MarshalJSON() ([]byte, error) {
	return easyjson.Marshal(tt)
}

// UnmarshalEasyJSON satisfies easyjson.Unmarshaler.
func (tt *TargetType) UnmarshalEasyJSON(in *jlexer.Lexer) {
	z := TargetType(in.String())
	switch z {
%s

	default:
		// in.AddError(errors.New("unknown TargetType"))
		*tt = z
	}
}

// UnmarshalJSON satisfies json.Unmarshaler.
func (tt *TargetType) UnmarshalJSON(buf []byte) error {
	return easyjson.Unmarshal(buf, tt)
}
`
)
