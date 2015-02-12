package i18n

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/getlantern/golog"
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

	currentLocale, err = newLocale("en_US") // TODO: look this up with jibber_jabber
	if err != nil {
		return err
	}
	return nil
}

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
