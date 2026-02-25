<template>
  <div class="modern-fm-main flex h-screen overflow-hidden bg-white">
    <!-- 左侧目录树 (参考 Alist) -->
    <aside class="w-72 border-r flex flex-col bg-slate-50">
      <div class="p-4 border-b font-bold flex justify-between items-center bg-white">
        <span>目录树</span>
        <i class="fas fa-sync-alt text-slate-400 cursor-pointer"></i>
      </div>
      <div class="tree-content overflow-y-auto flex-1 p-2">
        <div class="tree-node active">
          <i class="fas fa-chevron-down mr-2 text-slate-400"></i>
          <i class="fas fa-folder text-blue-500 mr-2"></i>
          <span>Root</span>
        </div>
        <div class="tree-node pl-6" v-for="node in mockTree" :key="node">
          <i class="fas fa-chevron-right mr-2 text-slate-300"></i>
          <i class="fas fa-folder text-blue-400 mr-2"></i>
          <span>{{ node }}</span>
        </div>
      </div>
    </aside>

    <!-- 右侧内容 -->
    <main class="flex-1 flex flex-col">
      <!-- 顶部路径 & 按钮 (参考 Alist) -->
      <header class="p-4 flex items-center justify-between border-b">
        <div class="flex items-center space-x-2">
          <i class="fas fa-home text-blue-500"></i>
          <span class="text-slate-400">/</span>
          <span class="font-medium">Backup</span>
          <span class="text-slate-400">/</span>
        </div>
        <div class="flex space-x-2">
          <button class="action-btn text-blue-600"><i class="fas fa-plus mr-1"></i> 添加</button>
          <button class="action-btn text-slate-600"><i class="fas fa-sync mr-1"></i> 刷新</button>
          <button class="action-btn bg-green-500 text-white border-none"><i class="fas fa-play mr-1"></i> Play Folder</button>
        </div>
        <div class="flex items-center space-x-2 border-l pl-4">
          <i class="fas fa-th-large p-2 rounded hover:bg-slate-100 cursor-pointer text-slate-400"></i>
          <i class="fas fa-list p-2 rounded bg-blue-50 text-blue-500 cursor-pointer"></i>
          <i class="fas fa-cog p-2 rounded hover:bg-slate-100 cursor-pointer text-slate-400"></i>
        </div>
      </header>

      <!-- 文件列表 (Alist 风格列表模式) -->
      <div class="flex-1 overflow-y-auto">
        <table class="w-full text-left border-collapse">
          <thead class="sticky top-0 bg-slate-50 border-b text-xs text-slate-500 uppercase tracking-wider">
            <tr>
              <th class="p-3 w-10"><input type="checkbox" class="rounded"></th>
              <th class="p-3 font-semibold">名称 <i class="fas fa-sort-down ml-1"></i></th>
              <th class="p-3 font-semibold w-32">大小</th>
              <th class="p-3 font-semibold w-32">类型</th>
              <th class="p-3 font-semibold w-48">修改时间</th>
            </tr>
          </thead>
          <tbody class="text-sm">
            <tr v-for="file in files" :key="file.name" class="hover:bg-blue-50 group border-b border-slate-50 transition-colors">
              <td class="p-3 text-center"><input type="checkbox" class="rounded border-slate-300"></td>
              <td class="p-3 flex items-center">
                <span class="text-2xl mr-3">
                   <i v-if="file.isDir" class="fas fa-folder text-yellow-500"></i>
                   <i v-else-if="isZip(file.name)" class="fas fa-file-archive text-orange-500"></i>
                   <i v-else-if="isImage(file.name)" class="fas fa-file-image text-blue-500"></i>
                   <i v-else class="fas fa-file text-slate-400"></i>
                </span>
                <span class="text-slate-700 font-medium">{{ file.name }}</span>
              </td>
              <td class="p-3 text-slate-500 font-mono text-xs">{{ file.isDir ? '—' : file.size }}</td>
              <td class="p-3 text-slate-400 text-xs">{{ file.type }}</td>
              <td class="p-3 text-slate-400 font-mono text-xs">{{ file.modTime }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </main>
  </div>
</template>

<style scoped>
.tree-node { padding: 8px 12px; border-radius: 6px; cursor: pointer; font-size: 13px; color: #475569; }
.tree-node.active { background: #e0f2fe; color: #0369a1; font-weight: 600; }
.tree-node:hover:not(.active) { background: #f1f5f9; }
.action-btn { padding: 6px 14px; border: 1px solid #e2e8f0; border-radius: 6px; font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.2s; display: flex; align-items: center; }
.action-btn:hover { background: #f8fafc; border-color: #cbd5e1; }
</style>
