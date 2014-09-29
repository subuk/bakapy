
if [ -z "$CPU_PRIORITY" ];then
    CPU_PRIORITY=19
fi


if [ -z "$IO_PRIORITY" ];then
    IO_PRIORITY=3
fi

if [ -z "$GZIP_EXTRA_ARGS" ];then
    GZIP_EXTRA_ARGS='-2'
fi


if [ -z "$BACKUP_TYPE" ];then
    echo "This command requires argument 'backup_type'"
    _fail
    exit 1
fi

if [ -z "$BACKUP_DIRS" ];then
    echo "This command requires argument 'backup_dirs'"
    _fail
    exit 1
fi

if [ -z "$LISTED_INCREMENTAL" ];then
    LISTED_INCREMENTAL=/var/tmp/.backup-directory-listed-incremental
fi

START_DATE=$(date "+%Y-%m-%d_%H%M%S")

echo "PATH: $PATH"
echo "GZIP_EXTRA_ARGS: $GZIP_EXTRA_ARGS"
echo "LISTED_INCREMENTAL: $LISTED_INCREMENTAL"
echo "BACKUP_TYPE: $BACKUP_TYPE"
echo "START_DATE: $START_DATE"
echo "IO_PRIORITY: $IO_PRIORITY"
echo "CPU_PRIORITY: $CPU_PRIORITY"
echo "EXTRA_TAR_OPTS: $EXTRA_TAR_OPTS"
echo "BACKUP_DIRS: $BACKUP_DIRS"

DEFAULT_TAR_OPTS="--numeric-owner --sparse --exclude=$LISTED_INCREMENTAL"


do_full_backup(){
    echo "Starting full backup"
    if [ -f "$LISTED_INCREMENTAL" ];then
        echo "Removing stale listed-incremental file $LISTED_INCREMENTAL"
        rm -f "$LISTED_INCREMENTAL"
    fi
    nice -n"$CPU_PRIORITY" ionice -c"$IO_PRIORITY" /bin/tar \
        --listed-incremental="$LISTED_INCREMENTAL" \
        $DEFAULT_TAR_OPTS \
        $EXTRA_TAR_OPTS \
        --create --file - $BACKUP_DIRS |gzip $GZIP_EXTRA_ARGS | _send_file "${START_DATE}_full.tar.gz"
}

do_diff_backup(){
    echo "Starting differencial backup"
    if [ ! -f "$LISTED_INCREMENTAL" ];then
        echo "Cannot found listed-incremental file, doing full backup"
        do_full_backup
        return
    fi

    cp -f "$LISTED_INCREMENTAL" "${LISTED_INCREMENTAL}.full"
    nice -n"$CPU_PRIORITY" ionice -c"$IO_PRIORITY" /bin/tar \
        --listed-incremental="$LISTED_INCREMENTAL" \
        $DEFAULT_TAR_OPTS \
        $EXTRA_TAR_OPTS \
        --create --file - $BACKUP_DIRS |gzip $GZIP_EXTRA_ARGS | _send_file "${START_DATE}_diff.tar.gz"
    mv -f "${LISTED_INCREMENTAL}.full" "$LISTED_INCREMENTAL"
}

do_inc_backup(){
    echo "Starting incremetal backup"
    if [ ! -f "$LISTED_INCREMENTAL" ];then
        echo "Cannot found listed-incremental file, doing full backup"
        do_full_backup
        return
    fi
    nice -n"$CPU_PRIORITY" ionice -c"$IO_PRIORITY" /bin/tar \
        --listed-incremental="$LISTED_INCREMENTAL" \
        $DEFAULT_TAR_OPTS \
        $EXTRA_TAR_OPTS \
        --create --file - $BACKUP_DIRS |gzip $GZIP_EXTRA_ARGS | _send_file "${START_DATE}_inc.tar.gz"
}

do_${BACKUP_TYPE}_backup

_finish
