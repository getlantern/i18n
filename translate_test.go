package i18n

import (
	"sync"
	"testing"

	"github.com/getlantern/i18n/locale"
	"github.com/getlantern/tarfs"
	"github.com/getlantern/testify/assert"
)

func TestTranslate(t *testing.T) {
	assertTranslation(t, "", "HELLO")
	UseOSLocale()
	assertTranslation(t, "I speak America English!", "ONLY_IN_EN_US")
	assertTranslation(t, "I speak Generic English!", "ONLY_IN_EN")
	assertTranslation(t, "", "NOT_EXISTED")

	WillReadFromDir("not-existed-dir")
	err := SetLocale("en_US")
	assert.Error(t, err, "should error if dir is not existed")

	WillReadFromDir("locale")
	assert.Error(t, SetLocale("e0"), "should error on malformed locale")
	assert.Error(t, SetLocale("e0-DO"), "should error on malformed locale")
	assert.Error(t, SetLocale("e0-DO.C"), "should error on malformed locale")
	assert.NoError(t, SetLocale("en"), "should change locale")
	assert.NoError(t, SetLocale("en_US"), "should change locale")
	assertTranslation(t, "Hello Q!", "HELLO", "Q")
	assertTranslation(t, "", "NOT_EXISTED", 1)
	assert.NoError(t, SetLocale("zh_CN"), "should change locale")
	assertTranslation(t, "Q你好!", "HELLO", "Q")

	// fallbacks
	assertTranslation(t, "I speak Mandarin!", "ONLY_IN_ZH_CN")
	assertTranslation(t, "I speak Chinese!", "ONLY_IN_ZH")
	assertTranslation(t, "I speak America English!", "ONLY_IN_EN_US")
	assertTranslation(t, "I speak Generic English!", "ONLY_IN_EN")
}

func TestTarFS(t *testing.T) {
	fs, err := tarfs.New(locale.Resources, "")
	assert.NoError(t, err, "should load tarfs")
	SetLocale("en")
	fromTar := func(path string) ([]byte, error) {
		return fs.Get(path)
	}
	WillReadByFunc(fromTar)
	assert.NoError(t, err, "should load locale from tarfs")
	assertTranslation(t, "", "ONLY_IN_ZH")
	SetLocale("zh_CN")
	assertTranslation(t, "I speak Chinese!", "ONLY_IN_ZH")
}

func TestGoroutine(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		assertTranslation(t, "I speak America English!", "ONLY_IN_EN_US")
		wg.Done()
	}()
	go func() {
		assertTranslation(t, "I speak Generic English!", "ONLY_IN_EN")
		wg.Done()
	}()
	wg.Wait()
}

func assertTranslation(t *testing.T, expected string, key string, args ...interface{}) {
	if s := T(key, args...); s != expected {
		t.Errorf("Expect T(\"%s\") to be \"%s\", got \"%s\"\n", key, expected, s)
	}
}
