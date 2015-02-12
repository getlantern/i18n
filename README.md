# i18n

A simple i18n library in Go, only supports text translation currently.

# Usage

Place all your translation json files under a directory, then `i18n.SetLocaleDir("<dir here>")`, or `i18n.SetLocaleFS(fs)` to load files from a customized `http.FileStream`.

Then simply (works across goroutines)

```go
t := i18n.GetT("fa_IR")
fmt.Println(t("HELLO"))
```
