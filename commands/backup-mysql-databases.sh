
if [ -z "$MYSQL_USER" ];then
    MYSQL_USER=$(whoami)
fi

export MYSQL_PWD

if [ -z "$DATABASES" ];then
    DATABASES=$(mysql -u$MYSQL_USER -B -N -L -e "show databases;"|grep -v -e "^information_schema$")
fi

START_DATE=$(date "+%Y-%m-%d_%H%M%S")

for dbname in $DATABASES;do
    echo Backing up database $dbname
    mysqldump -u$MYSQL_USER --routines --add-drop-table --force --default-character-set=utf8 $dbname| gzip| _send_file "${dbname}/${START_DATE}.sql.gz"
done

_finish
