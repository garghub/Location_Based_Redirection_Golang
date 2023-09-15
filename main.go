package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

// Below I defined "Rule" struct to represent a map
// that associates a location (e.g., country or region)
// with a specific backend server.
type Rule struct {
	Path      string            `json:"path"`
	Locations map[string]string `json:"locations"`
}

// "RuleSet" struct is defined to hold all the rules
type RuleSet struct {
	Rules []Rule `json:"rules"`
}

// Below defined function "loadRulesFromFile" reads the mapping rules
// from the "mapping_rules.json" file
// Then, I unmarshal the read rules into a "RuleSet" variable.
func loadRulesFromFile(filename string) (*RuleSet, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var ruleSet RuleSet

	// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by ruleSet.
	// If ruleSet is nil or not a pointer, Unmarshal returns an InvalidUnmarshalError.
	err = json.Unmarshal(data, &ruleSet)
	if err != nil {
		return nil, err
	}

	return &ruleSet, nil
}

// Below defined "applyRules" function handles the location-based server selection.
// It takes a request URL as input, retrieve the location information of the client from the incoming request
// and checks if there is a matching rule in the RuleSet.
// If a match is found, it returns the corresponding backend URL
// along with a boolean indicating the match status.
func applyRules(ruleSet *RuleSet, path, clientLocation string) (string, bool) {
	for _, rule := range ruleSet.Rules {
		if rule.Path == path {
			if backend, ok := rule.Locations[clientLocation]; ok {
				return backend, true
			} else {
				return rule.Locations["Default"], true
			}
		}
	}

	return "", false
}

// In an ideal scenario the http.Request object will carry the ipaddress for checking of location
// Unfortunately, since I am executing everything in local machine, the object contains my local ip, i.e., 127.0.0.1
// Therefore, "X-Forwarded-For" is a header variable that I used to pass the ip addresses of various countries to match
func getClientIP(r *http.Request) string {
	// Check for the "X-Forwarded-For" header to handle requests through a proxy
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return strings.Split(forwardedFor, ",")[0]
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}

// Below function is to match the ipaddress of the client and with the help of ip-api.com get the location
func getIPGeolocation(ip string) (string, error) {
	resp, err := resty.New().R().Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("IP geolocation request failed: %s", resp.Status())
	}

	// Parse the response JSON to extract the client's location
	var location struct {
		Country string `json:"country"`
		// Add other location fields as needed (e.g., city, latitude, longitude)
	}

	if err := json.Unmarshal(resp.Body(), &location); err != nil {
		return "", err
	}

	return location.Country, nil
}

// Here, In the main function, I am loading the rules from the file,
// setting up an HTTP server, and defining a request handler function
// that applies the rules to incoming requests.
func main() {
	ruleSet, err := loadRulesFromFile("mapping_rules.json")
	if err != nil {
		fmt.Println("Error loading mapping rules:", err)
		return
	}

	// Here the request URL is extracted and the rules are applied to find the corresponding backend URL.
	// If a match is found, I redirect the request to the backend URL.
	// If not, I return a "Not Found" response (404)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqURL := strings.TrimSuffix(r.URL.Path, "/")

		// Get the client's IP address from the request
		clientIP := getClientIP(r)

		// Use the IP geolocation service to get the client's location
		location, err := getIPGeolocation(clientIP)
		if err != nil {
			http.Error(w, "Error getting location", http.StatusInternalServerError)
			return
		}

		fmt.Println("Current location is found to be of:", location)

		backendURL, found := applyRules(ruleSet, reqURL, location)

		if found {
			http.Redirect(w, r, backendURL, http.StatusFound)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})

	fmt.Println("Proxy server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}
