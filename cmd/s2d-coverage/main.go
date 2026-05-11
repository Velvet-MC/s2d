package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/Velvet-MC/s2d/legacy"
	"github.com/Velvet-MC/s2d/schem"
	_ "github.com/Velvet-MC/s2d/sponge"
	"github.com/Velvet-MC/s2d/translate"
	_ "github.com/df-mc/dragonfly/server/block"
)

type config struct {
	java          bool
	failOnUnknown bool
	top           int
	paths         []string
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseArgs(args []string) (config, error) {
	cfg := config{java: true, top: 40}
	fs := flag.NewFlagSet("s2d-coverage", flag.ContinueOnError)
	fs.BoolVar(&cfg.java, "java", cfg.java, "scan every embedded Java block state")
	fs.BoolVar(&cfg.failOnUnknown, "fail-on-unknown", false, "exit non-zero when unknown states are found")
	fs.IntVar(&cfg.top, "top", cfg.top, "number of unknown states to print per report; -1 prints all")
	fs.Func("schem", "schematic file or directory to scan; may be repeated", func(value string) error {
		cfg.paths = append(cfg.paths, value)
		return nil
	})
	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	cfg.paths = append(cfg.paths, fs.Args()...)
	if cfg.top < -1 {
		return config{}, errors.New("-top must be -1 or greater")
	}
	return cfg, nil
}

func run(cfg config) error {
	unknownTotal := 0
	if cfg.java {
		report, err := translate.JavaStateCoverage()
		if err != nil {
			return err
		}
		unknownTotal += report.Unknown.Total
		printUnknownReport("java", report.TotalStates, report.Unknown, cfg.top)
	}

	if len(cfg.paths) > 0 {
		files, err := collectSchematicFiles(cfg.paths)
		if err != nil {
			return err
		}
		for _, file := range files {
			report, err := scanSchematic(file)
			if err != nil {
				return err
			}
			unknownTotal += report.Unknown.Total
			printUnknownReport(file, report.TotalStates, report.Unknown, cfg.top)
		}
	}

	if cfg.failOnUnknown && unknownTotal > 0 {
		return fmt.Errorf("coverage found %d unknown states", unknownTotal)
	}
	return nil
}

func collectSchematicFiles(paths []string) ([]string, error) {
	seen := map[string]struct{}{}
	var files []string
	for _, path := range paths {
		if path == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if isSchematicPath(path) {
				clean := filepath.Clean(path)
				if _, ok := seen[clean]; !ok {
					seen[clean] = struct{}{}
					files = append(files, clean)
				}
			}
			continue
		}
		err = filepath.WalkDir(path, func(child string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !isSchematicPath(child) {
				return nil
			}
			clean := filepath.Clean(child)
			if _, ok := seen[clean]; ok {
				return nil
			}
			seen[clean] = struct{}{}
			files = append(files, clean)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}

func isSchematicPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".schem", ".schematic":
		return true
	default:
		return false
	}
}

func scanSchematic(path string) (translate.CoverageReport, error) {
	f, err := os.Open(path)
	if err != nil {
		return translate.CoverageReport{}, err
	}
	defer func() { _ = f.Close() }()

	report := translate.CoverageReport{Unknown: translate.UnknownReport{Counts: map[string]int{}}}
	info, err := schem.Scan(path, f, func(schem.Block) error {
		report.TotalStates++
		return nil
	})
	if err != nil {
		return translate.CoverageReport{}, err
	}
	report.Unknown.Total = info.Unknowns.Total
	for state, count := range info.Unknowns.Counts {
		report.Unknown.Counts[state] = count
	}
	return report, nil
}

func printUnknownReport(name string, total int, unknown translate.UnknownReport, top int) {
	fmt.Printf("%s: checked=%d unknown_total=%d unknown_kinds=%d\n", name, total, unknown.Total, len(unknown.Counts))
	for _, row := range unknown.Top(top) {
		fmt.Printf("%8d  %s\n", row.Count, row.State)
	}
}
