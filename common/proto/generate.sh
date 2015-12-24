#/bin/bash
protoc --go_out=../ core.proto
protoc --java_out=/Users/jpadilla/AndroidStudioProjects/AreYouIN/app/src/main/java core.proto
