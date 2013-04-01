package gutil

import (
	"bytes"
	"io"
	"bufio"
)

func ReadLine(reader *bufio.Reader) (string, error) {
	buff := bytes.NewBufferString("")
	for {
		strByte, isPrefix, err := reader.ReadLine()
		if err != nil {
			return "", err
		}
		buff.Write(strByte)
		if !isPrefix {
			str := buff.String()
			return str, nil
		}
	}

	return buff.String(), nil
}

func WriteFull(writer io.Writer, buff []byte) error {
	for total, bufLen := 0, len(buff); total < bufLen; {
		one, err := writer.Write(buff)
		if err != nil {
			return err
		}

		total += one
	}

	return nil
}

func ReadFull(reader io.Reader, buff []byte) error {
	for total, bufLen := 0, len(buff); total < bufLen; {
		one, err := reader.Read(buff)
		if err != nil {
			return err
		}

		total += one
	}

	return nil
}
