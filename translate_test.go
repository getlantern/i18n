package i18n

import (
	"sync"
	"testing"

	"github.com/getlantern/i18n/locale"
	"github.com/getlantern/tarfs"
	"github.com/getlantern/testify/assert"
)

func TestTranslate(t *testing.T) {
	tr, err := GetT("en_US")
	assert.Error(t, err, "should error if no file system or dir given")

	SetLocaleDir("non-exist-dir")
	tr, err = GetT("en_US")
	assert.Error(t, err, "should error if dir is not existed")

	SetLocaleDir("locale")
	tr, err = GetT("en_US")
	assert.NoError(t, err, "should load existing locale")
	assertTranslation(t, tr, "HELLO", "Hello!")
	trCN, _ := GetT("zh_CN")
	assertTranslation(t, trCN, "HELLO", "你好!")
	_, err = GetT("zz_YY")
	assert.Error(t, err, "should not load non-existing locale")

	nonExist, _ := GetT("zh_CN")
	assertTranslation(t, nonExist, "NON_EXIST", "{NON_EXIST}")
}

func TestFallback(t *testing.T) {
	SetLocaleDir("locale")
	// default locale is en_US so we can fallback
	cn, _ := GetT("zh_CN")
	assertTranslation(t, cn, "ONLY_IN_EN", "I speak English!")

	// default locale is en_US and we don't have this key
	en, _ := GetT("en_US")
	assertTranslation(t, en, "ONLY_IN_ZH", "{ONLY_IN_ZH}")
	SetDefaultLocale("zh_CN")
	// default locale set to zh_CN and we can fallback now
	en2, _ := GetT("en_US")
	assertTranslation(t, en2, "ONLY_IN_ZH", "I speak Chinese!")
	// but only affects Ts after this setting
	assertTranslation(t, en, "ONLY_IN_ZH", "{ONLY_IN_ZH}")
}

func TestTarFS(t *testing.T) {
	fs, err := tarfs.New(locale.Resources, "")
	assert.NoError(t, err, "should load tarfs")
	SetLocaleFS(*fs)
	tr, err := GetT("en_US")
	assert.NoError(t, err, "should load locale from tarfs")
	assertTranslation(t, tr, "HELLO", "Hello!")
}

func TestGoroutine(t *testing.T) {
	SetDefaultLocale("en_US")
	SetLocaleDir("locale")
	cn, _ := GetT("zh_CN")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		assertTranslation(t, cn, "HELLO", "你好!")
		wg.Done()
	}()
	go func() {
		assertTranslation(t, cn, "ONLY_IN_EN", "I speak English!")
		wg.Done()
	}()
	wg.Wait()
}

func assertTranslation(t *testing.T, tr T, k string, e string) {
	if s := tr(k); s != e {
		t.Errorf("Expect T(\"%s\") to be \"%s\", got \"%s\"\n", k, e, s)
	}
}
