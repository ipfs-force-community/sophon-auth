#!/bin/sh

echo $@
/app/venus-auth run  &

sleep 1

echo "regist admin"
/app/venus-auth user add admin
token=`/app/venus-auth token gen --perm admin admin`
/app/venus-auth  user active admin

echo "token:"
echo ${token#*: }
echo "${token#*: }" > /env/token

wait
