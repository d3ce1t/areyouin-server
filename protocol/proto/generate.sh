#/bin/bash
protoc -I../../common/proto/ -I./ --go_out=Mcore.proto=peeple/areyouin/common:../ protocol.proto
protoc -I../../common/proto/ -I./ --java_out=/Users/jpadilla/AndroidStudioProjects/AreYouIN/app/src/main/java protocol.proto
#protoc -I../../common/proto/ -I./ --swift_out="/Users/jpadilla/Git/AreYouIN/areyouin-ios/AreYouIN/" protocol.proto
protoc -I../../common/proto/ -I./ --objc_out="/Users/jpadilla/Git/AreYouIN/areyouin-ios/AreYouIN/" protocol.proto
