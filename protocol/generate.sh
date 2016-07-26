#/bin/bash

# Core.proto
protoc -I./ \
  --go_out=./ \
  --java_out="/Users/jpadilla/AndroidStudioProjects/areyouin-android/app/src/main/java" \
  --objc_out="/Users/jpadilla/Git/Peeple/areyouin-ios/AreYouIN/" \
  core/core.proto

# Protocol.proto
protoc -Icore -I./ \
  --go_out=Mcore.proto=peeple/areyouin/protocol/core:./ \
  --java_out="/Users/jpadilla/AndroidStudioProjects/areyouin-android/app/src/main/java" \
  --objc_out="/Users/jpadilla/Git/Peeple/areyouin-ios/AreYouIN/" \
  protocol.proto
