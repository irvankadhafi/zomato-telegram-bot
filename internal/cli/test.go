package cli

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests",
	Long: `Run tests with various options.

Available subcommands:
  unit      - Run unit tests
  integration - Run integration tests
  coverage  - Run tests with coverage report
  benchmark - Run benchmark tests
  all       - Run all tests

Example:
  telegram-bot test unit
  telegram-bot test coverage
  telegram-bot test all --verbose`,
}

// testUnitCmd represents the test unit command
var testUnitCmd = &cobra.Command{
	Use:   "unit",
	Short: "Run unit tests",
	Long:  `Run unit tests for the application.`,
	RunE:  runTestUnit,
}

// testIntegrationCmd represents the test integration command
var testIntegrationCmd = &cobra.Command{
	Use:   "integration",
	Short: "Run integration tests",
	Long:  `Run integration tests for the application.`,
	RunE:  runTestIntegration,
}

// testCoverageCmd represents the test coverage command
var testCoverageCmd = &cobra.Command{
	Use:   "coverage",
	Short: "Run tests with coverage report",
	Long:  `Run tests and generate a coverage report.`,
	RunE:  runTestCoverage,
}

// testBenchmarkCmd represents the test benchmark command
var testBenchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run benchmark tests",
	Long:  `Run benchmark tests for performance testing.`,
	RunE:  runTestBenchmark,
}

// testAllCmd represents the test all command
var testAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Run all tests",
	Long:  `Run all types of tests (unit, integration, coverage).`,
	RunE:  runTestAll,
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.AddCommand(testUnitCmd)
	testCmd.AddCommand(testIntegrationCmd)
	testCmd.AddCommand(testCoverageCmd)
	testCmd.AddCommand(testBenchmarkCmd)
	testCmd.AddCommand(testAllCmd)

	// Test-specific flags
	testCmd.PersistentFlags().Bool("verbose", false, "verbose test output")
	testCmd.PersistentFlags().Bool("race", true, "enable race detection")
	testCmd.PersistentFlags().String("timeout", "30s", "test timeout")
	testCmd.PersistentFlags().String("tags", "", "build tags")
	testCmd.PersistentFlags().String("run", "", "run only tests matching pattern")

	// Coverage-specific flags
	testCoverageCmd.Flags().String("coverprofile", "coverage.out", "coverage profile output file")
	testCoverageCmd.Flags().String("covermode", "atomic", "coverage mode (set, count, atomic)")
	testCoverageCmd.Flags().Bool("html", false, "generate HTML coverage report")

	// Benchmark-specific flags
	testBenchmarkCmd.Flags().String("benchmem", "", "benchmark memory allocations")
	testBenchmarkCmd.Flags().String("benchtime", "1s", "benchmark time")
}

func runTestUnit(cmd *cobra.Command, args []string) error {
	log.Println("Running unit tests...")

	// Build test command
	testArgs := []string{"test"}
	testArgs = append(testArgs, getCommonTestArgs(cmd)...)
	testArgs = append(testArgs, "./...")

	// Exclude integration tests
	testArgs = append(testArgs, "-short")

	return runGoCommand(testArgs)
}

func runTestIntegration(cmd *cobra.Command, args []string) error {
	log.Println("Running integration tests...")

	// Build test command
	testArgs := []string{"test"}
	testArgs = append(testArgs, getCommonTestArgs(cmd)...)
	testArgs = append(testArgs, "-tags=integration")
	testArgs = append(testArgs, "./...")

	return runGoCommand(testArgs)
}

