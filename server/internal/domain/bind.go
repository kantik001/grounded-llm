package domain

var (
	bundleLocaleFn  func(string) string
	defaultLocaleFn func() string
)

// BindBundleLocale wires locale bundle resolution (from internal/locale).
func BindBundleLocale(fn func(string) string) {
	bundleLocaleFn = fn
}

// BindDefaultLocale wires the process default locale (from internal/locale).
func BindDefaultLocale(fn func() string) {
	defaultLocaleFn = fn
}

func bundleLocale(locale string) string {
	if bundleLocaleFn != nil {
		return bundleLocaleFn(locale)
	}
	return locale
}

func defaultLocale() string {
	if defaultLocaleFn != nil {
		return defaultLocaleFn()
	}
	return "en"
}
