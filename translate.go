package i18n

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/getlantern/golog"
	"github.com/getlantern/jibber_jabber"
)

var (
	log = golog.LoggerFor("i18n")

	getMessages   GetMessagesFunc
	defaultLocale loc
	currentLocale loc
	msgsByLang    map[string]messages
	mutex         sync.RWMutex
)

type GetMessagesFunc func(locale string) ([]byte, error)

// Trans translates the given key into a message based on the current locale,
// formatting the string using the supplied (optional) args. This method will
// fall back to other locales if the key isn't defined for the current locale.
// The search order (with examples) is as follows:
//
//   1. current locale        (zh_CN)
//   2. lang only             (zh)
//   3. default locale        (en_US)
//   4. lang only of default  (en)
//
func Trans(key string, args ...interface{}) string {
	mutex.RLock()
	defer mutex.RUnlock()
	s := msgsByLang[currentLocale.full][key]
	if s == "" {
		// Try only the language part
		s = msgsByLang[currentLocale.lang][key]
	}
	if s == "" {
		// Try the default locale
		s = msgsByLang[defaultLocale.full][key]
	}
	if s == "" {
		// Try only the language part of the default locale
		s = msgsByLang[defaultLocale.lang][key]
	}

	// Format string
	if s != "" && len(args) > 0 {
		s = fmt.Sprintf(s, args...)
	}

	return s
}

// Init initializes i18n using the given GetMessagesFn to look up JSON message
// catalogs for locales, and setting the default locale to the given value.
// The JSON message catalog is just a map of string keys to string messages. The
// messages can contain standard Go format strings.
//
// If the default locale is not in a valid format, this function will return
// an error.
func Init(getMessagesFn GetMessagesFunc, defaultlocale string) error {
	mutex.Lock()
	defer mutex.Unlock()
	getMessages = getMessagesFn
	var err error
	defaultLocale, err = newLocale(defaultlocale)
	if err != nil {
		return err
	}
	m := loadMessages(defaultLocale)
	msgsByLang = map[string]messages{defaultLocale.full: m}
	if !defaultLocale.isLangOnly() {
		// Also load the language-only resource (if available)
		msgsByLang[defaultLocale.lang] = loadMessages(defaultLocale.langOnly())
	}

	// Default current locale to defaultLocale
	currentLocale = defaultLocale
	userLocale, err := jibber_jabber.DetectIETF()
	if err != nil {
		log.Debugf("Error detecting user locale, defaulting to %v: %v", currentLocale, err)
	} else if userLocale == "C" {
		log.Debugf("Unable to detect user locale, defaulting to %v", currentLocale)
	} else {
		currentLocale, err = newLocale(userLocale)
		if err != nil {
			currentLocale = defaultLocale
			log.Debugf("Got invalid user locale %v, defaulting to %v: %v", userLocale, currentLocale, err)
		} else {
			log.Debugf("Set current locale to %v based on user's locale", currentLocale)
		}
	}
	return nil
}

// InitWithDir is like Init, but uses resources read from files in the given
// messagesDir with names corresponding to locales (e.g. en_US).
func InitWithDir(messagesDir string, defaultlocale string) error {
	return Init(func(locale string) ([]byte, error) {
		return ioutil.ReadFile(filepath.Join(messagesDir, locale))
	}, defaultlocale)
}

// SetLocale sets the current locale to the given value. If the locale is not in
// a valid format, this function will return an error and leave the current
// locale as is.
func SetLocale(locale string) error {
	mutex.Lock()
	defer mutex.Unlock()
	l, err := newLocale(locale)
	if err != nil {
		return err
	}

	m := msgsByLang[l.full]
	if m == nil {
		msgsByLang[l.full] = loadMessages(l)
	}
	if !l.isLangOnly() {
		// Also load the data for the language portion of the locale
		m = msgsByLang[l.lang]
		if m == nil {
			msgsByLang[l.lang] = loadMessages(l)
		}
	}

	currentLocale = l
	return nil
}

// loc is a parsed locale
type loc struct {
	full string
	lang string
}

func newLocale(full string) (loc, error) {
	if matched, _ := regexp.MatchString("^[a-z]{2}([_-][A-Z]{2}){0,1}$", full); !matched {
		return loc{}, fmt.Errorf("Malformated locale string %s", full)
	}
	full = strings.Replace(full, "-", "_", -1)
	parts := strings.Split(full, "_")
	lang := strings.ToLower(parts[0])
	return loc{
		full: full,
		lang: lang,
	}, nil
}

func (l loc) isLangOnly() bool {
	return l.full == l.lang
}

func (l loc) langOnly() loc {
	return loc{l.lang, l.lang}
}

func (l loc) String() string {
	return l.full
}

// messages is a message catalog from key to message
type messages map[string]string

func loadMessages(l loc) messages {
	m := make(messages, 0)
	buf, err := getMessages(l.full)
	if err != nil {
		log.Debugf("Error getting locale data for %v: %v", l, err)
		return m
	}
	if buf == nil {
		log.Debugf("No locale data found for %v", l)
		return m
	}
	if err = json.Unmarshal(buf, &m); err != nil {
		log.Errorf("Error decoding json for locale %v: %v", l, err)
	}
	return m
}
