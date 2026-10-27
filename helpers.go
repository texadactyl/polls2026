package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	DirPermissions = 0755
)

// getUtcDate - Get UTC date string, YYYY-MM-DD.
func getUtcDate() string {
	return time.Now().UTC().Format("2006-01-02")
}

// getUtcTime - Get UTC time string in the format of hh:mm:ss.ddd.
func getUtcTime() string {
	return time.Now().UTC().Format("15:04:05.000")
}

// writeOutputText - Write a text line to the given output file handle.
func writeOutputText(outHandle *os.File, textLine string) {
	if _, err := fmt.Fprintln(outHandle, textLine); err != nil {
		outPath, _ := filepath.Abs(filepath.Dir(outHandle.Name()))
		fatalErr(err, "WriteOutputText: fmt.Fprintln(%s) failed", outPath)
	}
}

// makeDir - If the specified directory does not yet exist, create it.
func makeDir(pathDir string) {
	info, err := os.Stat(pathDir)
	if err == nil {
		if !info.IsDir() {
			fatalText("MakeDir: Observed a simple file: %s (expected a directory)", pathDir)
		}
	} else if os.IsNotExist(err) {
		if err := os.Mkdir(pathDir, DirPermissions); err != nil {
			fatalErr(err, "MakeDir: os.Mkdir(%s) failed", pathDir)
		}
	} else {
		fatalErr(err, "MakeDir: os.Stat(%s) failed", pathDir)
	}
}

// storeText - Store a text file in the specified directory.
func storeText(targetDir, argFile, text string) {
	var err error
	fullPath := filepath.Join(targetDir, argFile)
	outHandle, err := os.Create(fullPath)
	if err != nil {
		fatalErr(err, "StoreText: os.Create(%s) failed", fullPath)
	}
	_, err = fmt.Fprintln(outHandle, text)
	if err != nil {
		fatalErr(err, "StoreText: fmt.Fprintln(%s) failed", fullPath)
	}
	err = outHandle.Close()
	if err != nil {
		fatalErr(err, "StoreText: fmt.Close(%s) failed", fullPath)
	}
}

// cleanerText - Replace or remove non-printable characters in a string.
func cleanerText(inString string, newlines bool) string {
	inRunes := []rune(inString)
	var outRunes []rune
	for _, rr := range inRunes {
		if newlines {
			if !unicode.IsPrint(rr) && rr != '\n' && rr != '\r' {
				outRunes = append(outRunes, ' ')
			} else {
				outRunes = append(outRunes, rr)
			}
		} else if unicode.IsPrint(rr) {
			outRunes = append(outRunes, rr)
		}
	}
	return string(outRunes)
}

// isDirectory - Check if the given path exists, accessible, and is a directory. Create the directory if not yet existing.
func isDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		err = os.Mkdir(path, 0755)
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", path)
	}
	return nil
}

// SortCaseInsensitive - Case-insensitive sort.
func sortCaseInsensitive(ptrSlice *[]string) {
	slices.SortFunc(*ptrSlice, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})
}

// urlGet - Download the data contents at a URL to a file, given by outPath.
func urlGet(url, outPath string) {
	resp, err := http.Get(url)
	if err != nil {
		fatalErr(err, "urlGet: http.Get(%s -> %s) failed", url, outPath)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fatalText("urlGet: http request(%s -> %s) failed, status code: %s", url, outPath, resp.Status)
	}

	out, err := os.Create(outPath)
	if err != nil {
		fatalErr(err, "urlGet: os.Create(%s -> %s) failed", url, outPath)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fatalErr(err, "urlGet: io.Copy(%s -> %s) failed", url, outPath)
	}
}

// hashStrArray - Hash a string array, producing a hex string representation of the hash output.
func hashStrArray(strSlice []string) string {
	hasher := md5.New()
	for _, str := range strSlice {
		hasher.Write([]byte(str))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// computeUnknown - Compute the unknown voter percentage.
func computeUnknown(dem, gop, ind float64) float64 {
	return 100.0 - (dem + gop + ind)
}

// stringToFloat64 - Convert a string to a float32.
func stringToFloat64(str string, fieldName string) float64 {
	result, err := strconv.ParseFloat(str, 64)
	if err != nil {
		fatalErr(err, "stringToFloat64: strconv.ParseFloat(input: %s, field: %s) failed", str, fieldName)
	}
	return result
}

// stringToInt64 - Convert a string to an int64.
func stringToInt64(str string, fieldName string) int64 {
	result, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		fatalErr(err, "stringToInt64: strconv.ParseInt(input: %s, field: %s) failed", str, fieldName)
	}
	return result
}

func monthStringToInt64(str string, fieldName string) int64 {
	switch str {
	case "Jan":
		return int64(1)
	case "Feb":
		return int64(2)
	case "Mar":
		return int64(3)
	case "Apr":
		return int64(4)
	case "May":
		return int64(5)
	case "Jun":
		return int64(6)
	case "Jul":
		return int64(7)
	case "Aug":
		return int64(8)
	case "Sep":
		return int64(9)
	case "Oct":
		return int64(10)
	case "Nov":
		return int64(11)
	case "Dec":
		return int64(12)
	default:
		fatalText("monthStringToInt64: unrecognizable month string: %s, field: %s", str, fieldName)
		return -1 // never reached
	}
}
