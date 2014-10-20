
START_DATE=$(date "+%Y-%m-%d_%H%M%S")

cat /etc/issue | gzip | _send_file "${START_DATE}_issue.gz"
dpkg --get-selections | gzip | _send_file "${START_DATE}_packages.gz"
