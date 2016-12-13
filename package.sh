#/usr/bin/bash
env GOOS=linux GOARCH=amd64 ./build.sh
mkdir -p areyouin-dist/extra
cp server/server areyouin-dist/areyouind
cp -r server/cert areyouin-dist/cert
cp server/extra/areyouin.example.yaml areyouin-dist/extra
cp server/extra/areyouin.service areyouin-dist/extra
cp server/extra/post-install.sh areyouin-dist/extra
cp server/extra/install.sh areyouin-dist
tar cvzf areyouin-dist.tgz areyouin-dist
rm -r areyouin-dist