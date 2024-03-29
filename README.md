# bilibili_video_converter

一个简单的 哔哩哔哩 bilibili APP 缓存视频 批量转码脚本工具(m4s 转 mp4)

> 即将 video.m4s 与 audio.m4s 合并转码为一个 video.mp4

# 原理说明

遍历视频缓存目录，调用 ffmpeg 进行转码

> 源码没什么有价值的东西，开源的代码自己看很安全，请放心使用

# 使用方法

1、下载 ffmpeg.exe 文件

> 请务必从官网下载，否则无法保证安全性!

https://ffmpeg.org/download.html

2、从手机/平板拷贝 哔哩哔哩 APP 视频缓存文件夹

```
# 仅需复制 download 文件夹(整个), 路径一般为:
/Android/data/tv.danmaku.bili/download

# 创建 download 文件夹(若不存在)
# 复制整个 download 文件夹到脚本同目录下
# 目录中有很多 9~10 位纯数字的文件夹
# 每一个文件夹是一个缓存视频
# 所以你也可以按需复制 download 文件夹里要转码的文件夹

# 创建 output 文件夹(若不存在)
```

![复制download文件夹](https://github.com/SimoLin/bilibili_video_converter/blob/main/image/download.png)

3、使用本脚本进行批量视频转码

```
# 下载编译好的二进制文件直接运行
# 请确保 fffmpeg.exe、download、output都在同一个目录下
.\bilibili_video_converter.exe

# 怕不安全可以按需修改源码后，直接运行(需要安装Go环境)
go run main.go
```

4、当然，你也可以使用 ffmpeg.exe 进行单个视频转码

```
# 直接使用 CMD 执行以下命令
.\fffmpeg.exe -i path\to\video.m4s -i path\to\audio.m4s -c:v copy -c:a copy -strict experimental -y path\to\output_video.mp4 -hide_banner -stats

# 替换路径后, 比如
.\fffmpeg.exe -i download\152348569\c1514563463\80\video.m4s -i download\152348569\c1514563463\80\audio.m4s -c:v copy -c:a copy -strict experimental -y output\输出视频.mp4 -hide_banner -stats

```

# 特别鸣谢/参考项目

https://github.com/mzky/m4s-converter

> 工具直接运行报错 ? 看源码功能挺多，改起来嫌麻烦，直接学习(抄过来)关键代码，仅保留 m4s 转码 mp4 的功能
