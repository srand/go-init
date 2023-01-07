package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Properties map[string]string

func ReadPropertiesFile(filename string) (Properties, error) {
	config := Properties{}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineno := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineno++

		if strings.HasPrefix(line, "#") {
			continue
		}

		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				config[key] = value
			} else {
				return nil, fmt.Errorf("error:%d: key missing when parsing '%s'", lineno, filename)
			}
		} else {
			return nil, fmt.Errorf("error:%d: expected '=' when parsing '%s'", lineno, filename)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}
