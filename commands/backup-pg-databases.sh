#!/bin/sh
#
#  Args:
#   * pguser
#   * pgpassword
#   * databases
#

export PGUSER
export PGPASSWORD

if [ -z "$DATABASES" ];then
    DATABASES=$(psql -l -A -t -d template1|awk -F '|' {'print $1'}|grep -v -e "template1$"  -e "template0$" -e "^postgres$" -e "postgres=CTc/postgres")
fi

for db in $DATABASES;do
    /usr/bin/pg_dump -E UTF8 "$db" |_send_file "$db"/$(date "+%Y-%m-%d").sql
done

_finish
