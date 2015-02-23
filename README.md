# Bakapy

[![Build Status](https://travis-ci.org/subuk/bakapy.svg?branch=master)](https://travis-ci.org/subuk/bakapy)

Backup framework

## Features

- Simple - you still write a shell scripts
- Only you decide how to store backup data
- Scripted notifications
- Statistic about each job
- Human readable metadata format
- Backup encryption on client
- Web interface
- Easy to install and configure

## Installation

DEB/RPM packages: https://github.com/subuk/bakapy/releases

## Configuration

See comments in /etc/bakapy/... files.

## Writing custom commands

Each backup job has one command. Command is a shell script for collecting data on the server. Command must use function _send_file for send file to storage. 

The simplest example:

    for d in usr etc root ;do
      tar -cf - /$d | _send_file "main/$d.tar"
    done


## Scheduler expression format

http://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format

A cron expression represents a set of times, using 6 space-separated fields.

    Field name   | Mandatory? | Allowed values  | Allowed special characters
    ----------   | ---------- | --------------  | --------------------------
    Seconds      | Yes        | 0-59            | * / , -
    Minutes      | Yes        | 0-59            | * / , -
    Hours        | Yes        | 0-23            | * / , -
    Day of month | Yes        | 1-31            | * / , - ?
    Month        | Yes        | 1-12 or JAN-DEC | * / , -
    Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?

Note: Month and Day-of-week field values are case insensitive.  "SUN", "Sun",
and "sun" are equally accepted.

### Special Characters

#### Asterisk ( * )
The asterisk indicates that the cron expression will match for all values of the field; e.g., using an asterisk in the 5th field (month) would indicate every month.

#### Slash ( / )
Slashes are used to describe increments of ranges. For example 3-59/15 in the
1st field (minutes) would indicate the 3rd minute of the hour and every 15
minutes thereafter. The form "*\/..." is equivalent to the form "first-last/...", that is, an increment over the largest possible range of the field.  The form "N/..." is accepted as meaning "N-MAX/...", that is, starting at N, use the increment until the end of that specific range.  It does not wrap around. 

#### Comma ( , )
Commas are used to separate items of a list. For example, using "MON,WED,FRI" in the 5th field (day of week) would mean Mondays, Wednesdays and Fridays.

#### Hyphen ( - )
Hyphens are used to define ranges. For example, 9-17 would indicate every
hour between 9am and 5pm inclusive.

#### Question mark ( ? )
Question mark may be used instead of '*' for leaving either day-of-month or
day-of-week blank.
