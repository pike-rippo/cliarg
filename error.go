package cliarg

import (
	"errors"
	"fmt"
)

func continuousSemicolonError() error {
	return errors.New("cliarg: syntax error: continuous semicolon")
}

func noSuchKeyError(key string) error {
	return fmt.Errorf("cliarg: no such key: \"%s\"", key)
}

func conflictFlagNameError(name string) error {
	return fmt.Errorf("cliarg: flag name collision: \"%s\"", name)
}
