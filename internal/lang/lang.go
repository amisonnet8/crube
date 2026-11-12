// Package lang decides which structure file to read and write from the target
// language: --lang of the CLI or lang of an MCP tool, else CRUBE_LANG
// (docs/design/data-model.md "Languages: one independent file per language").
//
// Each language has a file of its own: structure.txt for English (the
// default) and structure_<tag>.txt for any other language. Exactly one file is
// used at a time. The tag is used as it is given: there is no fallback from
// ja-JP to ja and no change of case.
package lang

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// EnvVar is the environment variable that names the target language.
const EnvVar = "CRUBE_LANG"

// DefaultFile is the structure file of the default language, English.
const DefaultFile = "structure.txt"

// tagPattern is the shape of a language tag (BCP 47, such as ja or zh-CN):
// subtags of 1 to 8 letters or digits joined by "-". It also keeps the file
// name inside .crube/.
var tagPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,8}(-[A-Za-z0-9]{1,8})*$`)

// IsDefault reports whether tag means the default language: unset (empty) or
// "en".
func IsDefault(tag string) bool {
	return tag == "" || tag == "en"
}

// Check reports whether tag can be used as the value of CRUBE_LANG.
func Check(tag string) error {
	return check(EnvVar, tag)
}

func check(name, tag string) error {
	if IsDefault(tag) || tagPattern.MatchString(tag) {
		return nil
	}
	return fmt.Errorf("invalid %s %q (use a language tag such as ja or zh-CN)", name, tag)
}

// Pick decides the target language: arg, the language given with the call
// (--lang of the CLI, lang of an MCP tool; named name in errors), when it is
// not empty, else env, the value of CRUBE_LANG. Either way the tag is
// checked, and "" means the default language.
func Pick(arg, name, env string) (string, error) {
	if arg != "" {
		return arg, check(name, arg)
	}
	return env, Check(env)
}

// Tags lists the languages that have a structure file in dir (.crube/):
// "en" first when structure.txt is there, then the tags of the
// structure_<tag>.txt files in name order. Other files are left out.
func Tags(dir string) ([]string, error) {
	list, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var tags []string
	for _, de := range list {
		name := de.Name()
		if name == DefaultFile {
			tags = append([]string{"en"}, tags...)
			continue
		}
		rest, prefixed := strings.CutPrefix(name, "structure_")
		tag, suffixed := strings.CutSuffix(rest, ".txt")
		if prefixed && suffixed && !de.IsDir() && tagPattern.MatchString(tag) && !IsDefault(tag) {
			tags = append(tags, tag)
		}
	}
	return tags, nil
}

// FileName returns the name of the structure file of the language tag, in
// .crube/: structure.txt for the default language, structure_<tag>.txt for
// any other. The file may not exist; init uses this name to create it.
func FileName(tag string) (string, error) {
	if err := Check(tag); err != nil {
		return "", err
	}
	if IsDefault(tag) {
		return DefaultFile, nil
	}
	return "structure_" + tag + ".txt", nil
}

// Select returns the name of the structure file to read and write for the
// language tag in the .crube/ directory dir: the file of the language if it
// is there as a regular file, and DefaultFile otherwise.
func Select(dir, tag string) (string, error) {
	name, err := FileName(tag)
	if err != nil || name == DefaultFile {
		return name, err
	}
	fi, err := os.Stat(filepath.Join(dir, name))
	switch {
	case err == nil && fi.Mode().IsRegular():
		return name, nil
	case err == nil || errors.Is(err, fs.ErrNotExist):
		return DefaultFile, nil
	default:
		return "", err
	}
}
