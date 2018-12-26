---
layout: post
title: Hacknews{{.Day}}新闻
category: Hacknews
tags: hacknews
keywords: hacknews
coverage: hacknews-banner.jpg
---

Hacker News 是一家关于计算机黑客和创业公司的社会化新闻网站，由保罗·格雷厄姆的创业孵化器 Y Combinator 创建。
与其它社会化新闻网站不同的是 Hacker News 没有踩或反对一条提交新闻的选项（不过评论还是可以被有足够 Karma 的用户投反对票）；只可以赞或是完全不投票。简而言之，Hacker News 允许提交任何可以被理解为“任何满足人们求知欲”的新闻。

## HackNews Hack新闻

{{range .News}}
- [{{.TitleEn}}]({{.Url}})
- `{{.TitleZh}}`{{end}}


## HackShows Hacks展示
{{range .Shows}}
- [{{.TitleEn}}]({{.Url}})
- `{{.TitleZh}}`{{end}}


