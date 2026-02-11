// lingua-cli: A command-line interface for natural language detection using lingua-go.
// This is a Go port of https://github.com/proycon/lingua-cli (Rust/lingua-rs).
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"

	lingua "github.com/pemistahl/lingua-go"
)

const version = "0.2.0"

// isoCodeToLanguage maps an ISO 639-1 code string to a lingua.Language.
// Returns lingua.Unknown and false if the code is not recognized.
func isoCodeToLanguage(code string) (lingua.Language, bool) {
	upper := strings.ToUpper(code)
	for _, lang := range lingua.AllLanguages() {
		if lang.IsoCode639_1().String() == upper {
			return lang, true
		}
	}
	return lingua.Unknown, false
}

// isoCode639_1 returns the lowercase ISO 639-1 code string for a language,
// matching the output format of the Rust lingua-cli.
func isoCode639_1(lang lingua.Language) string {
	return strings.ToLower(lang.IsoCode639_1().String())
}

// formatScore formats a confidence score to match the Rust lingua-cli output:
// exactly 1.0 is printed as "1", all other values use up to 16 significant decimal digits.
func formatScore(score float64) string {
	if score == 1.0 {
		return "1"
	}
	return strconv.FormatFloat(score, 'f', 16, 64)
}

// longEnough returns true if the text contains at least minLength alphabetic characters.
func longEnough(text string, minLength int) bool {
	count := 0
	for _, r := range text {
		if unicode.IsLetter(r) {
			count++
			if count >= minLength {
				return true
			}
		}
	}
	return false
}

// printConfidenceValues prints language detection results to stdout.
// If all is false, only the top result is printed.
// If a confidence threshold is set, results below it are suppressed (printing "unknown" instead).
func printConfidenceValues(
	results []lingua.ConfidenceValue,
	delimiter string,
	confidenceThreshold float64,
	hasThreshold bool,
	all bool,
) {
	found := false
	for _, result := range results {
		score := result.Value()
		if !hasThreshold || score >= confidenceThreshold {
			found = true
			fmt.Printf("%s%s%s\n", isoCode639_1(result.Language()), delimiter, formatScore(score))
		}
		if !all {
			break
		}
	}
	if !found {
		fmt.Printf("unknown%s\n", delimiter)
	}
}

// printLineWithConfidenceValues prints per-line detection results including the original line.
func printLineWithConfidenceValues(
	line string,
	results []lingua.ConfidenceValue,
	delimiter string,
	confidenceThreshold float64,
	hasThreshold bool,
	all bool,
) {
	printed := false
	for _, result := range results {
		score := result.Value()
		if !hasThreshold || score >= confidenceThreshold {
			fmt.Printf("%s%s%s%s%s\n",
				isoCode639_1(result.Language()), delimiter,
				formatScore(score), delimiter,
				line,
			)
			printed = true
		} else {
			fmt.Printf("unknown%s%s%s\n", delimiter, delimiter, line)
			printed = true
		}
		if !all {
			break
		}
	}
	if !printed {
		fmt.Printf("unknown%s%s%s\n", delimiter, delimiter, line)
	}
}

// printWithOffset prints multi-language detection results with byte offsets.
func printWithOffset(results []lingua.DetectionResult, text string, delimiter string) {
	for _, result := range results {
		start := result.StartIndex()
		end := result.EndIndex()
		fragment := text[start:end]
		fmt.Printf("%d%s%d%s%s%s%s\n",
			start, delimiter,
			end, delimiter,
			isoCode639_1(result.Language()), delimiter,
			fragment,
		)
	}
}

