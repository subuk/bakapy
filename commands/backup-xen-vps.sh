#
#  ARGS:
#    * XEN_DOMAINS
#    * EXCLUDE_LIST
#

if [ -z "$XEN_DOMAINS" ];then
    XEN_DOMAINS=$(xm list|sed 1,2d|awk '{print $1}')
    if [ -z "$XEN_DOMAINS" ];then
        echo "Cannot get list of xen domains and XEN_DOMAINS arg is not specified"
        exit 1
    fi
fi

for VPS in $XEN_DOMAINS; do
    for excluded_domain in $EXCLUDE_LIST;do
        if [ "$VPS" == "$excluded_domain" ];then
            echo "Domain ${VPS} is excluded"
            continue 2
        fi
    done
    devices=$(xm list --long ${VPS}|grep 'uname phy:'|awk -F')' '{print $1}'|awk '{print $2}'|sed s',phy:,,g')
    for lv_path in $devices;do

        vg_path=$(dirname $lv_path)
        lv_name=$(basename $lv_path)

        snap_name="${lv_name}_backup_snap"
        snap_path="${vg_path}/${snap_name}"

        lvcreate -s -L5G -n "$snap_name" "$lv_path"
        dd if="$snap_path" bs=10M| _send_file "${VPS}/$(date "+%Y-%m-%d")_${lv_name}.img"
        sleep 3
        lvremove -f "$snap_path"
    done

done

_finish
