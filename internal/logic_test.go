package internal

import (
	"io/ioutil"
	"strings"
	"testing"
)

func Test_Listing_0041(t *testing.T) {
	actual := GetASMFromFile("../asm/listing_0041_add_sub_cmp_jnz")
	actual_lines := strings.Split(actual, "\n")

	expected_bytes, err := ioutil.ReadFile("./testdata/test_listing_0041.asm")
	if err != nil {
		t.Fatalf("Failed to open testdata file: %s", err.Error())
	}

	expected := string(expected_bytes)
	expected_lines := strings.Split(expected, "\n")

	if len(expected_lines) != len(actual_lines) {
		t.Fatalf("Line counts not equal. Expected: %d, Actual: %d", len(expected_lines), len(actual_lines))
	}

	for i, v := range expected_lines {
		if(v != actual_lines[i]) {
			t.Errorf("Expected: %s, Actual: %s", v, actual_lines[i])
		}
	}
}
