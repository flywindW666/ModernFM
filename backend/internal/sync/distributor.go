package sync

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// SyncEvent 定义同步事件
type SyncEvent struct {
	Action   string    `json:"action"` // create, update, delete
	Path     string    `json:"path"`
	IsFolder bool      `json:"isFolder"`
	ModTime  time.Time `json:"modTime"`
}

// Node 定义远程节点信息
type RemoteNode struct {
	Addr string `json:"addr"`
	Key  string `json:"key"`
}

// DispatchSync 将操作同步到其它节点
func DispatchSync(nodes []RemoteNode, event SyncEvent) {
	for _, node := range nodes {
		go func(n RemoteNode) {
			payload, _ := json.Marshal(event)
			client := &http.Client{Timeout: 10 * time.Second}
			req, _ := http.NewRequest("POST", n.Addr+"/api/internal/sync", bytes.NewBuffer(payload))
			req.Header.Set("X-Sync-Key", n.Key)
			client.Do(req)
		}(node)
	}
}
