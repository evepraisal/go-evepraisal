#!/usr/bin/env bash
set -e

export RCLONE_CONFIG="/etc/evepraisal/rclone.conf"
echo "$(date) - restarting service to get nmap buffer"
/usr/sbin/service evepraisal restart

echo "$(date) - starting backup of supporting files"
mkdir -p /usr/local/evepraisal/backups/
rclone copy /usr/local/evepraisal/db/certs gdrive:backups/evepraisal/usr/local/evepraisal/db/certs
rclone copy /etc/evepraisal/evepraisal.toml gdrive:backups/evepraisal/etc/evepraisal/

echo "$(date) - starting backup of appraisal database"
curl 127.0.0.1:8090/backup/appraisals | bzip2 --stdout > /usr/local/evepraisal/backups/appraisals.bak.bz2
echo "$(date) - uploading backup of appraisal database"
rclone move /usr/local/evepraisal/backups/appraisals.bz2 gdrive:backups/evepraisal/usr/local/evepraisal/db/
echo "$(date) - finished backup"
