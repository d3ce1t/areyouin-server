#/usr/bin/bash

# Add areyouin user with no shell
useradd -r -s /bin/false areyouin

# Permissions
chown -R areyouin /opt/areyouin
chmod -R o-rwx /opt/areyouin/cert

# Configure service to start automatically
# NOTE: Debian assumed
cp extra/areyouin.service /lib/systemd/system
systemd enable areyouin.service