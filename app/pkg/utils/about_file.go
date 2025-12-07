package utils

import (
	"errors"
	"os"
)

func ReadFileToString(filePath string) (string, error) {
	if filePath == "" {
		return "", errors.New("file path is empty")
	}
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		err = os.WriteFile(filePath, []byte{}, 0644)
		if err != nil {
			return "", err
		}
		return "", nil
	}
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func WriteStringToFile(filePath string, content string) error {
	if filePath == "" {
		return errors.New("file path is empty")
	}
	data := []byte(content)
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
