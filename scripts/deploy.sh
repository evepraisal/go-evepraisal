#!/usr/bin/env bash
set -e

scp etc/systemd/system/evepraisal.service root@beta.evepraisal.com:/etc/systemd/system/evepraisal.service
ssh root@beta.evepraisal.com "
	systemctl daemon-reload
	rm /usr/local/bin/evepraisal"
scp target/evepraisal-linux-amd64 root@beta.evepraisal.com:/usr/local/bin/evepraisal
ssh root@beta.evepraisal.com "
	setcap 'cap_net_bind_service=+ep' /usr/local/bin/evepraisal
	systemctl enable evepraisal
	systemctl restart evepraisal"
