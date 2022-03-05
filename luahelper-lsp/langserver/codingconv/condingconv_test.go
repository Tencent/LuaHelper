package codingconv

import (
	"fmt"
	"testing"
)

func TestComplete1(t *testing.T) {
	var data byte = 0
	for data = 0; data < 255; data++ {
		if getSprintfNum(data) != getCalcNum(data) {
			t.Fatalf("getCalcNum err, %d", data)
		} else {
			t.Logf("equal data=%d", data)
		}
	}

	if getSprintfNum(255) != getCalcNum(255) {
		t.Fatalf("getCalcNum err, %d", 255)
	}

	t.Logf("ok")
}

func getSprintfNum(data byte) int {
	str := fmt.Sprintf("%b", data)
	var i int = 0
	for i < len(str) {
		if str[i] != '1' {
			break
		}
		i++
	}
	return i
}

func getCalcNum(data byte) int {
	var i int = 0
	for data > 0 {
		if data%2 == 0 {
			i = 0
		} else {
			i++
		}
		data /= 2
	}

	return i
}
