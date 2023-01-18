package main

import (
	"testing"
)

func TestCheckAndRetrunArgs(t *testing.T) {
	content := []string{"./testfile", "./somefile", "blowUpString"}
	msg := CheckAndReturnArgs(content)
	if msg != "" {
		t.Fatalf(`CheckAndRetrunArgs() = %q,  want "", error`, msg)
	}
}

func TestCheckAndRetrunArgsTwo(t *testing.T) {
	content := []string{"./testfile", "./somefile.json"}
	msg := CheckAndReturnArgs(content)
	if msg != "./somefile.json" {
		t.Fatalf(`CheckAndRetrunArgs() = %q,  want "./somefile", error`, msg)
	}
}

func TestCheckAndRetrunArgsTres(t *testing.T) {
	content := []string{"./testfile"}
	msg := CheckAndReturnArgs(content)
	if msg != "" {
		t.Fatalf(`CheckAndRetrunArgs() = %q,  want "", error`, msg)
	}
}

func TestCheckAndRetrunArgsFour(t *testing.T) {
	content := []string{"./testfile", "./somefile.php"}
	msg := CheckAndReturnArgs(content)
	if msg != "" {
		t.Fatalf(`CheckAndRetrunArgs() = %q,  want "", error`, msg)
	}
}

func TestCheckSearchAndIgnoreDuplicates(t *testing.T) {
	searchTerms := []string{"foo", "bar", "foo-bar"}
	ignoreTerms := []string{"foo-fee", "feeezzzeee"}
	duplicate := CheckSearchAndIgnoreDuplicates(searchTerms, ignoreTerms, "./SomeLocation")
	if duplicate != false {
		t.Fatalf(`CheckSearchAndIgnoreDuplicates() = %v,  want "false", error`, duplicate)
	}
}

func TestCheckSearchAndIgnoreDuplicatesTwo(t *testing.T) {
	searchTerms := []string{"foo", "bar", "foo-bar"}
	ignoreTerms := []string{"foo", "feeezzzeee"}
	duplicate := CheckSearchAndIgnoreDuplicates(searchTerms, ignoreTerms, "./SomeLocation")
	if duplicate != true {
		t.Fatalf(`CheckSearchAndIgnoreDuplicates() = %v,  want "true", error`, duplicate)
	}
}

func TestReadPlaceHolderFile(t *testing.T) {
	fileLocation := "./foobar"
	value := ReadPlaceHolderFile(fileLocation)
	if value != 0 {
		t.Fatalf(`ReadFile() = %v,  want "0", error`, value)
	}
}

func TestFindMatch(t *testing.T) {
	lineValue := "foo"
	searchTerms := []string{"cheese", "breeze", "gopher"}
	value := FindMatch(lineValue, searchTerms)
	if value != false {
		t.Fatalf(`FindMatch() = %v,  want "false", error`, value)
	}
}

func TestFindMatchTwo(t *testing.T) {
	lineValue := "foo-bar"
	searchTerms := []string{"foo", "who", "gopher"}
	value := FindMatch(lineValue, searchTerms)
	if value != true {
		t.Fatalf(`FindMatch() = %v,  want "true", error`, value)
	}
}

func TestReadFileForMatchesTwo(t *testing.T) {
	searchTerms := []string{"error", "warning", "exclusion"}
	ignoreTerms := []string{"foo", "errors", "gopher"}
	value, _ := ReadFileForMatches("./someFileLocation/someFile", searchTerms, ignoreTerms, 0)
	if len(value) != 1 {
		t.Fatalf(`ReadFileForMatches() = %v,  want "1", error`, len(value))
	}
}

func TestReadFileForMatchesThree(t *testing.T) {
	searchTerms := []string{"error", "warning", "exclusion"}
	ignoreTerms := []string{"error-2345", "warning-46352", "gopher"}
	value, _ := ReadFileForMatches("./someFileLocation/someFile2", searchTerms, ignoreTerms, 0)
	if len(value) != 1 {
		t.Fatalf(`ReadFileForMatches() = %v,  want "1", error`, len(value))
	}
}

func TestFindMatchThree(t *testing.T) {
	lineValue := "test342"
	searchTerms := []string{"cheese", "breeze", "[0-9]"}
	value := FindMatch(lineValue, searchTerms)
	if value != true {
		t.Fatalf(`FindMatch() = %v,  want "true", error`, value)
	}
}

func TestFindMatchFour(t *testing.T) {
	lineValue := "test34/2"
	searchTerms := []string{"cheese", "breeze", "[/]"}
	value := FindMatch(lineValue, searchTerms)
	if value != true {
		t.Fatalf(`FindMatch() = %v,  want "true", error`, value)
	}
}

func TestFindMatchFive(t *testing.T) {
	lineValue := "test342_%Y867"
	searchTerms := []string{"[_%Y]"}
	value := FindMatch(lineValue, searchTerms)
	if value != true {
		t.Fatalf(`FindMatch() = %v,  want "true", error`, value)
	}
}

func TestFindMatchSix(t *testing.T) {
	lineValue := "test%RY&*$#"
	searchTerms := []string{"[0-9]"}
	value := FindMatch(lineValue, searchTerms)
	if value != false {
		t.Fatalf(`FindMatch() = %v,  want "false", error`, value)
	}
}
