package main

import "testing"

func TestMtr(t *testing.T) {
	err := Mtr("www.baidu.com")
	if err != nil {
		panic(err)
	}
}
