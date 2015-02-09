REDIS_CONFIG="/etc/redis/redis.conf"
DUMP_DIR="$(cat $REDIS_CONFIG | grep "^dir " | awk '{print $2}')"

if [ -z "$DUMP_DIR" ]
then
  echo "Not found param 'dir' in $REDIS_CONFIG"
  exit 1
fi

DUMP_FILE="$(cat /etc/redis/redis.conf | grep "^dbfilename " | awk '{print $2}')"

if [ -z "$DUMP_FILE" ]
then
  echo "Not found param 'dbfilename' in $REDIS_CONFIG"
  exit 1
fi

DUMP_PATH="${DUMP_DIR}/${DUMP_FILE}"

LASTSAVE="$(redis-cli lastsave)"

redis-cli bgsave

if [ $? -ne 0 ]; then
  echo "Failed to start bgsave"
  exit 1
fi

echo Backing up database $DUMP_PATH

while [ $LASTSAVE == "$(redis-cli lastsave)" ]
do
  sleep 1
done

START_DATE=$(date "+%Y-%m-%d_%H%M%S")

cat $DUMP_PATH | gzip -f | _send_file "${START_DATE}-${DUMP_PATH}"
