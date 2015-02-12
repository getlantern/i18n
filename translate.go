package i18n

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/pivotal-cf-experimental/jibber_jabber"
)

// T consults the translated text by 'key' from the locale file specified when GetT(),
// then the default locale file, which defaults to 'en_US'
// but can be changed by SetDefaultLocale().
// If neither file containes it, will return 'key' surrounded by braces.
type T func(key string) (text string)

// GetUserT returns T according to current user's locale
func GetUserT() (t T, err error) {
	fallbackLocale := "en_US"
	userLocale, err := jibber_jabber.DetectIETF()
	if err != nil || userLocale == "C" {
		userLocale = fallbackLocale
	}
	if t, err = GetT(userLocale); err != nil {
		if t, err = GetT(fallbackLocale); err != nil {
			err = fmt.Errorf("Can't load user locale %s, nor fallback locale %s", userLocale, fallbackLocale)
		}
	}
	return
}

// GetT appends ".json" to locale passed in, and loads the file
// under the directory specified by SetLocaleDir()
// or from the file system specified by SetLocaleFS(),
// and returns the translate function.
// If the file didn't found, it will try the language part (i.e. fa for fa_IR).
// It also loads the default locale file which defaults to 'en_US.json'
// but can be changed in advance through SetDefaultLocale().
// The locale files should be json file with only key / text pairs of translations.
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
	if matched, _ := regexp.MatchString("^[a-z]{2}([_-][A-Z]{2}){0,1}$", locale); !matched {
		err = fmt.Errorf("Malformated locale string %s", locale)
		return
	}
	locale = strings.Replace(locale, "-", "_", -1)
	fileName := locale + ".json"
	var f http.File
	if f, err = fs.Open(fileName); err != nil {
		parts := strings.Split(locale, "_")
		langFileName := parts[0] + ".json"
		if f, err = fs.Open(langFileName); err != nil {
			err = fmt.Errorf("Neither %s nor %s can be opened: %s", fileName, langFileName, err)
			return
		}
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
