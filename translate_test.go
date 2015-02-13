package i18n

import (
	"fmt"
	"testing"

	"github.com/getlantern/testify/assert"
)

func TestAllMap(t *testing.T) {
	doTestAll(t, func(defaultLocale string) error {
		return Init(func(locale string) map[string]string {
			return testLocales[locale]
		}, defaultLocale)
	})
}

func TestAllDir(t *testing.T) {
	doTestAll(t, func(defaultLocale string) error {
		return InitWithDir("messages", defaultLocale)
	})
}

func doTestAll(t *testing.T, initFn func(defaultLocale string) error) {
	name := "Jimmy"

	err := initFn("adsfasdf")
	assert.Error(t, err, "Initializing with bad locale should fail")
	err = initFn("en_US")
	if assert.NoError(t, err, "Initializing with defined default locale should succeed") {
		assertTrans(t, "English translation should contain name", "en_US", "HELLO", name)
		assertTrans(t, "Default locale should fall back to language-only when necessary", "en", "ONLY_IN_EN", name)
		s := Trans("ONLY_IN_ZH")
		assert.Equal(t, "", s, "Non-existent key should return blank message")

		err = SetLocale("afewradsf")
		assert.Error(t, err, "Setting bad locale should fail")
		err = SetLocale("zh")
		if assert.NoError(t, err, "Setting non-defined locale should succeed") {
			assertTrans(t, "Non-defined translation should have fallen back to en_US", "en_US", "IN_EN")
		}

		err = SetLocale("zh_CN")
		if assert.NoError(t, err, "Setting defined locale should succeed") {
			assertTrans(t, "Chinese translation should have contained name", "zh_CN", "HELLO", name)
		}
	}
}

func assertTrans(t *testing.T, desc string, locale string, key string, args ...interface{}) {
	s := Trans(key, args...)
	assert.Equal(t, fmt.Sprintf(testLocales[locale][key], args...), s, desc)
}

var testLocales = map[string]map[string]string{
	"en_US": map[string]string{
		"HELLO":         "Hello %s!",
		"IN_EN":         "I speak America English!",
		"ONLY_IN_EN_US": "Howdie!",
	},

	"en": map[string]string{
		"IN_EN":      "I speak Generic English!",
		"ONLY_IN_EN": "I'm special!",
	},

	"zh_CN": map[string]string{
		"HELLO":      "你好 %s!",
		"ONLY_IN_ZH": "I speak Chinese!",
	},
}