func main() {
	// --- flag definitions ---
	languages := flag.String("l", "",
		"Comma seperated list of iso-639-1 codes of languages to detect, if not specified, all supported language will be used. Setting this improves accuracy and resource usage.")
	perLine := flag.Bool("n", false,
		"Classify language per line, this only works if text is not supplied directly as an argument")
	listLangs := flag.Bool("L", false,
		"List all supported languages")
	showAll := flag.Bool("a", false,
		"Show all confidence values (entire probability distribution), rather than just the winning score. Does not work with --multi")
	quick := flag.Bool("q", false,
		"Quick/low accuracy mode")
	multi := flag.Bool("m", false,
		"Classify multiple languages in mixed texts, will return matches along with UTF-8 byte offsets. Can not be combined with line mode.")
	confidenceVal := flag.Float64("c", 0,
		"Confidence threshold, only output results with at least this confidence value (0.0-1.0)")
	minLength := flag.Int("M", 0,
		"Minimum text length (without regard for whitespace, punctuation or numerals!). Shorter fragments will be classified as 'unknown'")
	minRelDist := flag.Float64("d", 0,
		"Minimum relative distance between top language probabilities (0.0-1.0).")
	delimiter := flag.String("D", "\t",
		"Output column delimiter.")
	showVersion := flag.Bool("V", false, "Print version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "lingua-cli is a command line tool for language classification, using the lingua-go library.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: lingua-cli [OPTIONS] [TEXT]...\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n  [TEXT]... \n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Detect whether -c and -d were explicitly provided via flag.Visit
	hasConfidence := false
	hasMinRelDist := false
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "c":
			hasConfidence = true
		case "d":
			hasMinRelDist = true
		}
	})

	if *showVersion {
		fmt.Printf("lingua-cli %s\n", version)
		os.Exit(0)
	}

	// --- list supported languages ---
	if *listLangs {
		all := lingua.AllLanguages()
		sort.Slice(all, func(i, j int) bool {
			return all[i].String() < all[j].String()
		})
		for _, lang := range all {
			fmt.Printf("%s - %s\n", strings.ToLower(lang.IsoCode639_1().String()), lang)
		}
		os.Exit(0)
	}

	// --- build detector ---
	var targetLanguages []lingua.Language
	if *languages != "" {
		for _, code := range strings.Split(*languages, ",") {
			code = strings.TrimSpace(code)
			if code == "" {
				continue
			}
			lang, ok := isoCodeToLanguage(code)
			if !ok {
				fmt.Fprintf(os.Stderr, "error: unknown ISO 639-1 language code: %q\n", code)
				os.Exit(1)
			}
			targetLanguages = append(targetLanguages, lang)
		}
	}

	var builder lingua.LanguageDetectorBuilder
	if len(targetLanguages) == 0 {
		builder = lingua.NewLanguageDetectorBuilder().FromAllLanguages()
	} else {
		builder = lingua.NewLanguageDetectorBuilder().FromLanguages(targetLanguages...)
	}

	if *quick {
		builder = builder.WithLowAccuracyMode()
	}
	if hasMinRelDist {
		builder = builder.WithMinimumRelativeDistance(*minRelDist)
	}

	detector := builder.Build()

	// --- process input ---
	positionalArgs := flag.Args()

	if len(positionalArgs) > 0 {
		// Text supplied as positional arguments
		text := strings.Join(positionalArgs, " ")
		if *minLength > 0 && !longEnough(text, *minLength) {
			fmt.Printf("unknown%s\n", *delimiter)
			return
		}
		if *multi {
			results := detector.DetectMultipleLanguagesOf(text)
			printWithOffset(results, text, *delimiter)
		} else {
			results := detector.ComputeLanguageConfidenceValues(text)
			printConfidenceValues(results, *delimiter, *confidenceVal, hasConfidence, *showAll)
		}
		return
	}

	// Read from stdin
	if *perLine {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if *minLength > 0 && !longEnough(line, *minLength) {
				fmt.Printf("unknown%s%s%s\n", *delimiter, *delimiter, line)
				continue
			}
			results := detector.ComputeLanguageConfidenceValues(line)
			printLineWithConfidenceValues(line, results, *delimiter, *confidenceVal, hasConfidence, *showAll)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
			os.Exit(1)
		}
	} else {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
			os.Exit(1)
		}
		text := string(raw)
		if *minLength > 0 && !longEnough(text, *minLength) {
			return
		}
		if *multi {
			results := detector.DetectMultipleLanguagesOf(text)
			printWithOffset(results, text, *delimiter)
		} else {
			results := detector.ComputeLanguageConfidenceValues(text)
			printConfidenceValues(results, *delimiter, *confidenceVal, hasConfidence, *showAll)
		}
	}
}
