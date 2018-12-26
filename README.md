### 博客系统和golang和Hacknews机器人

地址：[MojoTV.cn](https://mojotv.cn)

### 安装说明

1. fork库到自己的github
2. 修改名字为：`username.github.io` 或者 `organizationName.github.io`
3. clone库到本地，参考`_posts`中的目录结构自己创建适合自己的文章目录结构
4. 修改CNAME，或者删掉这个文件，使用默认域名
5. 修改`_config.yml`配置项
6. It's done!
7. 定义自己google 统计 adsense  `_includes/header.html`  第29~48行
8. 自定义自己的评论系统  `_layouts/post.html` 第15行
9. 站长管理平添校验代码  `_includes/header.html` 第14~20行
 
### 分支说明

- 三栏布局（master分支，基于[3-Jekyll](https://github.com/P233/3-Jekyll)）
- 三栏布局 (bootstrap-based分支，基于Bootstrap)
- 单栏布局（first-ui分支，基于Bootstrap）

## 这个项目还包含了以个go语言的机器人来每天定时翻译hacknews新闻

### 需要redis数据库
### 需要服务器或者电脑才能运行,在Github 上是不能执行的

### go语言代码同学们,你可以自己修改,或者删除

## 遇到的bug

### `{{` 于jekyll 的语法冲突
{{raw} {{endraw}}包裹冲突的代码