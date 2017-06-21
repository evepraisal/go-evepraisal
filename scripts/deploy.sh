#!/usr/bin/env bash
set -e

scp etc/systemd/system/evepraisal.service $USERNAME@$HOSTNAME:/etc/systemd/system/evepraisal.service
ssh $USERNAME@$HOSTNAME "
	systemctl daemon-reload
	rm /usr/local/bin/evepraisal"
scp target/evepraisal-linux-amd64 $USERNAME@$HOSTNAME:/usr/local/bin/evepraisal
ssh $USERNAME@$HOSTNAME "
	setcap 'cap_net_bind_service=+ep' /usr/local/bin/evepraisal
	systemctl enable evepraisal
	systemctl restart evepraisal"
