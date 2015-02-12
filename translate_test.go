package i18n

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/getlantern/testify/assert"
)

func TestAll(t *testing.T) {
	initTestLocales(t)

	name := "Jimmy"

	err := Init(lookup, "adsfasdf")
	assert.Error(t, err, "Initializing with bad locale should fail")
	err = Init(lookup, "en_US")
	if assert.NoError(t, err, "Initializing with defined default locale should succeed") {
		assertTrans(t, "English translation should have contained name", "en_US", "HELLO", name)

		err = SetLocale("afewradsf")
		assert.Error(t, err, "Setting bad locale should fail")
		err = SetLocale("zh")
		if assert.NoError(t, err, "Setting non-defined locale should succeed") {
			assertTrans(t, "Non-defined translation should have contained name", "en_US", "HELLO", name)
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

func lookup(locale string) ([]byte, error) {
	return testLocaleJsons[locale], nil
}

func initTestLocales(t *testing.T) {
	testLocaleJsons = make(map[string][]byte, len(testLocales))
	for k, v := range testLocales {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Unable to marshal json: %v", err)
		}
		testLocaleJsons[k] = b
	}
}

var testLocales = map[string]map[string]string{
	"en_US": map[string]string{
		"HELLO":      "Hello %s!",
		"ONLY_IN_EN": "I speak America English!",
	},

	"en": map[string]string{
		"ONLY_IN_EN": "I speak Generic English!",
	},

	"zh_CN": map[string]string{
		"HELLO":      "你好 %s!",
		"ONLY_IN_ZH": "I speak Chinese!",
	},
}

var testLocaleJsons map[string][]byte
