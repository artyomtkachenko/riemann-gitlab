package main

import "testing"

func Test_parseGitlabStatus(t *testing.T) {
	input := []byte(
		`run: logrotate: (pid 20055) 442s; run: log: (pid 5834) 76135s
run: redis: (pid 20084) 441s; run: log: (pid 5563) 76160s

down: nginx: 8s, normally up; run: log: (pid 1056) 1857174s
down: sidekiq: 2s, normally up; run: log: (pid 1055) 1857174s
down: unicorn: 2s, normally up; run: log: (pid 1068) 1857174s
`)

	testData := []struct {
		key              string
		expectedDuration int
	}{
		{"logrotate", 442},
		{"redis", 441},
		{"nginx", 8},
		{"sidekiq", 2},
		{"unicorn", 2},
	}

	result := parseGitlabStatus(input)

	for _, test := range testData {
		if result[test.key].duration != test.expectedDuration {
			t.Errorf("Expected: %d, got: %d", test.expectedDuration, result[test.key].duration)
		}
	}

}

func Test_getDuration(t *testing.T) {
	testData := []struct {
		realData string
		expected int
	}{
		{"442s;", 442},
		{"123s,", 123},
	}

	for _, test := range testData {
		if result := getDuration(test.realData); result != test.expected {
			t.Errorf("Expected: %d, got: %d", test.expected, result)
		}
	}
}
