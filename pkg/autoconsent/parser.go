package autoconsent

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var Rules AutoConsentRules

func init() {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(exePath)
	rulesPath := filepath.Join(dir, "rules.json")

	data, err := os.ReadFile(rulesPath)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &Rules); err != nil {
		panic(err)
	}
}
