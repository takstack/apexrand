#!/bin/sh
# wait-for-mysql.sh

set -e
  
host="$1"
shift
cmd="$@"
  
until mysql -uroot -pmysql --port 3306 '\q'; do
  >&2 echo "Mysql is unavailable - sleeping"
  sleep 1
done
  
>&2 echo "Mysql is up - executing command"
exec $cmd