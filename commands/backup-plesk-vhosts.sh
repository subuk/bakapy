
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

if [ -z "$LISTED_INCREMENTAL_DIR" ];then
    LISTED_INCREMENTAL_DIR=/root/.k-backup-vhosts-meta
    echo 'Autousing /root/.k-backup-vhosts-meta as listed_incremental_dir argument'
fi

START_DATE=$(date "+%Y-%m-%d")

do_full_backup_vhost(){
    vhost="$1"
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
    nice -n19 ionice -c3 /bin/tar --numeric-owner --listed-incremental="$LISTED_INCREMENTAL_DIR/$vhost" \
        --exclude=${vhost}/tmp \
        --exclude=${vhost}/bin \
        --exclude=${vhost}/usr \
        --exclude=${vhost}/lib \
        --exclude=${vhost}/lib64 \
        --exclude=${vhost}/var \
        --exclude=${vhost}/dev \
        --create --sparse --file - "$vhost" | _send_file "$vhost/${START_DATE}_full.tar"


}

do_diff_backup_vhost(){
    vhost="$1"
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

    nice -n19 ionice -c3 /bin/tar --numeric-owner --listed-incremental="$LISTED_INCREMENTAL_DIR/$vhost" \
        --exclude=${vhost}/tmp \
        --exclude=${vhost}/bin \
        --exclude=${vhost}/usr \
        --exclude=${vhost}/lib \
        --exclude=${vhost}/lib64 \
        --exclude=${vhost}/var \
        --exclude=${vhost}/dev \
        --create --sparse --file - "$vhost" | _send_file "$vhost/${START_DATE}_diff.tar"

    mv -f "$LISTED_INCREMENTAL_DIR/$vhost.full" "$LISTED_INCREMENTAL_DIR/$vhost"

}

do_inc_backup_vhost(){
    vhost="$1"
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
    nice -n19 ionice -c3 /bin/tar --numeric-owner --listed-incremental="$LISTED_INCREMENTAL_DIR/$vhost" \
        --exclude=${vhost}/tmp \
        --exclude=${vhost}/bin \
        --exclude=${vhost}/usr \
        --exclude=${vhost}/lib \
        --exclude=${vhost}/lib64 \
        --exclude=${vhost}/var \
        --exclude=${vhost}/dev \
        --exclude=${vhost}/statistics/webstat \
        --create --sparse --file - "$vhost" | _send_file "$vhost/${START_DATE}_inc.tar"

}

mkdir -p $LISTED_INCREMENTAL_DIR
chmod 700 $LISTED_INCREMENTAL_DIR

for vhost in $(ls -1 $VHOSTS_DIR);do
    do_${BACKUP_TYPE}_backup_vhost "$vhost"
done
_finish

