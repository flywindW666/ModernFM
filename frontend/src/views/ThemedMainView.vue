<script setup>
import { ref, onMounted, watch } from 'vue'
import RecursiveTree from '../components/RecursiveTree.vue'
import { Folder, FileText, Settings, RefreshCw, ChevronRight, Download, Trash2, Edit3, Grid, List } from 'lucide-vue-next'

const currentPath = ref(localStorage.getItem('fm-last-path') || '')
const files = ref([])
const treeData = ref([])
const isLoading = ref(false)
const viewMode = ref('grid') // 'grid' or 'list'

// --- 目录树核心逻辑 ---

const fetchSubFolders = async (path) => {
  try {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`)
    if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`)
    const data = await res.json()
    // 强制过滤出文件夹，排除隐藏目录（以点开头的）以提升目录树整洁度
    return data.filter(f => f.IsDir && !f.Name.startsWith('.')).sort((a, b) => a.Name.localeCompare(b.Name))
  } catch (e) {
    console.error(`Fetch subfolders failed for ${path}:`, e)
    return []
  }
}

const initializeTree = async () => {
  console.log('Initializing tree...');
  try {
    const rootSubs = await fetchSubFolders('')
    console.log('Root subs:', rootSubs);
    treeData.value = [{
      Name: '资源库',
      FullPath: '',
      IsDir: true,
      isOpen: true,
      loaded: true,
      children: rootSubs.map(f => ({ 
        Name: f.Name,
        FullPath: f.FullPath,
        IsDir: true,
        children: [], 
        isOpen: false, 
        loaded: false 
      }))
    }]
    console.log('Tree data initialized:', treeData.value);
  } catch (e) {
    console.error('Tree init failed', e)
  }
}

const toggleFolder = async (node) => {
  node.isOpen = !node.isOpen
  if (node.isOpen && !node.loaded) {
    node.loading = true
    try {
      const subs = await fetchSubFolders(node.FullPath)
      node.children = subs.map(f => ({ ...f, children: [], isOpen: false, loaded: false }))
      node.loaded = true
    } finally {
      node.loading = false
    }
  }
}

// --- 文件列表逻辑 ---

const navigateTo = async (path) => {
  currentPath.value = path
  localStorage.setItem('fm-last-path', path)
  isLoading.value = true
  try {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`)
    const data = await res.json()
    files.value = data.sort((a, b) => (b.IsDir - a.IsDir) || a.Name.localeCompare(b.Name))
  } catch (e) {
    console.error('Fetch files failed', e)
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  navigateTo(currentPath.value)
  initializeTree()
})

// --- 辅助功能 ---
const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<template>
  <div class="flex h-full w-full overflow-hidden bg-white dark:bg-[#0d1117] text-slate-900 dark:text-slate-100 transition-colors duration-300">
    
    <!-- 左侧目录树 (Requirement 1: Alist Style Sidebar) -->
    <aside class="w-72 border-r border-slate-200 dark:border-slate-800 flex flex-col shrink-0">
      <div class="h-14 flex items-center px-5 border-b border-slate-100 dark:border-slate-800 shrink-0">
        <span class="font-bold text-lg tracking-tight italic text-blue-600">ModernFM</span>
      </div>
      
      <div class="flex-1 overflow-y-auto p-3 custom-scrollbar">
        <RecursiveTree 
          v-for="node in treeData" 
          :key="node.FullPath" 
          :node="node" 
          :currentPath="currentPath"
          :navigateTo="navigateTo"
          :toggleFolder="toggleFolder"
        />
      </div>
    </aside>

    <!-- 主面板 -->
    <main class="flex-1 flex flex-col min-w-0">
      <!-- 顶部导航 -->
      <header class="h-14 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between px-6 shrink-0 bg-white/80 dark:bg-[#0d1117]/80 backdrop-blur-md sticky top-0 z-10">
        <div class="flex items-center gap-2 overflow-hidden">
          <div class="flex items-center gap-1 text-sm text-slate-500">
            <span class="cursor-pointer hover:text-blue-500" @click="navigateTo('')">根目录</span>
            <template v-for="(part, i) in currentPath.split('/').filter(Boolean)" :key="i">
              <ChevronRight class="w-3.5 h-3.5 opacity-40" />
              <span class="truncate max-w-[100px]">{{ part }}</span>
            </template>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <button @click="viewMode = viewMode === 'grid' ? 'list' : 'grid'" class="p-2 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg text-slate-500">
            <component :is="viewMode === 'grid' ? List : Grid" class="w-4 h-4" />
          </button>
          <button @click="navigateTo(currentPath)" class="p-2 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg text-slate-500">
            <RefreshCw class="w-4 h-4" :class="{'animate-spin': isLoading}" />
          </button>
        </div>
      </header>

      <!-- 文件内容区域 -->
      <div class="flex-1 overflow-y-auto p-6 custom-scrollbar relative">
        <div v-if="isLoading" class="absolute inset-0 flex items-center justify-center bg-white/40 dark:bg-black/40 z-20 backdrop-blur-sm">
          <RefreshCw class="w-8 h-8 animate-spin text-blue-500" />
        </div>

        <!-- 网格视图 -->
        <div v-if="viewMode === 'grid'" class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8 gap-4">
          <div 
            v-for="file in files" 
            :key="file.FullPath"
            class="group p-4 bg-slate-50 dark:bg-slate-900/40 rounded-2xl border border-transparent hover:border-blue-500/30 hover:bg-white dark:hover:bg-slate-800 transition-all cursor-pointer text-center"
            @dblclick="file.IsDir ? navigateTo(file.FullPath) : null"
          >
            <div class="aspect-square flex items-center justify-center mb-3">
              <component 
                :is="file.IsDir ? Folder : FileText" 
                class="w-10 h-10" 
                :class="file.IsDir ? 'text-amber-500/80' : 'text-blue-400'" 
              />
            </div>
            <p class="text-xs font-medium truncate px-1 text-slate-700 dark:text-slate-300">{{ file.Name }}</p>
          </div>
        </div>

        <!-- 列表视图 -->
        <div v-else class="space-y-1">
          <div class="flex items-center text-[11px] font-bold text-slate-400 uppercase tracking-wider px-4 mb-2">
            <div class="flex-1">名称</div>
            <div class="w-32 text-right">大小</div>
            <div class="w-48 text-right">修改时间</div>
          </div>
          <div 
            v-for="file in files" 
            :key="file.FullPath"
            class="group flex items-center px-4 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-800/60 rounded-xl cursor-pointer"
            @dblclick="file.IsDir ? navigateTo(file.FullPath) : null"
          >
            <component :is="file.IsDir ? Folder : FileText" class="w-4 h-4 mr-3" :class="file.IsDir ? 'text-amber-500' : 'text-blue-400'" />
            <span class="flex-1 text-sm truncate font-medium">{{ file.Name }}</span>
            <span class="w-32 text-right text-xs text-slate-400">{{ file.IsDir ? '-' : formatSize(file.Size) }}</span>
            <span class="w-48 text-right text-xs text-slate-400">{{ new Date(file.ModTime).toLocaleString() }}</span>
          </div>
        </div>
        
        <!-- 空状态 -->
        <div v-if="!isLoading && files.length === 0" class="h-64 flex flex-col items-center justify-center text-slate-400 italic">
          <Folder class="w-12 h-12 mb-3 opacity-20" />
          <p>当前目录为空</p>
        </div>
      </div>
    </main>
  </div>
</template>
