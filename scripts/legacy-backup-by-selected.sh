#!/usr/bin/env bash
set -e

appraisals=$(awk -vORS=, '{ print $1 }' used-last-6-months.txt | sed 's/,$//')
echo $appraisals

psql -c "Copy (SELECT * FROM \"Appraisals\" WHERE \"Id\" IN ($appraisals)) TO STDOUT WITH CSV HEADER DELIMITER ',';" | gzip > selected.csv.gz
rclone move ./selected.csv.gz gdrive:dbbackups/
