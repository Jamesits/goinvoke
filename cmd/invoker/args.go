package main

import "flag"

var (
	dllPath        string
	outputType     string
	outputFileName string
	trimPrefix     string
	buildTags      string
	selfGenerate   bool
)

func init() {
	flag.StringVar(&dllPath, "dll", "", "path to the DLL")
	flag.StringVar(&outputType, "type", "", "type name; default <Dllname>")
	flag.StringVar(&outputFileName, "output", "", "output file name; default srcdir/<type>_dll.go")
	flag.StringVar(&trimPrefix, "trimprefix", "", "trim the `prefix` from the generated method names")
	flag.StringVar(&buildTags, "tags", "", "comma-separated list of build tags to apply")
	flag.BoolVar(&selfGenerate, "generate", false, "generate a go:generate directive in the output file, so future `go generate`s will update the file; path to `invoker` must have no space in it")
	flag.Parse()
}
