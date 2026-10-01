package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ValidateOutput(output string) bool {
	switch output {
	case "mp4", "mp3", "jpg", "png", "webp":
		return true
	default:
		return false
	}
}

func ValidateFormatID(value string) bool {
	if value == "original" {
		return true
	}
	if value == "" || len(value) > 128 || strings.ContainsAny(value, `/\\\x00\r\n`) {
		return false
	}
	for _, char := range value {
		if !(char == '+' || char == '-' || char == '_' || char == '.' || char >= '0' && char <= '9' || char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z') {
			return false
		}
	}
	return true
}

type DownloadJob struct {
	SelectedFormat string `json:"selected_format"`
	OutputFormat   string `json:"output_format"`
}

func main() {
	jsonData := `{"selected_format":"dash-1430991455627102v","output_format":"mp4"}`
	var job DownloadJob
	if err := json.Unmarshal([]byte(jsonData), &job); err != nil {
		fmt.Println("unmarshal err:", err)
		return
	}
	
	validOutput := ValidateOutput(job.OutputFormat)
	validFormat := ValidateFormatID(job.SelectedFormat)
	
	if !validOutput || !validFormat {
		fmt.Printf("invalid job format. Output: %v, Format: %v\n", validOutput, validFormat)
	} else {
		fmt.Println("Job format is VALID")
	}
}
