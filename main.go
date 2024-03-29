package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type VideoConverter struct {
	PathFFMpeg    string
	PathCache     string
	PathOutput    string
	PathSeparator string
	PathVideoDirs []string
}

type VideoInfo struct {
	Title     string
	OwnerName string
}

func NewVideoConverter(path_ffmpeg string, path_cache string, path_video_output string) (video_converter *VideoConverter) {
	wd, _ := os.Getwd()
	path_sep := string(os.PathSeparator)
	video_converter = &VideoConverter{
		PathFFMpeg:    wd + path_sep + path_ffmpeg,
		PathCache:     path_cache,
		PathOutput:    path_video_output,
		PathSeparator: path_sep,
	}
	return
}

// 遍历获取每个缓存视频的路径
func (v *VideoConverter) GetVideoDirs() (err error) {
	v.PathVideoDirs = []string{}
	dirs, err := os.ReadDir(v.PathCache)
	if err != nil {
		return
	}
	for _, fi := range dirs {
		if fi.IsDir() {
			path_dir := v.PathCache + v.PathSeparator + fi.Name()
			dirs_next, err := os.ReadDir(path_dir)
			if err != nil {
				return err
			}
			v.PathVideoDirs = append(v.PathVideoDirs, path_dir+v.PathSeparator+dirs_next[0].Name())
		}
	}
	return
}

// 遍历每个视频目录,获取视频文件名称,并转换M4S为MP4文件
func (v *VideoConverter) ConverterVideo() (err error) {
	for index, dir := range v.PathVideoDirs {
		fmt.Printf("[%v]开始转换任务%v", index+1, dir)
		video_info, _ := v.GetVideoInfo(dir + v.PathSeparator + "entry.json")
		// 默认输出文件名为 UP主用户名-视频名称
		output_video_name := video_info.OwnerName + "-" + video_info.Title
		v.ConverterM4sToMp4(dir, output_video_name)
	}
	return
}

// 从视频JSON文件读取视频文件信息
func (v *VideoConverter) GetVideoInfo(path_json string) (video_info *VideoInfo, err error) {
	video_info = new(VideoInfo)
	content, err := os.ReadFile(path_json)
	if err != nil {
		fmt.Println(err)
		return
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(content), &data); err == nil {
		video_info.Title = data["title"].(string)
		video_info.OwnerName = data["owner_name"].(string)
	} else {
		fmt.Println(err)
	}
	return
}

// 调用ffmpeg转换视频文件
func (v *VideoConverter) ConverterM4sToMp4(path_input string, output_video_name string) (err error) {

	flag_path_correct := false
	// 尝试读取80和64文件夹下的m4s文件
	for _, child_path := range []string{"80", "64"} {
		temp_path := path_input + v.PathSeparator + child_path
		_, err = os.ReadDir(temp_path)
		if err == nil {
			flag_path_correct = true
			path_input = temp_path
			break
		}
	}

	if !flag_path_correct {
		fmt.Printf("[!]视频路径错误:%v", path_input)
		return
	}

	path_video := path_input + v.PathSeparator + "video.m4s"
	path_audio := path_input + v.PathSeparator + "audio.m4s"
	path_output := v.PathOutput + v.PathSeparator + v.FilterVideoTitle(output_video_name) + ".mp4"

	args := []string{
		"-i", path_video,
		"-i", path_audio,
		"-c:v", "copy",
		"-c:a", "copy",
		"-strict", "experimental",
		"-y",
		path_output,
		"-hide_banner",
		"-stats",
	}
	cmd := exec.Command(v.PathFFMpeg, args...)

	if err := cmd.Start(); err != nil {
		fmt.Printf("执行FFmpeg命令失败: %s", err)
	} else {
		fmt.Printf("[+]转换完成%v\n", path_output)
	}

	return
}

// 将视频标题中包含的特殊字符进行转换，否则容易导出不了
func (v *VideoConverter) FilterVideoTitle(title string) (result string) {
	title = strings.ReplaceAll(title, "<", "《")
	title = strings.ReplaceAll(title, ">", "》")
	title = strings.ReplaceAll(title, `\`, "#")
	title = strings.ReplaceAll(title, `"`, "'")
	title = strings.ReplaceAll(title, "/", "_")
	title = strings.ReplaceAll(title, "|", "_")
	title = strings.ReplaceAll(title, "?", "_")
	title = strings.ReplaceAll(title, "*", "_")
	title = strings.ReplaceAll(title, "【", "[")
	title = strings.ReplaceAll(title, "】", "]")
	title = strings.ReplaceAll(title, "！", "!")
	title = strings.TrimSpace(title)
	return title
}

func main() {

	// ffmpeg.exe的路径，默认放同个目录下
	path_ffmpeg := "ffmpeg.exe"
	// 缓存视频文件的文件夹
	path_cache := "download"
	// 输出视频文件的文件夹(若不存在请新建一下)
	path_video_output := "output"

	video_converter := NewVideoConverter(path_ffmpeg, path_cache, path_video_output)
	video_converter.GetVideoDirs()
	video_converter.ConverterVideo()

}
