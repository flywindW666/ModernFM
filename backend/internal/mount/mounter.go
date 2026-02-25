package mount

import (
	"fmt"
	"os/exec"
	"runtime"
)

// MountRemote 远程挂载逻辑（支持 SMB/NFS/WebDAV）
type MountConfig struct {
	Type     string `json:"type"` // smb, nfs, webdav
	Remote   string `json:"remote"`
	Local    string `json:"local"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func ExecuteMount(cfg MountConfig) error {
	var cmd *exec.Cmd
	
	switch cfg.Type {
	case "smb":
		// 示例：mount -t cifs //192.168.1.100/share /data/remotes/smb ...
		args := fmt.Sprintf("//%s %s -o username=%s,password=%s", cfg.Remote, cfg.Local, cfg.Username, cfg.Password)
		cmd = exec.Command("mount", "-t", "cifs", args)
	case "webdav":
		// 使用 rclone 或 mount.davfs
		cmd = exec.Command("mount", "-t", "davfs", cfg.Remote, cfg.Local)
	}

	if cmd != nil {
		return cmd.Run()
	}
	return fmt.Errorf("unsupported mount type")
}
