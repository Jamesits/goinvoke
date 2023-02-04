package main

import "flag"

var (
	dllPath          string
	outputType       string
	outputFileName   string
	trimPrefix       string
	buildTags        string
	selfGenerate     bool
	preserveRealArg0 bool
)

func init() {
	flag.StringVar(&dllPath, "dll", "", "path to the DLL")
	flag.StringVar(&outputType, "type", "", "type name; default <Dllname>")
	flag.StringVar(&outputFileName, "output", "", "output file name; default srcdir/<type>_dll.go")
	flag.StringVar(&trimPrefix, "trim-prefix", "", "trim the `prefix` from the generated method names")
	flag.StringVar(&buildTags, "tags", "", "comma-separated list of build tags to apply")
	flag.BoolVar(&selfGenerate, "generate", false, "generate a go:generate directive in the output file, so future `go generate`s will update the file; require `invoker` in the PATH")
	flag.BoolVar(&preserveRealArg0, "preserve-arg0", false, "preserve the actual path to `invoker`; will generate machine-specific information and might contain your private information")
	flag.Parse()
}
