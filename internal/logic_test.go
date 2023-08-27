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
	t.Fail()

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

	for i, v := range expected_lines {
		if(strings.TrimSpace(v) != strings.TrimSpace(actual_lines[i])) {
			t.Errorf("Expected: %s | Actual: %s", v, actual_lines[i])
		}
	}
}
