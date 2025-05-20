package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "  good  job  ",
			expected: []string{"good", "job"},
		},
		{
			input:    "  very  good job  ",
			expected: []string{"very", "good", "job"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)

		if len(actual) != len(c.expected) {
			t.Errorf("Actual different from expected! %s != %s", actual, c.expected)
			t.FailNow()
		}

		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("Actual different from expected!")
				t.FailNow()
			}
		}
	}
}
