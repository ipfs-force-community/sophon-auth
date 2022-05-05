#!/bin/sh

echo $@
/app/venus-auth run &

sleep 3

echo "regist admin"
/app/venus-auth user add --name="admin"
token=`/app/venus-auth token gen --perm admin admin`
echo ${token#*: }
echo "${token#*: }" > /env/token

wait