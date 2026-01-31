package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davesmith10/rice-cli/internal/validate"
	"github.com/spf13/cobra"
)

func validateCmd() *cobra.Command {
	var strict, jsonOutput, quiet bool

	cmd := &cobra.Command{
		Use:   "validate [path]",
		Short: "Validate a bundle or directory",
		Long:  `Validate a bundle or directory against the ricecake specification.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(args[0], strict, jsonOutput, quiet)
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "Enable strict validation (warnings become errors)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "Only output errors")

	return cmd
}

func runValidate(path string, strict, jsonOutput, quiet bool) error {
	validator := validate.New(path, strict)
	report, err := validator.Validate()
	if err != nil {
		return err
	}

	if jsonOutput {
		return outputJSON(report)
	}

	return outputText(report, quiet, strict)
}

func outputJSON(report *validate.Report) error {
	output := map[string]interface{}{
		"path":    report.Path,
		"results": report.Results,
		"errors":  report.Errors,
		"warns":   report.Warns,
		"valid":   report.Errors == 0,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func outputText(report *validate.Report, quiet, strict bool) error {
	fmt.Printf("Validating: %s\n\n", report.Path)

	// Group results by category
	categories := make(map[string][]validate.Result)
	categoryOrder := []string{"Structure", "Manifest", "Audio", "Images", "Security", "Copyright"}

	for _, result := range report.Results {
		categories[result.Category] = append(categories[result.Category], result)
	}

	for _, category := range categoryOrder {
		results := categories[category]
		if len(results) == 0 {
			continue
		}

		passed := 0
		failed := 0
		warnings := 0

		for _, r := range results {
			if r.Passed {
				passed++
			} else if r.Severity == "warning" {
				warnings++
			} else {
				failed++
			}
		}

		total := passed + failed + warnings
		status := "[PASS]"
		if failed > 0 {
			status = "[FAIL]"
		} else if warnings > 0 {
			status = "[WARN]"
		}

		if !quiet || failed > 0 || warnings > 0 {
			fmt.Printf("%s %s checks (%d/%d)\n", status, category, passed, total)

			// Show failures and warnings
			for _, r := range results {
				if !r.Passed {
					severity := "ERROR"
					if r.Severity == "warning" {
						severity = "WARN"
					}
					fmt.Printf("  - [%s] %s: %s\n", severity, r.Check, r.Message)
				}
			}
		}
	}

	fmt.Println()

	errors := report.Errors
	if strict {
		errors += report.Warns
	}

	if errors > 0 {
		fmt.Printf("Validation failed: %d error(s), %d warning(s)\n", report.Errors, report.Warns)
		os.Exit(2)
	} else if report.Warns > 0 {
		fmt.Printf("Validation complete: %d warning(s), 0 errors\n", report.Warns)
		fmt.Println("Bundle is valid for building.")
	} else {
		fmt.Println("Validation complete: 0 warnings, 0 errors")
		fmt.Println("Bundle is valid for building.")
	}

	return nil
}
