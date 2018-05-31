package plist

import (
	"fmt"
)

func ProduceXML(proxyUrl string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
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
            <string>%s</string>
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
</plist>`, proxyUrl)
}
