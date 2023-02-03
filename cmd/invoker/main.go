package main

import (
	"bytes"
	"flag"
	"github.com/jamesits/goinvoke/utils"
	"github.com/saferwall/pe"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// note: exit code conforms to sysexits.h
func main() {
	var err error

	// check and mitigate arguments
	if len(dllPath) == 0 {
		flag.Usage()
		os.Exit(64)
	}
	if utils.IsImplicitRelativePath(dllPath) {
		// mimic the behavior roughly where LoadLibrary() is called with LOAD_LIBRARY_SEARCH_SYSTEM32 flag set

		system32, err := utils.GetSystemDirectory()
		if err != nil {
			log.Printf("unable to get System32 directory: %v", err)
			os.Exit(72)
		}

		dllPath = filepath.Join(system32, dllPath)
	}

	s, err := os.Stat(dllPath)
	if err != nil {
		log.Printf("\"%s\" not found: %v", s, err)
		os.Exit(66)
	}
	if !s.Mode().IsRegular() {
		log.Printf("\"%s\" is not a file", s)
		os.Exit(66)
	}

	if len(outputType) == 0 {
		outputType = formatPublicType(utils.BaseName(dllPath))
	}

	if len(outputFileName) == 0 {
		outputFileName = filepath.Join(".", strings.ToLower(utils.BaseName(dllPath))+"_dll.go")
	}

	tags := strings.Split(buildTags, ",")

	// parse the package metadata
	packageName, err := getPackageName([]string{"."}, tags)
	if err != nil {
		log.Printf("unable to parse package name: %v", err)
		os.Exit(78)
	}

	d := templateData{
		ImportPath:      "github.com/jamesits/goinvoke",
		SelfPackageName: "goinvoke",
		CommandLine:     os.Args,

		SelfGenerate: selfGenerate,

		DestinationPackageName: packageName,
		TypeName:               outputType,
		DllFileName:            filepath.Base(dllPath),
	}

	// parse the PE header
	peMeta, err := pe.New(dllPath, &pe.Options{})
	if err != nil {
		log.Printf("unable to read the DLL: %v\n", err)
		os.Exit(66)
	}
	err = peMeta.Parse()
	if err != nil {
		log.Printf("unable to parse the DLL: %v\n", err)
		os.Exit(65)
	}

	for _, v := range peMeta.Export.Functions {
		d.Exports = append(d.Exports, export{
			TypeName: formatPublicType(strings.TrimLeft(v.Name, trimPrefix)),
			Function: v.Name,
		})
	}

	var b bytes.Buffer
	err = srcTemplate.Execute(&b, d)
	if err != nil {
		log.Printf("unable to fill the template: %v\n", err)
		os.Exit(70)
	}

	// format the output
	src, err := format.Source(b.Bytes())
	if err != nil {
		log.Printf("unable to format the source code: %v\n", err)
		os.Exit(70)
	}

	// flush to the destination file
	// perm is before umask as per doc (https://pkg.go.dev/os#WriteFile) so we are safe to use 0666 here
	err = os.WriteFile(outputFileName, src, 0666)
	if err != nil {
		log.Printf("unable to write to file \"%s\": %v\n", outputFileName, err)
		os.Exit(73)
	}

	return
}
