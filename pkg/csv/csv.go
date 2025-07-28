package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"

	"github.com/gocarina/gocsv"
)

func getExpectedHeaders(v interface{}) []string {
	var headers []string
	t := reflect.TypeOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("csv")
		if tag != "" && tag != "-" {
			headers = append(headers, tag)
		}
	}
	return headers
}

func UnmarshalWithHeaderValidation(seeker io.ReadSeeker, out interface{}) error {
	csvForHeaderCheck := csv.NewReader(seeker)
	header, err := csvForHeaderCheck.Read()
	if err != nil {
		return fmt.Errorf("failed to read header for validation: %w", err)
	}

	elemType := reflect.TypeOf(out).Elem().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	expectedHeaders := getExpectedHeaders(reflect.New(elemType).Interface())
	if !reflect.DeepEqual(header, expectedHeaders) {
		return fmt.Errorf("CSV header mismatch. Got: %v, Want: %v", header, expectedHeaders)
	}

	_, err = seeker.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek back to the start of the reader: %w", err)
	}

	if err := gocsv.Unmarshal(seeker, out); err != nil {
		return fmt.Errorf("gocsv unmarshal failed: %w", err)
	}

	return nil
}
