#
#  Args:
#    * backup_type
#    * listed_incremental
#    * extra_tar_opts
#

if [ -z "$BACKUP_TYPE" ];then
    echo "This command requires argument 'backup_type'"
    _fail
    exit 1
fi

if [ -z "$LISTED_INCREMENTAL" ];then
    LISTED_INCREMENTAL=/root/.k-backup-system-listed-incremental
    echo 'Assuming /root/.k-backup-system-meta as listed_incremental_dir argument'
fi

DEFAULT_TAR_OPTS="--numeric-owner --sparse --create --file - \
--exclude=/tmp \
--exclude=/dev \
--exclude=/proc \
--exclude=/sys \
--exclude=/var/run/ \
--exclude=/var/tmp \
--exclude=/srv/db/mysql \
--exclude=/srv/db/pgsql \
--exclude=/var/lock \
--exclude=/var/cache \
--exclude=/var/log \
--exclude=/var/spool/postfix \
--exclude=/var/lib/mysql \
--exclude=/var/lib/pgsql \
--exclude=/var/lib/nginx/tmp \
--exclude=/var/lib/php \
--exclude=$LISTED_INCREMENTAL
"

START_DATE=$(date "+%Y-%m-%d")

do_full_backup(){
    echo "Starting full backup"
    if [ -f "$LISTED_INCREMENTAL" ];then
        echo "removing stale listed-incremental file $LISTED_INCREMENTAL"
        rm -f "$LISTED_INCREMENTAL"
    fi
    nice -n19 ionice -c3 /bin/tar --listed-incremental="$LISTED_INCREMENTAL" $DEFAULT_TAR_OPTS $EXTRA_TAR_OPTS / | _send_file "${START_DATE}_full.tar"
}

do_diff_backup(){
    vhost="$1"
    echo "Starting differencial backup"
    if [ ! -f "$LISTED_INCREMENTAL" ];then
        echo "Cannot found listed-incremental file, doing full backup"
        do_full_backup
        return
    fi

    cp -f "$LISTED_INCREMENTAL" "${LISTED_INCREMENTAL}.full"
    nice -n19 ionice -c3 /bin/tar --listed-incremental="$LISTED_INCREMENTAL" $DEFAULT_TAR_OPTS $EXTRA_TAR_OPTS / | _send_file "${START_DATE}_diff.tar"
    mv -f "${LISTED_INCREMENTAL}.full" "$LISTED_INCREMENTAL"
}

do_inc_backup(){
    if [ ! -f "$LISTED_INCREMENTAL" ];then
        echo "Cannot found listed-incremental file, doing full backup"
        do_full_backup "$vhost"
        return
    fi
    nice -n19 ionice -c3 /bin/tar --listed-incremental="$LISTED_INCREMENTAL" $DEFAULT_TAR_OPTS $EXTRA_TAR_OPTS / | _send_file "${START_DATE}_inc.tar"
}

mkdir -p $LISTED_INCREMENTAL_DIR
chmod 700 $LISTED_INCREMENTAL_DIR

do_${BACKUP_TYPE}_backup

_finish

