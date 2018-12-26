#!/bin/bash

app="dejavuzhou.github.io"
ps -aux|grep $app |grep -v grep|cut -c 9-15|xargs kill -9

rm -rf $app
rm -rf gitpage.log
go build
nohup ./$app > ./gitpage.log 2>&1 &
ps -ef | grep $app

