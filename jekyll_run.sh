#!/usr/bin/env bash
#--destination _deploy

#scp -Cr _site root@115.28.94.157:/home/tech.mojotv.cn
rm -rf _config.yml;
cp _config_code.yml _config.yml;

if [ $1 = 'build' ] ;then
    echo jekyll building;
    bundle exec jekyll build;
    #scp -Cr _site root@115.28.94.157:/home/tech.mojotv.cn;
    #scp -Cr _site root@115.28.94.157:/home/tech.mojotv.cn;
    curl -H 'Content-Type:text/plain' --data-binary @_site/sitemap.txt "http://data.zz.baidu.com/urls?appid=1573826274415344&token=uQb9Q3G0AFzKOmIM&type=batch";
    curl -H 'Content-Type:text/plain' --data-binary @_site/sitemap.txt "http://data.zz.baidu.com/urls?appid=1573826274415344&token=uQb9Q3G0AFzKOmIM&type=realtime";
fi

if [ $1 = 'serve' ] ;then
    echo jekyll serving;
    bundle exec jekyll serve;
fi
rm -rf _config.yml;
cp _config_tech.yml _config.yml;
