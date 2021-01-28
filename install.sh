#!/bin/bash
if [ $UID != 0 ]; then
  echo "Installation must be run as root"
  exit 1
fi

# Stop existing service
systemctl stop soqtt.service

# User setup
adduser --system soqtt
mkdir -p /opt/soqtt

# Files
cp -r `dirname $0`/* /opt/soqtt

# Service setup
ln -sf /opt/soqtt/soqtt.service /etc/systemd/system
systemctl daemon-reload
systemctl enable soqtt.service
systemctl start soqtt.service
