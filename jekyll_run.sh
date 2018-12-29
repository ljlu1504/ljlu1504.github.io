#!/usr/bin/env bash
#--destination _deploy

#scp -Cr _site root@115.28.94.157:/home/tech.mojotv.cn
rm -rf _config.yml;
cp _config_code.yml _config.yml;

if [ $1 = 'build' ] ;then
    echo jekyll building;
    bundle exec jekyll build;
    #scp -Cr _site root@115.28.94.157:/home/tech.mojotv.cn;
fi

if [ $1 = 'serve' ] ;then
    echo jekyll serving;
    bundle exec jekyll serve;
fi
rm -rf _config.yml;
cp _config_tech.yml _config.yml;
