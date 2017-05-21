#!/usr/bin/env bash
set -e
# slightly malformed input data
input_start=2016-2-1
input_end=2017-1-1

# After this, startdate and enddate will be valid ISO 8601 dates,
# or the script will have aborted when it encountered unparseable data
# such as input_end=abcd
startdate=$(date -I -d "$input_start") || exit -1
enddate=$(date -I -d "$input_end")     || exit -1

d="$startdate"
while [ "$d" != "$enddate" ]; do
  tend=$(date -I -d "$d + 1 day")
  echo "$(date): starting backup $d";
  psql -c "Copy (SELECT * FROM \"Appraisals\" WHERE to_timestamp(\"Created\") >= '$d' AND to_timestamp(\"Created\") < '$tend') TO STDOUT WITH CSV HEADER DELIMITER ',';" | gzip > $d.csv.gz
  rclone move ./$d.csv.gz gdrive:dbbackups/
  d=$(date -I -d "$d + 1 day")
done