package autoconsent

import (
	"encoding/json"
	"os"
)

var Rules AutoConsentRules

func init() {
	data, err := os.ReadFile("rules.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &Rules); err != nil {
		panic(err)
	}
}
