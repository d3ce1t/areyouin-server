#/usr/bin/bash
mkdir -p /opt/areyouin
cp areyouind /opt/areyouin
cp -r cert /opt/areyouin/cert
cp extra/areyouin.example.yaml /opt/areyouin
./extra/post-install.sh
echo 'AreYouIN server installed'
echo 'Please run systemctl start areyouin.service to star the server right now'