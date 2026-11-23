package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/acme/obs-manifest-schema/internal/manifest"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 2
	}
	switch args[0] {
	case "validate":
		return runValidate(args[1:])
	case "migrate":
		return runMigrate(args[1:])
	case "version":
		fmt.Printf("obsctl %s\n", version)
		return 0
	case "help", "-h", "--help":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		printUsage()
		return 2
	}
}

func runValidate(args []string) int {
	flags := flag.NewFlagSet("validate", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	warningsAsErrors := flags.Bool("warnings-as-errors", false, "return a non-zero status when naming or deprecation warnings exist")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "validate requires at least one manifest path")
		return 2
	}

	validator, err := manifest.NewValidator()
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialize validator: %v\n", err)
		return 1
	}
	failed := false
	for _, path := range flags.Args() {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: error: %v\n", path, err)
			failed = true
			continue
		}
		report, _ := validator.Validate(data)
		for _, issue := range report.Issues {
			fmt.Fprintf(os.Stderr, "%s: %s\n", path, issue.String())
		}
		if !report.Valid() || (*warningsAsErrors && report.HasWarnings()) {
			failed = true
			continue
		}
		fmt.Printf("%s: valid (%s)\n", path, report.APIVersion)
	}
	if failed {
		return 1
	}
	return 0
}

func runMigrate(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: obsctl migrate v1 v2 [-o output.yaml] observability.yaml")
		return 2
	}
	from, to := args[0], args[1]
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	output := flags.String("o", "", "write migrated YAML to this path; default is stdout")
	if err := flags.Parse(args[2:]); err != nil {
		return 2
	}
	if flags.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "migrate requires exactly one input manifest")
		return 2
	}
	inputPath := flags.Arg(0)
	data, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", inputPath, err)
		return 1
	}
	validator, err := manifest.NewValidator()
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialize validator: %v\n", err)
		return 1
	}
	migrated, err := validator.Migrate(data, from, to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		return 1
	}
	if *output == "" {
		_, _ = os.Stdout.Write(migrated)
		return 0
	}
	if sameFile(inputPath, *output) {
		fmt.Fprintln(os.Stderr, "refusing to overwrite the source file; choose a different -o path")
		return 1
	}
	if err := os.WriteFile(*output, migrated, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *output, err)
		return 1
	}
	fmt.Printf("migrated %s -> %s\n", inputPath, *output)
	return 0
}

func sameFile(first, second string) bool {
	firstAbs, firstErr := filepath.Abs(first)
	secondAbs, secondErr := filepath.Abs(second)
	return firstErr == nil && secondErr == nil && firstAbs == secondAbs
}

func printUsage() {
	_, _ = fmt.Fprintln(os.Stderr, `obsctl validates and migrates ServiceObservability manifests.

Usage:
  obsctl validate [--warnings-as-errors] observability.yaml [more files...]
  obsctl migrate v1 v2 [-o observability-v2.yaml] observability.yaml
  obsctl version`)
}
