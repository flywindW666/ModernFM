<script setup>
import { ref, onMounted, computed, watchEffect } from 'vue'

// 1. 状态管理
const themeMode = ref(localStorage.getItem('fm-theme') || 'auto')
const currentPath = ref(localStorage.getItem('fm-last-path') || '')
const files = ref([])
const isLoading = ref(false)
const searchQuery = ref('')

// 2. 主题逻辑
const applyTheme = (mode) => {
  const root = document.documentElement
  const isDark = mode === 'dark' || (mode === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)
  root.classList.toggle('dark', isDark)
}

watchEffect(() => {
  localStorage.setItem('fm-theme', themeMode.value)
  applyTheme(themeMode.value)
})

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
  if (themeMode.value === 'auto') applyTheme('auto')
})

const toggleTheme = (mode) => themeMode.value = mode

// 3. 数据获取 (核心修复)
const fetchFiles = async (path = '') => {
  isLoading.ref = true
  try {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`)
    if (!res.ok) throw new Error('Failed to fetch files')
    const data = await res.json()
    
    // 过滤数据：后端 SQL 使用 LIKE path%，我们需要过滤出当前目录的直接子项
    // 假设 FullPath 为 "dir1/file.txt", path 为 "dir1"
    const normalizedPath = path ? (path.endsWith('/') ? path : path + '/') : ''
    
    files.value = data.filter(f => {
      if (!path) return !f.FullPath.includes('/')
      if (f.FullPath === path) return false // 排除目录自身
      const rel = f.FullPath.slice(normalizedPath.length)
      return rel && !rel.includes('/')
    }).sort((a, b) => (b.IsDir - a.IsDir) || a.Name.localeCompare(b.Name))

    currentPath.value = path
    localStorage.setItem('fm-last-path', path)
  } catch (err) {
    console.error('Fetch error:', err)
  } finally {
    isLoading.value = false
  }
}

// 4. 路径导航
const navigateTo = (path) => fetchFiles(path)
const goBack = () => {
  const parts = currentPath.value.split('/').filter(Boolean)
  parts.pop()
  navigateTo(parts.join('/'))
}

const breadcrumbs = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  return [{ name: 'Root', path: '' }, ...parts.map((p, i) => ({
    name: p,
    path: parts.slice(0, i + 1).join('/')
  }))]
})

// 5. 搜索过滤
const filteredFiles = computed(() => {
  if (!searchQuery.value) return files.value
  const q = searchQuery.value.toLowerCase()
  return files.value.filter(f => f.Name.toLowerCase().includes(q))
})

onMounted(() => {
  fetchFiles(currentPath.value)
})

// 格式化工具
const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateStr) => {
  return new Date(dateStr).toLocaleString('zh-CN', { 
    year: 'numeric', month: '2-digit', day: '2-digit', 
    hour: '2-digit', minute: '2-digit' 
  })
}
</script>

<template>
  <div class="app-container min-h-screen transition-colors duration-300 bg-white dark:bg-[#121212] text-slate-900 dark:text-slate-200 font-sans">
    <div class="flex h-screen overflow-hidden">
      
      <!-- 左侧边栏 -->
      <aside class="w-64 border-r border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-[#1a1a1a] flex flex-col">
        <div class="p-4 flex items-center justify-between border-b border-slate-200 dark:border-slate-800">
          <h1 class="text-xl font-black tracking-tight text-blue-600 dark:text-blue-400 italic">ModernFM</h1>
          <div class="flex bg-slate-200 dark:bg-slate-700 rounded-full p-1 scale-90">
            <button @click="toggleTheme('light')" :class="{'bg-white dark:bg-slate-500 shadow-sm': themeMode === 'light'}" class="w-7 h-7 flex items-center justify-center rounded-full transition-all"><i class="fas fa-sun text-xs"></i></button>
            <button @click="toggleTheme('dark')" :class="{'bg-white dark:bg-slate-500 shadow-sm': themeMode === 'dark'}" class="w-7 h-7 flex items-center justify-center rounded-full transition-all ml-1"><i class="fas fa-moon text-xs"></i></button>
            <button @click="toggleTheme('auto')" :class="{'bg-white dark:bg-slate-500 shadow-sm': themeMode === 'auto'}" class="px-2 h-7 flex items-center justify-center rounded-full transition-all ml-1 text-[10px] font-bold">Auto</button>
          </div>
        </div>

        <nav class="flex-1 overflow-y-auto p-3 space-y-1">
           <div @click="navigateTo('')" :class="{'bg-blue-600 text-white shadow-lg shadow-blue-500/30': currentPath === ''}" 
                class="flex items-center p-2.5 rounded-xl cursor-pointer transition-all hover:bg-blue-500/10 group">
             <i class="fas fa-database mr-3 w-5 text-center" :class="currentPath === '' ? 'text-white' : 'text-blue-500'"></i>
             <span class="text-sm font-semibold">所有文件</span>
           </div>
           <div class="pt-4 pb-2 px-2 text-[10px] font-bold uppercase tracking-widest text-slate-400">快速访问</div>
           <div class="flex items-center p-2.5 rounded-xl cursor-pointer hover:bg-slate-200 dark:hover:bg-slate-800 transition-all text-slate-600 dark:text-slate-400">
             <i class="fas fa-clock mr-3 w-5 text-center"></i><span class="text-sm">最近使用</span>
           </div>
           <div class="flex items-center p-2.5 rounded-xl cursor-pointer hover:bg-slate-200 dark:hover:bg-slate-800 transition-all text-slate-600 dark:text-slate-400">
             <i class="fas fa-star mr-3 w-5 text-center"></i><span class="text-sm">我的收藏</span>
           </div>
        </nav>
      </aside>

      <!-- 右侧主内容 -->
      <main class="flex-1 flex flex-col bg-white dark:bg-[#121212]">
        <!-- 顶部工具栏 -->
        <header class="h-16 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-6">
          <div class="flex items-center space-x-1 overflow-hidden">
             <template v-for="(bc, idx) in breadcrumbs" :key="bc.path">
               <span v-if="idx > 0" class="text-slate-300 dark:text-slate-700 px-1">/</span>
               <button @click="navigateTo(bc.path)" 
                       class="px-2 py-1 rounded-md hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors text-sm font-medium whitespace-nowrap"
                       :class="idx === breadcrumbs.length - 1 ? 'text-blue-600 dark:text-blue-400' : 'text-slate-500'">
                 {{ bc.name }}
               </button>
             </template>
          </div>

          <div class="flex items-center space-x-4">
            <div class="relative group">
              <i class="fas fa-search absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors"></i>
              <input v-model="searchQuery" type="text" placeholder="搜索文件..." 
                     class="w-48 pl-9 pr-4 py-2 bg-slate-100 dark:bg-slate-800 border-none rounded-xl text-xs focus:ring-2 focus:ring-blue-500/50 transition-all">
            </div>
            <button class="w-9 h-9 flex items-center justify-center rounded-xl bg-blue-600 text-white hover:bg-blue-700 shadow-lg shadow-blue-500/20"><i class="fas fa-upload text-sm"></i></button>
          </div>
        </header>

        <!-- 文件展示区 -->
        <div class="flex-1 overflow-y-auto relative">
          <div v-if="isLoading" class="absolute inset-0 flex items-center justify-center bg-white/50 dark:bg-black/50 z-10">
            <i class="fas fa-circle-notch fa-spin text-3xl text-blue-500"></i>
          </div>

          <table class="w-full border-collapse">
            <thead class="sticky top-0 bg-slate-50/90 dark:bg-[#181818]/90 backdrop-blur-md border-b border-slate-200 dark:border-slate-800 z-[5]">
              <tr class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">
                <th class="p-4 w-12 text-center"><input type="checkbox" class="rounded border-slate-300 dark:border-slate-700 accent-blue-500"></th>
                <th class="p-4 text-left">名称</th>
                <th class="p-4 text-left w-32">大小</th>
                <th class="p-4 text-left w-48">修改时间</th>
                <th class="p-4 w-16"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50">
              <tr v-if="currentPath" @dblclick="goBack" class="hover:bg-slate-50 dark:hover:bg-slate-800/30 cursor-pointer group transition-colors">
                <td class="p-4 text-center"></td>
                <td class="p-4 flex items-center font-bold text-blue-500">
                  <i class="fas fa-level-up-alt mr-4 -rotate-90"></i> ..
                </td>
                <td colspan="3"></td>
              </tr>

              <tr v-for="file in filteredFiles" :key="file.FullPath" 
                  @dblclick="file.IsDir ? navigateTo(file.FullPath) : null"
                  class="group hover:bg-blue-50/50 dark:hover:bg-blue-900/10 cursor-pointer transition-all">
                <td class="p-4 text-center">
                  <input type="checkbox" class="rounded border-slate-300 dark:border-slate-700 accent-blue-500">
                </td>
                <td class="p-4">
                  <div class="flex items-center space-x-4">
                    <div class="w-10 h-10 rounded-xl flex items-center justify-center text-xl transition-transform group-hover:scale-110 shadow-sm"
                         :class="file.IsDir ? 'bg-amber-100 dark:bg-amber-900/20 text-amber-500' : 'bg-slate-100 dark:bg-slate-800 text-slate-400'">
                      <i :class="file.IsDir ? 'fas fa-folder' : 'fas fa-file'"></i>
                    </div>
                    <div class="flex flex-col min-w-0">
                      <span class="text-sm font-semibold truncate group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">{{ file.Name }}</span>
                      <span class="text-[10px] text-slate-400">{{ file.Extension || (file.IsDir ? 'Folder' : 'File') }}</span>
                    </div>
                  </div>
                </td>
                <td class="p-4 text-xs font-mono text-slate-500">{{ file.IsDir ? '—' : formatSize(file.Size) }}</td>
                <td class="p-4 text-xs font-mono text-slate-400">{{ formatDate(file.ModTime) }}</td>
                <td class="p-4 text-right opacity-0 group-hover:opacity-100 transition-opacity">
                   <button class="p-2 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-lg"><i class="fas fa-ellipsis-v text-slate-400"></i></button>
                </td>
              </tr>
            </tbody>
          </table>
          
          <!-- 空状态 -->
          <div v-if="!isLoading && filteredFiles.length === 0" class="flex flex-col items-center justify-center py-20 opacity-30">
            <i class="fas fa-folder-open text-6xl mb-4"></i>
            <p class="text-sm font-medium">当前目录为空</p>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<style>
@import url('https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css');

::-webkit-scrollbar { width: 5px; height: 5px; }
::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 10px; }
.dark ::-webkit-scrollbar-thumb { background: #334155; }
::-webkit-scrollbar-track { background: transparent; }

body { selection-background-color: #3b82f6; }
</style>

<style>
/* 注入 Tailwind 不支持的自定义滚动条 */
::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 10px; }
.dark ::-webkit-scrollbar-thumb { background: #334155; }
::-webkit-scrollbar-track { background: transparent; }

.tree-item { font-size: 13px; font-weight: 500; }
</style>
