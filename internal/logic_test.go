package internal

import (
	"io/ioutil"
	"strings"
	"testing"
)

func Test_Listing_0041_stringBuilder(t *testing.T) {
	expected_bytes, err := ioutil.ReadFile("./testdata/test_listing_0041.asm")
	if err != nil {
		t.Fatalf("Failed to open testdata file: %s", err.Error())
	}
	expectedStr := string(expected_bytes)

	actualStr, _ := GetASMFromFile("../asm/listing_0041_add_sub_cmp_jnz")

	compareLines(t, expectedStr, actualStr)
}

func Test_Listing_0041_InstructionPrinter(t *testing.T) {
	expected_bytes, err := ioutil.ReadFile("./testdata/test_listing_0041.asm")
	if err != nil {
		t.Fatalf("Failed to open testdata file: %s", err.Error())
	}
	expectedStr := string(expected_bytes)

	_, instructions := GetASMFromFile("../asm/listing_0041_add_sub_cmp_jnz")
	actualStr := PrintInstructions(instructions)

	compareLines(t, expectedStr, actualStr)
}

func compareLines(t *testing.T, expectedStr string, actualStr string) {
	actual_lines := strings.Split(actualStr, "\n")
	expected_lines := strings.Split(expectedStr, "\n")

	if len(expected_lines) != len(actual_lines) {
		t.Fatalf("Line counts not equal. Expected: %d, Actual: %d", len(expected_lines), len(actual_lines))
	}

	var count = 0
	for i, v := range expected_lines {
		expected_line := strings.ReplaceAll(v, ",", "")
		expected_line = strings.TrimSpace(expected_line)
		actual_line   := strings.ReplaceAll(actual_lines[i], ",", "")
		actual_line = strings.TrimSpace(actual_line)

		// didn't want to have weird spacer in the printer logic just to satisfy this test... -- Carver (8-30-23)
		expected_line = strings.ReplaceAll(expected_line, " ... ;", "")
		actual_line = strings.ReplaceAll(actual_line, " ... ;", "")

		if(expected_line != actual_line) {
			count++
			t.Errorf("Count: %d --> Expected: %s | Actual: %s", count, expected_line, actual_line)
		}
	}
}
