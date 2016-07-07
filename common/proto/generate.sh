#/bin/bash
protoc --go_out=../ core.proto
protoc --java_out=/Users/jpadilla/AndroidStudioProjects/areyouin-android/app/src/main/java core.proto
#protoc --swift_out="/Users/jpadilla/Git/AreYouIN/areyouin-ios/AreYouIN/" core.proto
protoc --objc_out="/Users/jpadilla/Git/Peeple/areyouin-ios/AreYouIN/" core.proto
