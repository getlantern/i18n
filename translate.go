package i18n

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// T consults the translated text by 'key' from the locale file specified when GetT(),
// then the default locale file, which defaults 'en_US'
// but can be changed by SetDefaultLocale().
// If neither file containes it, will return 'key' surrounded by braces.
type T func(key string) (text string)

// GetT appends ".json" to locale passed in, and loads the file
// under the directory specified by SetLocaleDir()
// or from the file system specified by SetLocaleFS(),
// and returns the translation function.
// It also loads the default locale file which defaults to 'en_US.json'
// but can be changed in advance through SetDefaultLocale().
// The locale file should be json file with only key / text pairs of translations.
func GetT(locale string) (t T, err error) {
	var m, defaultMap trMap
	if m, err = loadMap(locale); err != nil {
		err = fmt.Errorf("Error load locale %s: %s", locale, err)
		return
	}
	if locale == defaultLocale {
		defaultMap = m
	} else if defaultMap, err = loadMap(defaultLocale); err != nil {
		err = fmt.Errorf("Error load default locale %s: %s", locale, err)
		return
	}
	t = func(k string) (s string) {
		if s = m[k]; s == "" {
			if s = defaultMap[k]; s == "" {
				s = "{" + k + "}"
			}
		}
		return
	}
	return
}

// SetDefaultLocale sets the locale to use if T() can't found a translation for a key,
// only affects GetT() after this call.
func SetDefaultLocale(l string) {
	defaultLocale = l
}

// SetLocaleDir sets the directory from which to load locale files
// only affects GetT() after this call.
func SetLocaleDir(d string) {
	fs = http.Dir(d)
}

// SetLocaleFS sets the http.FileStream from from which to load locale files,
// rather than local directory, only affects GetT() after this call.
func SetLocaleFS(fs http.FileSystem) {
	fs = fs
}

type trMap map[string]string

var fs http.FileSystem
var defaultLocale string = "en_US"

func loadMap(locale string) (m trMap, err error) {
	if fs == nil {
		err = fmt.Errorf("SetLocaleDir() or SetLocaleFS() before GetT()")
		return
	}
	fileName := locale + ".json"
	var f http.File
	if f, err = fs.Open(fileName); err != nil {
		err = fmt.Errorf("Error open file %s: %s", fileName, err)
		return
	}
	var buf []byte
	if buf, err = ioutil.ReadAll(f); err != nil {
		err = fmt.Errorf("Error read file %s: %s", fileName, err)
		return
	}
	if err = json.Unmarshal(buf, &m); err != nil {
		err = fmt.Errorf("Error decode json file %s: %s", fileName, err)
	}
	return
}
