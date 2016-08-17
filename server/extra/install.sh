#/usr/bin/bash
mkdir -p /opt/areyouin
cp areyouind /opt/areyouin
cp -r cert /opt/areyouin/cert
cp extra/areyouin.example.yaml /opt/areyouin
chown -R areyouin /opt/areyouin
chmod -R o-rwx /opt/areyouin/cert
./extra/post-install.sh
echo 'AreYouIN server installed'
echo 'Please run systemctl start areyouin.service to star the server right now'