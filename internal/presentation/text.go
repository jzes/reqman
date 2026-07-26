package presentation

import "unicode/utf8"

func removeLastRune(text string) (string, bool) {
	if text == "" {
		return text, false
	}
	_, size := utf8.DecodeLastRuneInString(text)
	return text[:len(text)-size], true
}
