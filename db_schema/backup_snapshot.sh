#!/bin/bash
#
# Usage
# ./backup keyspace snapshot
#

for dir in `find /var/lib/cassandra/data/$1 -name snapshots`; do
	echo $dir
	parentdir=$(dirname $dir)
	dirName=$(basename $parentdir)
	pos=`expr index "$dirName" -`
	newDir=${dirName:0:pos-1}
	mkdir -p $2/$newDir
	cp -v $dir/$2/* $2/$newDir/
done

tar czvf $2.tgz $2
rm -r $2