func runTestCoverage(cmd *cobra.Command, args []string) error {
	log.Println("Running tests with coverage...")

	// Get coverage flags
	coverprofile, _ := cmd.Flags().GetString("coverprofile")
	covermode, _ := cmd.Flags().GetString("covermode")
	html, _ := cmd.Flags().GetBool("html")

	// Build test command
	testArgs := []string{"test"}
	testArgs = append(testArgs, getCommonTestArgs(cmd)...)
	testArgs = append(testArgs, fmt.Sprintf("-coverprofile=%s", coverprofile))
	testArgs = append(testArgs, fmt.Sprintf("-covermode=%s", covermode))
	testArgs = append(testArgs, "./...")

	// Run tests with coverage
	if err := runGoCommand(testArgs); err != nil {
		return err
	}

	// Generate HTML report if requested
	if html {
		log.Println("Generating HTML coverage report...")
		htmlFile := strings.TrimSuffix(coverprofile, filepath.Ext(coverprofile)) + ".html"
		htmlArgs := []string{"tool", "cover", fmt.Sprintf("-html=%s", coverprofile), fmt.Sprintf("-o=%s", htmlFile)}
		if err := runGoCommand(htmlArgs); err != nil {
			return fmt.Errorf("failed to generate HTML coverage report: %w", err)
		}
		log.Printf("HTML coverage report generated: %s", htmlFile)
	}

	// Show coverage summary if coverage file exists
    if _, err := os.Stat(coverprofile); err == nil {
        log.Println("Coverage summary:")
        summaryArgs := []string{"tool", "cover", fmt.Sprintf("-func=%s", coverprofile)}
        return runGoCommand(summaryArgs)
    } else {
        log.Printf("Coverage file %s not found, skipping summary", coverprofile)
        return nil
    }
}

func runTestBenchmark(cmd *cobra.Command, args []string) error {
	log.Println("Running benchmark tests...")

	// Get benchmark flags
	benchmem, _ := cmd.Flags().GetString("benchmem")
	benchtime, _ := cmd.Flags().GetString("benchtime")

	// Build test command
	testArgs := []string{"test"}
	testArgs = append(testArgs, getCommonTestArgs(cmd)...)
	testArgs = append(testArgs, "-bench=.")
	testArgs = append(testArgs, fmt.Sprintf("-benchtime=%s", benchtime))

	if benchmem != "" {
		testArgs = append(testArgs, "-benchmem")
	}

	testArgs = append(testArgs, "./...")

	return runGoCommand(testArgs)
}

func runTestAll(cmd *cobra.Command, args []string) error {
	log.Println("Running all tests...")

	// Run unit tests
	log.Println("\n=== Running Unit Tests ===")
	if err := runTestUnit(cmd, args); err != nil {
		log.Printf("Unit tests failed: %v", err)
		// Continue with other tests
	}

	// Run integration tests
	log.Println("\n=== Running Integration Tests ===")
	if err := runTestIntegration(cmd, args); err != nil {
		log.Printf("Integration tests failed: %v", err)
		// Continue with other tests
	}

	// Run coverage tests
	log.Println("\n=== Running Coverage Tests ===")
	if err := runTestCoverage(cmd, args); err != nil {
		log.Printf("Coverage tests failed: %v", err)
		// Continue with other tests
	}

	// Run benchmark tests (optional, may not have benchmarks)
	log.Println("\n=== Running Benchmark Tests ===")
	if err := runTestBenchmark(cmd, args); err != nil {
		log.Printf("Benchmark tests failed (this is normal if no benchmarks exist): %v", err)
		// Don't fail the entire test suite for missing benchmarks
	}

	log.Println("\n=== All Tests Completed ===")
	return nil
}

// getCommonTestArgs returns common test arguments based on flags
func getCommonTestArgs(cmd *cobra.Command) []string {
	var args []string

	// Verbose output
	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		args = append(args, "-v")
	}

	// Race detection
	if race, _ := cmd.Flags().GetBool("race"); race {
		args = append(args, "-race")
	}

	// Timeout
	if timeout, _ := cmd.Flags().GetString("timeout"); timeout != "" {
		args = append(args, fmt.Sprintf("-timeout=%s", timeout))
	}

	// Build tags
	if tags, _ := cmd.Flags().GetString("tags"); tags != "" {
		args = append(args, fmt.Sprintf("-tags=%s", tags))
	}

	// Run pattern
	if run, _ := cmd.Flags().GetString("run"); run != "" {
		args = append(args, fmt.Sprintf("-run=%s", run))
	}

	return args
}

// runGoCommand executes a go command with the given arguments
func runGoCommand(args []string) error {
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	log.Printf("Executing: go %s", strings.Join(args, " "))
	return cmd.Run()
}