package gutil

import (
	"bytes"
	"errors"
)

func findUnquoteToken(arr []byte, i int) (string, int, error) {
	buff := bytes.NewBufferString("")
	for blen := len(arr); i < blen; i++ {
		ch := arr[i]
		if IsWhiteChar(ch) {
			return buff.String(), i - 1, nil
		} else if ch == '\'' || ch == '"' {
			return "", -1, errors.New("unexcepted ' or \" in not dquoted token .")
		} else {
			buff.WriteByte(ch)
		}
	}
	return buff.String(), i - 1, nil
}

func findQuoteToken(arr []byte, i int) (string, int, error) {
	buff := bytes.NewBufferString("")
	for blen, i := len(arr), i+1; i < blen; i++ {
		ch := arr[i]
		if ch == '"' {
			if ll := buff.Len(); ll > 0 && buff.Bytes()[ll-1] == '\\' {
				buff.WriteByte(ch)
			} else {
				return buff.String(), i, nil
			}
		} else {
			buff.WriteByte(ch)
		}
	}

	return "", i, errors.New("unend dquote token .")
}

func findToken(arr []byte, i int) (string, int, error) {
	if arr[i] == '"' {
		return findQuoteToken(arr, i)
	}

	return findUnquoteToken(arr, i)
}

func IsWhiteChar(ch byte) bool {
	return ch == ' ' || ch == ' '
}

func skipWhiteChar(arr []byte, i int) (int, error) {
	if !IsWhiteChar(arr[i]) {
		return -1, errors.New("should be white char, but not .")
	}

	i = i + 1
	for blen := len(arr); i < blen; i++ {
		if !IsWhiteChar(arr[i]) {
			break
		}
	}

	return i, nil
}

func ParseStrLine(strLine string) ([]string, error) {
	result := make([]string, 0)

	for i, arr := 0, bytes.NewBufferString(strLine).Bytes(); i < len(arr); {
		token, ni, err := findToken(arr, i)
		if err != nil {
			return nil, err
		}

		result = append(result, token)
		if (ni + 1) >= len(arr) {
			break
		}

		ni = ni + 1
		ni, err = skipWhiteChar(arr, ni)
		if err != nil {
			return nil, err
		}

		i = ni
	}

	return result, nil
}
