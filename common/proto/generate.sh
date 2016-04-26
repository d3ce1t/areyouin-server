#/bin/bash
protoc --go_out=../ core.proto
protoc --java_out=/Users/jpadilla/AndroidStudioProjects/AreYouIN/app/src/main/java core.proto
#protoc --swift_out="/Users/jpadilla/Git/AreYouIN/areyouin-ios/AreYouIN/" core.proto
protoc --objc_out="/Users/jpadilla/Git/AreYouIN/areyouin-ios/AreYouIN/" core.proto
