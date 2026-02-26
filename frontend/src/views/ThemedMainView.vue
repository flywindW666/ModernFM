<script setup>
import { ref, onMounted, computed, watchEffect } from 'vue'

// 1. 状态管理
const themeMode = ref(localStorage.getItem('fm-theme') || 'auto')
const currentPath = ref(localStorage.getItem('fm-last-path') || '')
const files = ref([])
const treeData = ref([]) // 目录树数据
const isLoading = ref(false)
const searchQuery = ref('')
const isMobileMenuOpen = ref(false) // 移动端菜单控制

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

const toggleTheme = (mode) => themeMode.value = mode

// 3. 数据获取
// 获取主列表文件
const fetchFiles = async (path = '') => {
  isLoading.value = true
  try {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`)
    if (!res.ok) throw new Error('Failed to fetch files')
    const data = await res.json()
    
    // 过滤出当前目录的直接子项
    const normalizedPath = path ? (path.endsWith('/') ? path : path + '/') : ''
    files.value = data.filter(f => {
      if (!path) return !f.FullPath.includes('/')
      if (f.FullPath === path) return false
      const rel = f.FullPath.slice(normalizedPath.length)
      return rel && !rel.includes('/')
    }).sort((a, b) => (b.IsDir - a.IsDir) || a.Name.localeCompare(b.Name))

    currentPath.value = path
    localStorage.setItem('fm-last-path', path)
    isMobileMenuOpen.value = false // 切换路径后关闭移动端菜单
  } catch (err) {
    console.error('Fetch error:', err)
  } finally {
    isLoading.value = false
  }
}

// 获取所有文件夹构建目录树 (简单实现，全量获取后前端构建)
const fetchTree = async () => {
  try {
    const res = await fetch('/api/files/list?path=') // 获取全量进行前端过滤
    const data = await res.json()
    const folders = data.filter(f => f.IsDir).sort((a, b) => a.FullPath.localeCompare(b.FullPath))
    
    // 构建树形结构
    const tree = []
    const map = { '': { name: 'Root', path: '', children: tree, isOpen: true } }
    
    folders.forEach(f => {
      const parts = f.FullPath.split('/')
      const name = parts.pop()
      const parentPath = parts.join('/')
      const node = { name, path: f.FullPath, children: [], isOpen: false }
      map[f.FullPath] = node
      if (map[parentPath]) {
        map[parentPath].children.push(node)
      } else {
        // 如果父节点还没入 map，说明是顶级或异常，直接放 root
        tree.push(node)
      }
    })
    treeData.value = [{ name: 'Root', path: '', children: tree, isOpen: true }]
  } catch (err) {
    console.error('Tree fetch error:', err)
  }
}

// 4. 导航逻辑
const navigateTo = (path) => fetchFiles(path)
const toggleFolder = (node) => { node.isOpen = !node.isOpen }

const breadcrumbs = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  return [{ name: 'Root', path: '' }, ...parts.map((p, i) => ({
    name: p,
    path: parts.slice(0, i + 1).join('/')
  }))]
})

const filteredFiles = computed(() => {
  if (!searchQuery.value) return files.value
  const q = searchQuery.value.toLowerCase()
  return files.value.filter(f => f.Name.toLowerCase().includes(q))
})

onMounted(() => {
  fetchFiles(currentPath.value)
  fetchTree()
})

// 工具
const formatSize = (bytes) => {
  if (!bytes) return '—'
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + sizes[i]
}
</script>

<template>
  <div class="h-screen flex flex-col bg-slate-50 dark:bg-[#0f1115] text-slate-800 dark:text-slate-200 transition-colors duration-300 font-sans overflow-hidden">
    
    <!-- 1. 顶部导航栏 (CloudDrive2 风格) -->
    <header class="h-14 bg-white dark:bg-[#161b22] border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-4 shrink-0 z-30 shadow-sm">
      <div class="flex items-center space-x-3">
        <!-- 移动端菜单按钮 -->
        <button @click="isMobileMenuOpen = !isMobileMenuOpen" class="lg:hidden p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg">
          <i class="fas" :class="isMobileMenuOpen ? 'fa-times' : 'fa-bars'"></i>
        </button>
        <div class="flex items-center space-x-2 cursor-pointer" @click="navigateTo('')">
          <div class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white shadow-lg shadow-blue-500/20">
            <i class="fas fa-cloud"></i>
          </div>
          <span class="text-lg font-bold tracking-tight hidden sm:block">ModernFM</span>
        </div>
      </div>

      <!-- 面包屑 (适配 CloudDrive2) -->
      <nav class="hidden md:flex items-center flex-1 mx-8 bg-slate-100 dark:bg-[#0d1117] px-4 py-1.5 rounded-full border border-slate-200 dark:border-slate-700 overflow-hidden">
        <template v-for="(bc, idx) in breadcrumbs" :key="bc.path">
          <i v-if="idx > 0" class="fas fa-chevron-right text-[10px] mx-2 opacity-30"></i>
          <button @click="navigateTo(bc.path)" class="hover:text-blue-500 text-sm whitespace-nowrap transition-colors" :class="{'font-bold text-blue-600 dark:text-blue-400': idx === breadcrumbs.length - 1}">
            {{ bc.name }}
          </button>
        </template>
      </nav>

      <!-- 操作区 -->
      <div class="flex items-center space-x-2">
        <div class="relative hidden sm:block">
          <input v-model="searchQuery" type="text" placeholder="搜索..." class="w-40 md:w-64 pl-9 pr-4 py-1.5 bg-slate-100 dark:bg-[#0d1117] border border-slate-200 dark:border-slate-700 rounded-full text-xs focus:ring-2 focus:ring-blue-500/50 outline-none">
          <i class="fas fa-search absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 text-[10px]"></i>
        </div>
        <div class="flex bg-slate-200 dark:bg-slate-800 rounded-full p-1 scale-90">
          <button @click="toggleTheme('light')" class="w-6 h-6 flex items-center justify-center rounded-full transition-all" :class="{'bg-white dark:bg-slate-600 shadow-sm': themeMode === 'light'}"><i class="fas fa-sun text-[10px]"></i></button>
          <button @click="toggleTheme('dark')" class="w-6 h-6 flex items-center justify-center rounded-full transition-all" :class="{'bg-white dark:bg-slate-600 shadow-sm': themeMode === 'dark'}"><i class="fas fa-moon text-[10px]"></i></button>
        </div>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative">
      
      <!-- 2. 左侧目录树 (响应式设计) -->
      <aside :class="{'translate-x-0': isMobileMenuOpen, '-translate-x-full': !isMobileMenuOpen}" 
             class="fixed inset-y-0 left-0 w-72 bg-white dark:bg-[#161b22] border-r border-slate-200 dark:border-slate-800 z-40 transition-transform duration-300 lg:relative lg:translate-x-0 lg:z-0 flex flex-col pt-14 lg:pt-0">
        <div class="p-4 flex items-center justify-between border-b border-slate-100 dark:border-slate-800 lg:hidden">
          <span class="font-bold">导航</span>
          <button @click="isMobileMenuOpen = false"><i class="fas fa-times"></i></button>
        </div>
        <div class="flex-1 overflow-y-auto p-2 custom-scrollbar">
           <!-- 树组件递归 (内联实现) -->
           <div v-for="node in treeData" :key="node.path" class="select-none">
             <div class="flex items-center p-2 rounded-lg cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors" 
                  :class="{'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400': currentPath === node.path}">
               <i @click.stop="toggleFolder(node)" class="fas fa-caret-right mr-2 w-4 text-center transition-transform" :class="{'rotate-90': node.isOpen, 'opacity-0': !node.children.length}"></i>
               <i class="fas mr-2 text-sm" :class="node.path === '' ? 'fa-hdd' : (node.isOpen ? 'fa-folder-open' : 'fa-folder')"></i>
               <span @click="navigateTo(node.path)" class="text-sm font-medium truncate">{{ node.name }}</span>
             </div>
             <!-- 子目录 -->
             <div v-if="node.isOpen" class="ml-4 border-l border-slate-100 dark:border-slate-800">
               <div v-for="child in node.children" :key="child.path">
                  <div class="flex items-center p-2 rounded-lg cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800"
                       :class="{'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400': currentPath === child.path}">
                    <i @click.stop="toggleFolder(child)" class="fas fa-caret-right mr-2 w-4 text-center transition-transform" :class="{'rotate-90': child.isOpen, 'opacity-0': !child.children.length}"></i>
                    <i class="fas fa-folder mr-2 text-sm text-amber-500"></i>
                    <span @click="navigateTo(child.path)" class="text-sm truncate">{{ child.name }}</span>
                  </div>
                  <!-- 简单二级实现，实际可用组件递归 -->
                  <div v-if="child.isOpen" class="ml-4 border-l border-slate-100 dark:border-slate-800">
                    <div v-for="gchild in child.children" :key="gchild.path" @click="navigateTo(gchild.path)"
                         class="flex items-center p-2 rounded-lg cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800 text-xs"
                         :class="{'text-blue-600 font-bold': currentPath === gchild.path}">
                      <i class="fas fa-folder mr-2 opacity-50"></i> {{ gchild.name }}
                    </div>
                  </div>
               </div>
             </div>
           </div>
        </div>
      </aside>

      <!-- 3. 右侧主浏览区 -->
      <main class="flex-1 flex flex-col min-w-0 bg-white dark:bg-[#0d1117] relative">
        <!-- 移动端路径显示 -->
        <div class="md:hidden px-4 py-2 border-b border-slate-200 dark:border-slate-800 flex items-center text-xs text-slate-400 overflow-x-auto whitespace-nowrap">
           <i class="fas fa-home mr-2"></i> {{ currentPath || '/' }}
        </div>

        <!-- 文件工具栏 -->
        <div class="h-12 px-4 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between shrink-0 bg-slate-50/50 dark:bg-[#161b22]/50">
          <div class="flex items-center space-x-4">
             <button @click="navigateTo(currentPath.split('/').slice(0,-1).join('/'))" :disabled="!currentPath" class="p-1.5 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-md disabled:opacity-30">
               <i class="fas fa-arrow-left"></i>
             </button>
             <span class="text-xs font-bold uppercase text-slate-400 tracking-widest">{{ filteredFiles.length }} 项内容</span>
          </div>
          <div class="flex items-center space-x-2">
             <button class="p-2 hover:bg-blue-100 dark:hover:bg-blue-900/30 text-blue-600 rounded-lg transition-colors"><i class="fas fa-upload"></i></button>
             <button class="p-2 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-lg"><i class="fas fa-sync-alt"></i></button>
          </div>
        </div>

        <!-- 文件列表 -->
        <div class="flex-1 overflow-y-auto relative custom-scrollbar">
          <div v-if="isLoading" class="absolute inset-0 flex items-center justify-center bg-white/60 dark:bg-black/40 z-10 backdrop-blur-sm">
            <div class="flex flex-col items-center">
              <i class="fas fa-spinner fa-spin text-3xl text-blue-500 mb-2"></i>
              <span class="text-xs font-medium">加载中...</span>
            </div>
          </div>

          <table class="w-full border-collapse">
            <thead class="sticky top-0 bg-white/90 dark:bg-[#0d1117]/90 backdrop-blur-md border-b border-slate-200 dark:border-slate-800 z-[5] text-left">
              <tr class="text-[10px] font-black text-slate-400 uppercase tracking-tighter">
                <th class="px-4 py-3 w-10 text-center"><input type="checkbox" class="rounded accent-blue-500"></th>
                <th class="px-4 py-3">文件名</th>
                <th class="px-4 py-3 hidden sm:table-cell w-32">大小</th>
                <th class="px-4 py-3 hidden md:table-cell w-48">修改时间</th>
                <th class="px-4 py-3 w-12"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50">
              <tr v-for="file in filteredFiles" :key="file.FullPath" 
                  @dblclick="file.IsDir ? navigateTo(file.FullPath) : null"
                  @click.shift="file.selected = !file.selected"
                  class="group hover:bg-blue-50/50 dark:hover:bg-blue-900/10 cursor-pointer transition-all">
                <td class="px-4 py-2.5 text-center">
                  <input type="checkbox" v-model="file.selected" class="rounded accent-blue-500">
                </td>
                <td class="px-4 py-2.5">
                  <div class="flex items-center space-x-3">
                    <div class="w-9 h-9 rounded-lg flex items-center justify-center text-lg shadow-sm"
                         :class="file.IsDir ? 'bg-amber-100 dark:bg-amber-900/30 text-amber-500' : 'bg-blue-50 dark:bg-blue-900/20 text-blue-500'">
                      <i :class="file.IsDir ? 'fas fa-folder' : 'fas fa-file-alt'"></i>
                    </div>
                    <div class="flex flex-col min-w-0">
                      <span class="text-sm font-semibold truncate group-hover:text-blue-600 transition-colors">{{ file.Name }}</span>
                      <span class="text-[10px] text-slate-400 sm:hidden">{{ formatSize(file.Size) }}</span>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-2.5 hidden sm:table-cell text-xs font-mono text-slate-500">{{ formatSize(file.Size) }}</td>
                <td class="px-4 py-2.5 hidden md:table-cell text-xs text-slate-400">{{ new Date(file.ModTime).toLocaleDateString() }}</td>
                <td class="px-4 py-2.5 text-right opacity-0 group-hover:opacity-100 transition-opacity">
                   <button class="p-1.5 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-md"><i class="fas fa-ellipsis-h text-slate-400"></i></button>
                </td>
              </tr>
            </tbody>
          </table>
          
          <!-- 空状态 -->
          <div v-if="!isLoading && filteredFiles.length === 0" class="flex flex-col items-center justify-center py-32 opacity-20">
            <i class="fas fa-folder-open text-6xl mb-4"></i>
            <p class="font-bold">此目录空无一物</p>
          </div>
        </div>
      </main>

      <!-- 移动端遮罩 -->
      <div v-if="isMobileMenuOpen" @click="isMobileMenuOpen = false" class="fixed inset-0 bg-black/50 z-30 lg:hidden backdrop-blur-sm"></div>
    </div>
  </div>
</template>

<style>
@import url('https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css');

.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 10px; }
.dark .custom-scrollbar::-webkit-scrollbar-thumb { background: #334155; }

/* 移动端横向滚动条隐藏 */
.overflow-x-auto { scrollbar-width: none; }
.overflow-x-auto::-webkit-scrollbar { display: none; }
</style>

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
