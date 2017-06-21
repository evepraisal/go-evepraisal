#!/usr/bin/env bash
set -e

# Copy up service file
scp etc/systemd/system/evepraisal.service $USERNAME@$HOSTNAME:/etc/systemd/system/evepraisal.service

# Reload service file and remove current binary
ssh $USERNAME@$HOSTNAME "
	systemctl daemon-reload
	rm -f /usr/local/bin/evepraisal"

# Copy up new binary
scp target/evepraisal-linux-amd64 $USERNAME@$HOSTNAME:/usr/local/bin/evepraisal

# Set binary capabilities,  enable the service (if it isn't already) and restart it
ssh $USERNAME@$HOSTNAME "
	setcap 'cap_net_bind_service=+ep' /usr/local/bin/evepraisal
	systemctl enable evepraisal
	systemctl restart evepraisal"
