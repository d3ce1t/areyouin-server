#!/bin/bash
echo Restoring $1 keyspace from snapshot $2
for dir in /var/lib/cassandra/data/$1/*/; do
	baseName=$(basename $dir)
	echo $baseName
	pos=`expr index "$baseName" -`
	dirName=${baseName:0:pos-1}
	echo $dirName
	#src=${d}snapshots/$1
	#dst=$d
	#echo Copy $src into $dst
	#cp -v $src/* $dst/
	#chown -R cassandra:cassandra $dst
done
