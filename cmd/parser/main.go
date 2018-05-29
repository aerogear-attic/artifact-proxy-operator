package main

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type PlistArray struct {
	Integer []int `xml:"integer"`
}

func main() {

	sourceXML := `<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version="1.0">
	  <dict>
		<key>items</key>
		<array>
		  <dict>
			<key>assets</key>
			<array>
			  <dict>
				<key>kind</key>
				<string>software-package</string>
				<key>url</key>
				<string>https://rhmap.csteam2.skunkhenry.com/digman/ios-v3/dist/c5c83e51-3cb1-44fe-bb65-0662f9b137d8/ios~7.0~4~SimpleiOSObjectiveCPushApp.ipa?digger=diggers.qed1-farm2-osx1</string>
			  </dict>
			</array>
			<key>metadata</key>
			<dict>
			  <key>bundle-identifier</key>
			  <string>$(PRODUCT_BUNDLE_IDENTIFIER)</string>
			  <key>bundle-version</key>
			  <string>1.0</string>
			  <key>kind</key>
			  <string>software</string>
			  <key>title</key>
			  <string>SimpleiOSObjectiveCPushApp</string>
			</dict>
		  </dict>
		</array>
	  </dict>
	</plist>`
	result := map[string]interface{}{}
	dec := xml.NewDecoder(strings.NewReader(sourceXML))
	dec.Strict = false
	var workingKey string

	for {
		token, _ := dec.Token()
		if token == nil {
			break
		}
		switch start := token.(type) {
		case xml.StartElement:
			fmt.Printf("startElement = %+v\n", start)
			switch start.Name.Local {
			case "key":
				var k string
				err := dec.DecodeElement(&k, &start)
				if err != nil {
					fmt.Println(err.Error())
				}
				workingKey = k
			case "string":
				var s string
				err := dec.DecodeElement(&s, &start)
				if err != nil {
					fmt.Println(err.Error())
				}
				result[workingKey] = s
				workingKey = ""
			case "integer":
				var i int
				err := dec.DecodeElement(&i, &start)
				if err != nil {
					fmt.Println(err.Error())
				}
				result[workingKey] = i
				workingKey = ""
			case "array":
				var ai PlistArray
				err := dec.DecodeElement(&ai, &start)
				if err != nil {
					fmt.Println(err.Error())
				}
				result[workingKey] = ai
				workingKey = ""
			default:
				fmt.Errorf("Unrecognized token")
			}
		}
	}
	fmt.Printf("%+v", result)
}
