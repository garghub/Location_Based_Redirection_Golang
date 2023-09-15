// main_test.go

package main

import (
	"testing"
)

func TestApplyRules(t *testing.T) {

	ruleSet := &RuleSet{
		Rules: []Rule{
			{
				Path: "/search",
				Locations: map[string]string{
					"United States": "https://duckduckgo.com/?q=news",
					"Luxembourg":    "https://www.bing.com/search?q=news",
					"Default":       "https://www.google.com/search?q=news",
				},
			},
		},
	}

	// Test when a matching rule is found
	result, found := applyRules(ruleSet, "/search", "United States")

	// result, found := applyRules(ruleSet, "/search")
	if !found {
		t.Errorf("Expected rule to be found, but not found")
	}
	if result != "https://duckduckgo.com/?q=news" {
		t.Errorf("Expected redirection URL: https://duckduckgo.com/?q=news, got: %s", result)
	}

	// Test when no matching rule is found
	result, found = applyRules(ruleSet, "/unknown/url", "US")
	if found {
		t.Errorf("Expected rule not to be found, but found")
	}
	if result != "" {
		t.Errorf("Expected empty redirection URL, got: %s", result)
	}
}
