package transcode

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// StreamVideo 实时调用 ffmpeg 进行 HLS 或 MP4 转码流输出
func StreamVideo(c *gin.Context) {
	relPath := c.Query("path")
	fullPath := filepath.Join("/data", relPath)

	// 设置响应头为流媒体格式
	c.Header("Content-Type", "video/mp4")
	c.Header("Transfer-Encoding", "chunked")

	// FFmpeg 实时转码指令 (示例：转为浏览器兼容的 H264)
	// -i: 输入文件
	// -vcodec libx264: 视频编码
	// -acodec aac: 音频编码
	// -f mp4: 输出格式
	// -movflags frag_keyframe+empty_moov: 关键参数，允许 MP4 流式传输
	cmd := exec.Command("ffmpeg", "-i", fullPath, 
		"-vcodec", "libx264", "-preset", "ultrafast", 
		"-acodec", "aac", "-b:a", "128k", 
		"-f", "mp4", "-movflags", "frag_keyframe+empty_moov", "pipe:1")

	// 将 ffmpeg 的标准输出直接导向 Gin 的响应体
	cmd.Stdout = c.Writer
	if err := cmd.Start(); err != nil {
		c.JSON(500, gin.H{"error": "转码启动失败: " + err.Error()})
		return
	}
	
	cmd.Wait()
}

// GeneratePlayerLink 生成第三方播放器跳转链接
func GeneratePlayerLink(path, host string, playerType string) string {
	streamUrl := fmt.Sprintf("http://%s/api/files/download?path=%s", host, path)
	
	switch playerType {
	case "vlc":
		return "vlc://" + streamUrl
	case "infuse":
		return "infuse://x-callback-url/play?url=" + streamUrl
	case "nplayer":
		return "nplayer-" + streamUrl
	case "potplayer":
		return "potplayer://" + streamUrl
	}
	return streamUrl
}
