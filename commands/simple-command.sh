
if [ -z "$BACKUP_DIRS" ];then
    echo "This command requires backup_dirs var"
    exit 1
fi

for d in $BACKUP_DIRS;do
  tar -cf - $d | _send_file "main/$d.tar"
done
_finish
