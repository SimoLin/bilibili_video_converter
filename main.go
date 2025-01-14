package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/elastic/elastic-agent-libs/mapstr"
)

var list_all_video_quality = []string{"120", "116", "112", "80", "64", "32", "16"}

// var dict_all_video_quality = map[string]string{
// 	"120": "4K(超清)",
// 	"116": "1080P(60帧)",
// 	"112": "1080P(高码率)",
// 	"80":  "1080P(高清)",
// 	"64":  "720P(高清)",
// 	"32":  "480P(清晰)",
// 	"16":  "360P(流畅)",
// }

type VideoInfo struct {
	Type       string
	AVID       string
	BVID       string
	CID        string
	Title      string
	OwnerName  string
	Index      string
	IndexTitle string
}

type VideoConverter struct {
	PathFFMpeg          string
	PathDownload        string
	PathOutput          string
	PathSeparator       string
	MapEntryToVideoInfo map[string]*VideoInfo
}

func NewVideoConverter(path_ffmpeg string, path_download string, path_output string) (video_converter *VideoConverter) {
	wd, _ := os.Getwd()
	path_separator := string(os.PathSeparator)
	video_converter = &VideoConverter{
		PathFFMpeg:    wd + path_separator + path_ffmpeg,
		PathDownload:  path_download,
		PathOutput:    path_output,
		PathSeparator: path_separator,
	}
	return
}

// 遍历获取每个缓存视频的路径
func (v *VideoConverter) GetVideoDirsToEntry() (err error) {
	v.MapEntryToVideoInfo = map[string]*VideoInfo{}

	err = filepath.Walk(v.PathDownload, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if info.Name() == "entry.json" {
			video_info, _ := v.GetVideoInfo(path)
			v.MapEntryToVideoInfo[path] = video_info
		}
		return nil
	})

	return
}

// 遍历每个视频目录,获取视频文件名称,并转换M4S为MP4文件
func (v *VideoConverter) ConverteVideo() (err error) {
	count := 0
	for path_to_entry_file, video_info := range v.MapEntryToVideoInfo {
		count += 1
		fmt.Printf("[%v]开始转换任务%v\n", count, path_to_entry_file)
		// 输入文件夹路径为entry.json同级文件夹
		input_video_path, _ := filepath.Split(path_to_entry_file)
		// 默认输出文件名为 {BVID}_{Index}.mp4，不带中文防止ffmpeg保存失败
		output_video_name := video_info.BVID + "_" + video_info.Index
		v.ConverterM4sToMp4(input_video_path, output_video_name)
	}
	return
}

// 全部视频保存完成后，遍历一遍重命名为中文名称
func (v *VideoConverter) RenameVideo() (err error) {
	fmt.Println("[+]等待10S再开始重命名文件，防止文件占用")
	time.Sleep(time.Duration(10 * time.Second))
	fmt.Println("[+]开始重命名视频文件为中文，请耐心等待")
	for _, video_info := range v.MapEntryToVideoInfo {
		path_to_video_dir := ""
		path_to_new_video := ""
		switch video_info.Type {
		case "视频":
			path_to_video_dir = v.PathOutput + v.PathSeparator + video_info.OwnerName
			path_to_new_video = path_to_video_dir + v.PathSeparator + video_info.Title + "_" + video_info.Index + "_" + video_info.IndexTitle + ".mp4"
		case "番剧":
			path_to_video_dir = v.PathOutput + v.PathSeparator + video_info.Title
			path_to_new_video = path_to_video_dir + v.PathSeparator + video_info.Index + "_" + video_info.IndexTitle + ".mp4"
		default:
			fmt.Printf("[!]视频类型错误:%v\n", video_info)
		}
		os.MkdirAll(path_to_video_dir, os.ModePerm)
		output_video_name := video_info.BVID + "_" + video_info.Index
		path_to_old_video := v.PathOutput + v.PathSeparator + output_video_name + ".mp4"
		err = os.Rename(path_to_old_video, path_to_new_video)
		if err != nil {
			fmt.Printf("[!]视频(%v)重命名错误:%v\n", output_video_name, err)
		}
	}
	fmt.Println("[+]重命名完成")
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
		// 通用
		video_info.Title = MustGetValue[string](data, "title")

		// 识别媒体类型
		_, is_page_data := data["page_data"]
		if is_page_data {
			video_info.Type = "视频"
		}
		_, is_ep := data["ep"]
		if is_ep {
			video_info.Type = "番剧"
		}

		switch video_info.Type {
		case "视频":
			video_info.OwnerName = MustGetValue[string](data, "owner_name")
			video_info.AVID = MustGetValue[string](data, "avid")
			video_info.BVID = MustGetValue[string](data, "bvid")
			video_info.CID = MustGetValue[string](data, "page_data.cid")
			video_info.Index = fmt.Sprintf("%v", MustGetValue[float64](data, "page_data.page"))
			video_info.IndexTitle = MustGetValue[string](data, "page_data.part")
		case "番剧":
			video_info.AVID = MustGetValue[string](data, "ep.av_id")
			video_info.BVID = MustGetValue[string](data, "ep.bvid")
			video_info.CID = MustGetValue[string](data, "ep.episode_id")
			video_info.Index = MustGetValue[string](data, "ep.index")
			video_info.IndexTitle = MustGetValue[string](data, "ep.index_title")
		default:
			fmt.Printf("[!]视频类型错误:%v\n", data)
		}

	} else {
		fmt.Println(err)
	}
	return
}

// MustGetValue 读取map[string]interface中的值并返回
func MustGetValue[T any](m map[string]any, key string) (result T) {
	value, err := mapstr.M(m).GetValue(key)
	if value, ok := value.(T); err == nil && ok {
		return value
	}
	return result
}

// ValueInSlice 判断元素是否在 slice 中
//
//	存在返回 true ,不存在返回 false
func ValueInSlice[T comparable](target T, array_list []T) bool {
	for _, item := range array_list {
		if target == item {
			return true
		}
	}
	return false
}

// 调用ffmpeg转换视频文件
func (v *VideoConverter) ConverterM4sToMp4(input_video_path string, output_video_name string) (err error) {

	flag_path_correct := false
	for _, child_path := range list_all_video_quality {
		temp_path := input_video_path + v.PathSeparator + child_path
		_, err = os.ReadDir(temp_path)
		if err == nil {
			flag_path_correct = true
			input_video_path = temp_path
			break
		}
	}

	if !flag_path_correct {
		fmt.Printf("----[!]视频路径错误:%v\n", input_video_path)
		return
	}

	path_video := input_video_path + v.PathSeparator + "video.m4s"
	path_audio := input_video_path + v.PathSeparator + "audio.m4s"
	path_output := v.PathOutput + v.PathSeparator + output_video_name + ".mp4"

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
		fmt.Printf("----[+]转换完成,保存路径为%v\n", path_output)
	}

	return
}

func main() {

	// ffmpeg.exe的路径，默认放同个目录下
	path_ffmpeg := "ffmpeg.exe"
	// 缓存视频文件的文件夹
	path_download := "download"
	// 输出视频文件的文件夹(若不存在请新建一下)
	path_video_output := "output"

	video_converter := NewVideoConverter(path_ffmpeg, path_download, path_video_output)
	video_converter.GetVideoDirsToEntry()
	video_converter.ConverteVideo()
	video_converter.RenameVideo()

}
