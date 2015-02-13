package i18n

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/getlantern/golog"
	"github.com/getlantern/jibber_jabber"
)

var (
	localeRegexp  string          = "^[a-z]{2}([_-][A-Z]{2}){0,1}$"
	log                           = golog.LoggerFor("i18n")
	fs            http.FileSystem = http.Dir("locale")
	defaultLocale string          = "en_US"
	defaultLang   string          = "en"
	trMutex       sync.RWMutex
	trMap         map[string]string
)

// T translates the given key into a message based on the current locale,
// formatting the string using the supplied (optional) args. This method will
// fall back to other locales if the key isn't defined for the current locale.
// The search order (with examples) is as follows:
//
//   1. current locale        (zh_CN)
//   2. lang only             (zh)
//   3. default locale        (en_US)
//   4. lang only of default  (en)
//
func T(key string, args ...interface{}) string {
	trMutex.RLock()
	defer trMutex.RUnlock()
	s := trMap[key]
	// Format string
	if s != "" && len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}

	return s
}

// SetLocaleDir sets the directory from which to load locale files
// if they are not under the default directory 'locale'
func SetLocaleDir(d string) {
	fs = http.Dir(d)
}

// SetLocaleFS tells i18n to load locale files from a http.FileStream
// interface rather than local directory.
func SetLocaleFS(fs http.FileSystem) {
	fs = fs
}

// SetLocale sets the current locale to the given value. If the locale is not in
// a valid format, this function will return an error and leave the current
// locale as is.
func SetLocale(locale string) error {
	if matched, _ := regexp.MatchString(localeRegexp, locale); !matched {
		return fmt.Errorf("Malformated locale string %s", locale)
	}
	locale = strings.Replace(locale, "-", "_", -1)
	parts := strings.Split(locale, "_")
	lang := parts[0]
	newTrMap := make(map[string]string)
	mergeLocaleToMap(newTrMap, defaultLang)
	mergeLocaleToMap(newTrMap, defaultLocale)
	mergeLocaleToMap(newTrMap, lang)
	mergeLocaleToMap(newTrMap, locale)
	if len(newTrMap) == 0 {
		return fmt.Errorf("Not found any translations, locale not set")
	}
	trMutex.Lock()
	defer trMutex.Unlock()
	trMap = newTrMap
	return nil
}

func mergeLocaleToMap(dst map[string]string, locale string) {
	if m, e := loadMapFromFile(locale); e != nil {
		log.Tracef("Locale %s not loaded: %s", locale, e)
	} else {
		for k, v := range m {
			dst[k] = v
		}
	}
}

func loadMapFromFile(locale string) (m map[string]string, err error) {
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

// Detect OS locale on startup, no setup required if files under 'locale'
func init() {
	userLocale, err := jibber_jabber.DetectIETF()
	if err != nil || userLocale == "C" {
		userLocale = defaultLocale
	}
	SetLocale(defaultLocale)
}
