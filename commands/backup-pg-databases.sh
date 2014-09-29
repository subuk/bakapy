#!/bin/sh
#
#  Args:
#   * pguser
#   * pgpassword
#   * databases
#

export PGUSER
export PGPASSWORD
START_DATE=$(date "+%Y-%m-%d_%H%M%S")

if [ -z "$DATABASES" ];then
    DATABASES=$(psql postgres -At -c 'select datname from pg_database where not datistemplate and datallowconn order by datname')
fi

for db in $DATABASES;do
    /usr/bin/pg_dump -E UTF8 "$db" | gzip|_send_file "$db"/"${START_DATE}.sql.gz"
done

_finish
