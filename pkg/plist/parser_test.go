package plist

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestModifyXML(t *testing.T) {
	inputXml := `<?xml version="1.0" encoding="UTF-8"?>
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
            <string>https://rhmap.csteam2.skunkhenry.com</string>
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
</plist>
`

	expectResponse := `<?xml version="1.0" encoding="UTF-8"?>
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
            <string>http://test.com</string>
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
</plist>
`
	input := bytes.NewReader([]byte(inputXml))
	resp, err := ModifyXML(input, "url", "http://test.com")
	if err != nil {
		t.Fatal("error updating xml " + err.Error())
	}
	if strings.TrimSpace(fmt.Sprintln(resp)) != strings.TrimSpace(expectResponse) {
		fmt.Printf("Expected \n%s\n", expectResponse)
		fmt.Printf("But Got \n%s\n", fmt.Sprintln(resp))
		t.Fatal("unexpected xml response")
	}
}
