package i18n

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

var (
	getLocaleData GetLocaleDataFunc
	defaultLocale string
	currentLocale string
	locales       map[string]trData
	mutex         sync.RWMutex
)

type GetLocaleDataFunc func(locale string) ([]byte, error)

func Init(getLocaleDataFn GetLocaleDataFunc, defaultlocale string) error {
	mutex.Lock()
	defer mutex.Unlock()
	getLocaleData = getLocaleDataFn
	var err error
	defaultLocale, err = normalizeLocale(defaultlocale)
	if err != nil {
		return err
	}
	m, err := loadMap(getLocaleData, defaultLocale)
	if err != nil {
		return err
	}
	locales = map[string]trData{defaultLocale: m}

	currentLocale = "en_US" // TODO: look this up with jibber_jabber
	return nil
}

func Trans(key string, args ...interface{}) string {
	mutex.RLock()
	defer mutex.RUnlock()
	s := locales[currentLocale][key]
	if s == "" {
		// Try just the language part
		parts := strings.Split(currentLocale, "_")
		langOnly := parts[0]
		m := locales[langOnly]
		if m == nil {
			// Try the default locale
			m = locales[defaultLocale]
		}
		s = m[key]
	}

	if s != "" && len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}
	return s
}

func SetLocale(locale string) error {
	mutex.Lock()
	defer mutex.Unlock()
	var err error
	locale, err = normalizeLocale(locale)
	if err != nil {
		return err
	}
	m := locales[locale]
	if m == nil {
		m, err = loadMap(getLocaleData, locale)
		if err != nil {
			return err
		}
		locales[locale] = m
	}
	currentLocale = locale
	return nil
}

func normalizeLocale(locale string) (string, error) {
	if matched, _ := regexp.MatchString("^[a-z]{2}([_-][A-Z]{2}){0,1}$", locale); !matched {
		return "", fmt.Errorf("Malformated locale string %s", locale)
	}
	locale = strings.Replace(locale, "-", "_", -1)
	parts := strings.Split(locale, "_")
	lang := strings.ToLower(parts[0])
	if len(parts) == 1 {
		return lang, nil
	}
	region := strings.ToUpper(parts[1])
	return fmt.Sprintf("%v_%v", lang, region), nil
}

type trData map[string]string

func loadMap(getLocaleData GetLocaleDataFunc, locale string) (m trData, err error) {
	if matched, _ := regexp.MatchString("^[a-z]{2}([_-][A-Z]{2}){0,1}$", locale); !matched {
		err = fmt.Errorf("Malformated locale string %v", locale)
		return
	}
	locale = strings.Replace(locale, "-", "_", -1)
	buf, err := getLocaleData(locale)
	if err != nil {
		err = fmt.Errorf("No locale data found for %v", err)
		// parts := strings.Split(locale, "_")
		// langOnly := parts[0]
		// buf, err = getLocaleData(langOnly)
		// if err != nil {
		// 	err = fmt.Errorf("Neither %v nor %v have localization data: %v", locale, langOnly, err)
		// 	return
		// }
	}

	m = make(trData, 0)
	if err = json.Unmarshal(buf, &m); err != nil {
		err = fmt.Errorf("Error decoding json for locale %v: %v", locale, err)
	}

	return
}
