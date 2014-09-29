
if [ -z "$VHOSTS_DIR" ];then
    echo "This command requires argument 'vhosts_dir'"
    _fail
    exit 1
fi

if [ -z "$BACKUP_TYPE" ];then
    echo "This command requires argument 'backup_type'"
    _fail
    exit 1
fi

if [ -z "$GZIP_EXTRA_ARGS" ];then
    GZIP_EXTRA_ARGS='-2'
fi

if [ -z "$LISTED_INCREMENTAL_DIR" ];then
    LISTED_INCREMENTAL_DIR="$VHOSTS_DIR/.backup-incrementals"
fi

if [ -z "$CPU_PRIORITY" ];then
    CPU_PRIORITY=19
fi

if [ -z "$IO_PRIORITY" ];then
    IO_PRIORITY=3
fi

START_DATE=$(date "+%Y-%m-%d")

echo "PATH: $PATH"
echo "GZIP_EXTRA_ARGS: $GZIP_EXTRA_ARGS"
echo "VHOSTS_DIR: $VHOSTS_DIR"
echo "LISTED_INCREMENTAL_DIR: $LISTED_INCREMENTAL_DIR"
echo "BACKUP_TYPE: $BACKUP_TYPE"
echo "START_DATE: $START_DATE"
echo "IO Priority: $IO_PRIORITY"
echo "CPU Priority: $CPU_PRIORITY"
echo "EXTRA_TAR_OPTS: $EXTRA_TAR_OPTS"

DEFAULT_TAR_OPTS="--numeric-owner --sparse"

do_full_backup_vhost(){
    local vhost="$1"
    echo "Starting full backup for vhost $vhost"
    cd "$VHOSTS_DIR"
    if [ ! -d "$vhost" ];then
        echo "vhost $vhost is not directory, skipping"
        continue
    fi
    if [ -f "$LISTED_INCREMENTAL_DIR/$vhost" ];then
        echo "removing statefile $LISTED_INCREMENTAL_DIR/$vhost"
        rm -f "$LISTED_INCREMENTAL_DIR/$vhost"
    fi

    nice -n"$CPU_PRIORITY" ionice -c"$IO_PRIORITY" /bin/tar \
        $DEFAULT_TAR_OPTS \
        $EXTRA_TAR_OPTS \
        --listed-incremental="$LISTED_INCREMENTAL_DIR/$vhost" \
        --create --file - "$vhost" |gzip $GZIP_EXTRA_ARGS | _send_file "$vhost/${START_DATE}_full.tar.gz"
}

do_diff_backup_vhost(){
    local vhost="$1"
    echo "Starting differencial backup for vhost $vhost"
    cd "$VHOSTS_DIR"
    if [ ! -d "$vhost" ];then
        echo "vhost $vhost is not directory, skipping"
        continue
    fi
    if [ ! -f "$LISTED_INCREMENTAL_DIR/$vhost" ];then
        echo "Cannot found listed-incremental file for vhost $vhost, doing full backup"
        do_full_backup_vhost "$vhost"
        return
    fi

    cp -f "$LISTED_INCREMENTAL_DIR/$vhost" "$LISTED_INCREMENTAL_DIR/$vhost.full"

    nice -n"$CPU_PRIORITY" ionice -c"$IO_PRIORITY" /bin/tar \
        $DEFAULT_TAR_OPTS \
        $EXTRA_TAR_OPTS \
        --listed-incremental="$LISTED_INCREMENTAL_DIR/$vhost" \
        --create --file - "$vhost" |gzip $GZIP_EXTRA_ARGS | _send_file "$vhost/${START_DATE}_diff.tar.gz"

    mv -f "$LISTED_INCREMENTAL_DIR/$vhost.full" "$LISTED_INCREMENTAL_DIR/$vhost"
}

do_inc_backup_vhost(){
    local vhost="$1"
    echo "Starting incremental backup for vhost $vhost"
    cd "$VHOSTS_DIR"
    if [ ! -d "$vhost" ];then
        echo "vhost $vhost is not directory, skipping"
        continue
    fi
    if [ ! -f "$LISTED_INCREMENTAL_DIR/$vhost" ];then
        echo "Cannot found listed-incremental file for vhost $vhost, doing full backup"
        do_full_backup_vhost "$vhost"
        return
    fi
    nice -n"$CPU_PRIORITY" ionice -c"$IO_PRIORITY" /bin/tar \
        $DEFAULT_TAR_OPTS \
        $EXTRA_TAR_OPTS \
        --listed-incremental="$LISTED_INCREMENTAL_DIR/$vhost" \
        --create --file - "$vhost" |gzip $GZIP_EXTRA_ARGS | _send_file "$vhost/${START_DATE}_inc.tar.gz"

}

mkdir -p $LISTED_INCREMENTAL_DIR
chmod 700 $LISTED_INCREMENTAL_DIR

for vhost in $(ls -1 $VHOSTS_DIR);do
    do_${BACKUP_TYPE}_backup_vhost "$vhost"
done
_finish

