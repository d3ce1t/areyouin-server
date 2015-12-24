#/bin/bash
protoc -I../../common/proto/ -I./ --go_out=Mcore.proto=peeple/areyouin/common:../ protocol.proto
protoc -I../../common/proto/ -I./ --java_out=/Users/jpadilla/AndroidStudioProjects/AreYouIN/app/src/main/java protocol.proto
