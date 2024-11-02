package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadDotenv(filename string) (map[string]string, error) {
	envMap := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var multilineKey, multilineValue string
	inMultiline := false

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if !inMultiline {
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if strings.Contains(line, "=") && (strings.Count(line, `"`) == 1 || strings.Count(line, `'`) == 1) {
				parts := strings.SplitN(line, "=", 2)
				multilineKey = strings.TrimSpace(parts[0])
				multilineValue = strings.TrimSpace(parts[1])
				inMultiline = true
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid format on line %d: %s", lineNum, line)
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, `"'`)
			envMap[key] = value
			os.Setenv(key, value)

		} else {
			multilineValue += "\n" + line

			if strings.HasSuffix(multilineValue, `"`) || strings.HasSuffix(multilineValue, `'`) {
				inMultiline = false
				multilineValue = strings.Trim(multilineValue, `"'`)
				envMap[multilineKey] = multilineValue
				os.Setenv(multilineKey, multilineValue)
				multilineKey = ""
				multilineValue = ""
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if inMultiline {
		return nil, fmt.Errorf("unterminated multiline value starting at line %d", lineNum)
	}

	return envMap, nil
}
