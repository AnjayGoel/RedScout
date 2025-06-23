package models

import (
	"fmt"
	"regexp"
	"strings"
)

const patternPlaceholder = "{id}"

type Key []string

func (k Key) Pop() (Key, error) {
	if len(k) == 0 {
		return nil, fmt.Errorf("key is empty")
	}

	if len(k) == 1 {
		return Key{}, nil // return empty key if only one part exists
	}

	newKey := make(Key, len(k)-1)
	copy(newKey, k[:len(k)-1])
	return newKey, nil
}

func (k Key) IsEmpty() bool {
	return len(k) == 0 || (len(k) == 1 && k[0] == "")
}

func (k Key) String() string {
	return strings.Join(k, ":")
}

type KeyParser struct {
	delimiter  string
	idPatterns []*regexp.Regexp
}

func NewKeyParser(Delimiter string, IDPattern []*regexp.Regexp) *KeyParser {
	return &KeyParser{
		delimiter:  Delimiter,
		idPatterns: IDPattern,
	}
}

func (kp *KeyParser) matchesPattern(part string) bool {
	for _, regex := range kp.idPatterns {
		if regex.MatchString(part) {
			return true
		}
	}
	return false
}

func (kp *KeyParser) NewKey(s string, inferIds bool) Key {
	if !strings.Contains(s, kp.delimiter) {
		return []string{s}
	}

	parts := strings.Split(s, kp.delimiter)
	if inferIds {
		for i, part := range parts {
			if kp.matchesPattern(part) {
				parts[i] = patternPlaceholder
			}
		}
	}

	return parts
}

func (kp *KeyParser) IsA(k Key, prefix Key) bool {
	if len(prefix) > len(k) {
		return false
	}

	for i := 0; i < len(prefix); i++ {
		if prefix[i] == patternPlaceholder {
			if !kp.matchesPattern(k[i]) {
				return false
			}
		} else if k[i] != prefix[i] {
			return false
		}
	}

	return true
}

func (kp *KeyParser) Namespace(k Key, prefix Key, inferIds bool) (string, error) {
	if len(k) == 0 {
		return "", fmt.Errorf("key is empty")
	}

	if len(prefix) == 0 {
		return k[0], nil
	}

	if !kp.IsA(k, prefix) {
		return "", fmt.Errorf("key %s is not a child of prefix %s", strings.Join(k, ":"), strings.Join(prefix, ":"))
	}

	if len(k) == len(prefix) {
		return "", fmt.Errorf("key %s is exactly the same as prefix %s", strings.Join(k, ":"), strings.Join(prefix, ":"))
	}

	namespace := k[len(prefix)]

	if inferIds && kp.matchesPattern(namespace) {
		namespace = patternPlaceholder
	}
	return namespace, nil
}

func (kp *KeyParser) Append(k Key, part string, inferIds bool) (Key, error) {
	if part == "" {
		return k, fmt.Errorf("key is empty")
	}
	if inferIds && kp.matchesPattern(part) {
		part = patternPlaceholder
	}

	if len(k) == 0 {
		return Key{part}, nil
	}

	newKey := make(Key, len(k)+1)
	copy(newKey, k)
	newKey[len(k)] = part
	return newKey, nil
}
