package docs

import "testing"

func TestSplitCamelCase(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"AccessControl", "Access Control"},
		{"IAMRole", "IAM Role"},
		{"EC2Instance", "EC2 Instance"},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			got := SplitCamelCase(c.input)
			if got != c.expected {
				t.Errorf("Expected %s, got %s", c.expected, got)
			}
		})
	}
}
