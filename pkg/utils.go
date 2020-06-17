package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func GetProjectVersion() string {
	contentAsBytes, err := ioutil.ReadFile(".project-version")

	if err != nil {
		fmt.Print(err)
	}

	contentAsString := string(contentAsBytes)
	contentAsString = strings.TrimSuffix(contentAsString, "\n")

	return contentAsString
}
