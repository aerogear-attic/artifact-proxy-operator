package plist

import (
	"fmt"
	"testing"
)

func TestProduceXml(t *testing.T) {
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
</plist>`
	xml := ProduceXML("http://test.com")
	if xml != expectResponse {
		fmt.Printf("Expected \n%s\n", expectResponse)
		fmt.Printf("But Got \n%s\n", xml)
		t.Fatal("unexpected xml response")
	}
	ProduceXML("http://test.com")
}
