package cliarg

import (
	"errors"
	"fmt"
)

func noSuchKeyError(key string) error {
	return fmt.Errorf("no such key: %s", key)
}

func conflictFlagNameError() error {
	return errors.New("flag name collision")
}
