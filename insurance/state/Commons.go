package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

//ArgsMap map for ars array in interface form
type ArgsMap map[string]interface{}

// Returns a map containing the JSON object represented by args[0]
func getUnmarshalledArgument(args []string) (interface{}, error) {
	var event interface{}
	var err error

	if len(args) != 1 {
		err = errors.New("Expecting one JSON event object")
		return nil, err
	}
	eventBytes := []byte(args[0])
	err = json.Unmarshal(eventBytes, &event)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal arg: %s", err)
		return nil, err
	}
	if event == nil {
		err = fmt.Errorf("unmarshal arg created nil event")
		return nil, err
	}

	argsMap, found := event.(map[string]interface{})
	if !found {
		err := fmt.Errorf("arg is not a map shape")
		return nil, err
	}
	return argsMap, nil
}

func getStringParamFromArgs(name string, args interface{}) (string, error) {
	var result string
	var amargs ArgsMap
	var margs map[string]interface{}
	var found bool
	amargs, found = args.(ArgsMap)
	if !found {
		margs, found = args.(map[string]interface{})
		if !found {
			err := fmt.Errorf("not passed a map, type is %T", args)
			return "", err
		}
		amargs = ArgsMap(margs)
	}
	result, found = getObjectAsString(amargs, name)
	if !found {
		// not found
		err := errors.New(name + " is missing")
		return "", err
	}
	return result, nil
}

func getObjectAsString(objIn interface{}, qname string) (string, bool) {
	tbytes, found := getObject(objIn, qname)
	if found {
		t, found := tbytes.(string)
		if found {
			return t, true
		}
	}
	return "", false
}

func getObject(objIn interface{}, qname string) (interface{}, bool) {
	// return a copy of the selected object
	// handles full qualified name, starting at object's root
	obj, found := objIn.(map[string]interface{})
	if !found {
		objam, found := objIn.(ArgsMap)
		if !found {
			return nil, false
		}
		obj = map[string]interface{}(objam)
	}
	searchObj := map[string]interface{}(obj)
	s := strings.Split(qname, ".")
	// crawl the levels
	for i, v := range s {
		if i+1 < len(s) {
			tmp, found := searchObj[v]
			if found {
				searchObj, found = tmp.(map[string]interface{})
				if !found {
					searchObj, found = tmp.(ArgsMap)
					if !found {
					}
				}
			} else {

			}
		} else {
			returnObj, found := searchObj[v]
			if !found {
				// this debug statement is not useful normally as we must be able to
				// handle assetID as part of iot common and as parameter on its own
				// so we get false warnings on read functions, but do enable it if
				// having problems with deep nested structures
				return nil, false
			}
			return returnObj, true
		}
	}
	return nil, false
}
