Bakapy
======

Backup framework.

How to use:
- Write shell script for backup data (command)
- Create job configuration with command, schedule and expire date for files created by this command
- View reports about backup jobs (bakapy-show-meta storage_dir/*)

Installation
------------

DEB-based distros:

    wget https://github.com/subuk/bakapy/releases/download/v${version}/bakapy_${version}_amd64.${debianRelease}.deb
    dpkg -i bakapy_${version}_amd64.${debianRelease}.deb

RPM-based:

    rpm -ivh https://github.com/subuk/bakapy/releases/download/v${version}/bakapy-${version}-1.${dist}.src.rpm

Configuration
-------------

Configuration examples:
- bakapy.conf.ex.yaml
- jobs.conf.ex.yaml

Writing custom commands
-----------------------

Each backup job has one command. Command is a shell script for collecting data on the server. Command must use function _send_file for send file to storage. 

The simplest example:

    for d in usr etc root ;do
      tar -cf - /$d | _send_file "main/$d.tar"
    done
