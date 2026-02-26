package general_util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/webbben/2d-game-engine/logz"
)

// WriteToJSON writes the given data to a JSON file. The outputFilePath you give should include ".json" at the end (assuming you want an extension on the filename).
//
// Note: Here are some things that can't be written to JSON:
//
// - Interfaces
//
// - Non-public properties of a struct (they aren't visible to the function its passed to)
func WriteToJSON(v any, outputFilePath string) error {
	if !filepath.IsAbs(outputFilePath) {
		return fmt.Errorf("given path is not abs (%s); please pass an absolute path", outputFilePath)
	}

	var data []byte

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(outputFilePath, data, 0o644)
}

func GetListOfFiles(directoryPath string, getAbsPaths bool) []string {
	files := []string{}
	fileInfos, err := os.ReadDir(directoryPath)
	if err != nil {
		logz.Panicln("GetListOfFiles", "failed to read files in directory:", err)
	}
	for _, fi := range fileInfos {
		if fi.IsDir() {
			continue
		}
		name := fi.Name()
		if getAbsPaths {
			name = filepath.Join(directoryPath, name)
		}
		files = append(files, name)
	}

	return files
}

func LoadJSONIntoStruct(filePath string, v any) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		logz.Panicln("LoadJSONIntoStruct", "failed to read file data:", err)
	}
	err = json.Unmarshal(b, v)
	if err != nil {
		logz.Panicln("LoadJSONIntoStruct", "failed to unmarshal data:", err)
	}
}
