<script setup>
import { ref, onMounted, computed, watch, onUnmounted } from 'vue'

/**
 * ModernFM - CloudDrive2 Inspired UI
 * Clean-room rewrite
 */

// --- 状态管理 ---
const theme = ref(localStorage.getItem('fm-theme') || 'light')
const currentPath = ref(localStorage.getItem('fm-last-path') || '')
const files = ref([])
const foldersTree = ref([]) // 用于左侧树的文件夹数据
const isLoading = ref(false)
const searchQuery = ref('')
const isMobileSidebarOpen = ref(false)
const windowWidth = ref(window.innerWidth)

// --- 主题控制 ---
const toggleTheme = () => {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
  localStorage.setItem('fm-theme', theme.value)
}

watch(theme, (newTheme) => {
  document.documentElement.classList.toggle('dark', newTheme === 'dark')
}, { immediate: true })

// --- 响应式监听 ---
const handleResize = () => {
  windowWidth.value = window.innerWidth
  if (windowWidth.value >= 1024) isMobileSidebarOpen.value = false
}

// --- 数据交互 ---
const fetchFiles = async (path = '') => {
  isLoading.value = true
  try {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`)
    if (!res.ok) throw new Error('API Error')
    const data = await res.json()
    
    // 后端返回的是前缀匹配的所有文件，前端需要过滤出当前层级的直接子项
    const prefix = path ? (path.endsWith('/') ? path : path + '/') : ''
    
    files.value = data.filter(item => {
      // 排除路径本身
      if (item.FullPath === path) return false
      
      const rel = item.FullPath.startsWith(prefix) ? item.FullPath.slice(prefix.length) : item.FullPath
      // 如果剩余部分包含斜杠，说明是深层子项，排除
      return rel && !rel.includes('/')
    }).sort((a, b) => (b.IsDir - a.IsDir) || a.Name.localeCompare(b.Name))

    currentPath.value = path
    localStorage.setItem('fm-last-path', path)
  } catch (err) {
    console.error('Fetch failed:', err)
  } finally {
    isLoading.value = false
  }
}

// 全量获取文件夹以构建目录树
const fetchFolderTree = async () => {
  try {
    const res = await fetch('/api/files/list?path=')
    const data = await res.json()
    const allFolders = data.filter(f => f.IsDir).sort((a, b) => a.FullPath.localeCompare(b.FullPath))
    
    // 构建简单层级结构
    const buildTree = (folders, parentPath = '') => {
      return folders
        .filter(f => {
          const parts = f.FullPath.split('/')
          const parent = parts.slice(0, -1).join('/')
          return parent === parentPath
        })
        .map(f => ({
          ...f,
          children: buildTree(folders, f.FullPath),
          isOpen: false
        }))
    }
    
    foldersTree.value = [{ 
      Name: 'Root', 
      FullPath: '', 
      IsDir: true, 
      children: buildTree(allFolders, ''),
      isOpen: true 
    }]
  } catch (err) {
    console.error('Tree fetch failed:', err)
  }
}

// --- 导航逻辑 ---
const navigateTo = (path) => {
  fetchFiles(path)
  if (windowWidth.value < 1024) isMobileSidebarOpen.value = false
}

const breadcrumbs = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  const result = [{ name: 'Root', path: '' }]
  let acc = ''
  parts.forEach(p => {
    acc = acc ? `${acc}/${p}` : p
    result.push({ name: p, path: acc })
  })
  return result
})

const filteredFiles = computed(() => {
  if (!searchQuery.value) return files.value
  const q = searchQuery.value.toLowerCase()
  return files.value.filter(f => f.Name.toLowerCase().includes(q))
})

// --- 工具函数 ---
const formatSize = (bytes) => {
  if (!bytes) return '—'
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + ['B', 'KB', 'MB', 'GB', 'TB'][i]
}

const formatDate = (ds) => new Date(ds).toLocaleDateString('zh-CN')

// --- 生命周期 ---
onMounted(() => {
  fetchFiles(currentPath.value)
  fetchFolderTree()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})

// 递归文件夹组件模拟 (内联)
const TreeItem = {
  props: ['node', 'currentPath', 'navigateTo'],
  template: `
    <div class="tree-node">
      <div @click="node.children.length ? node.isOpen = !node.isOpen : null" 
           class="flex items-center py-1.5 px-2 rounded-md cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors group"
           :class="{'text-blue-600 font-bold bg-blue-50 dark:bg-blue-900/20': currentPath === node.FullPath}">
        <i class="fas fa-caret-right mr-2 w-3 transition-transform" 
           :class="{'rotate-90': node.isOpen, 'opacity-0': !node.children.length}"></i>
        <i class="fas mr-2" :class="node.FullPath === '' ? 'fa-hdd' : (node.isOpen ? 'fa-folder-open' : 'fa-folder')"></i>
        <span @click.stop="navigateTo(node.FullPath)" class="text-sm truncate flex-1">{{ node.Name }}</span>
      </div>
      <div v-if="node.isOpen && node.children.length" class="ml-4 border-l border-gray-200 dark:border-gray-700 pl-1">
        <tree-item v-for="child in node.children" :key="child.FullPath" :node="child" :current-path="currentPath" :navigate-to="navigateTo" />
      </div>
    </div>
  `
}
</script>

<template>
  <div class="fixed inset-0 flex flex-col bg-white dark:bg-[#0d1117] text-gray-900 dark:text-gray-100 transition-colors duration-200 antialiased font-sans">
    
    <!-- Header -->
    <header class="h-14 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between px-4 bg-white/80 dark:bg-[#161b22]/80 backdrop-blur-md z-50 shadow-sm">
      <div class="flex items-center gap-3">
        <button @click="isMobileSidebarOpen = true" class="lg:hidden p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg">
          <i class="fas fa-bars"></i>
        </button>
        <div class="flex items-center gap-2 cursor-pointer" @click="navigateTo('')">
          <div class="w-8 h-8 bg-gradient-to-br from-blue-500 to-blue-700 rounded-lg flex items-center justify-center text-white shadow-md">
            <i class="fas fa-cube"></i>
          </div>
          <span class="font-black text-xl tracking-tight hidden sm:inline-block">ModernFM</span>
        </div>
      </div>

      <nav class="hidden md:flex items-center bg-gray-100 dark:bg-gray-900 px-4 py-1.5 rounded-full border border-gray-200 dark:border-gray-800 max-w-xl flex-1 mx-6 overflow-hidden">
        <template v-for="(bc, i) in breadcrumbs" :key="bc.path">
          <i v-if="i > 0" class="fas fa-chevron-right text-[10px] mx-2 opacity-30 text-gray-400"></i>
          <button @click="navigateTo(bc.path)" class="text-sm hover:text-blue-500 transition-colors whitespace-nowrap truncate" :class="{'font-bold text-blue-600 dark:text-blue-400': i === breadcrumbs.length - 1}">
            {{ bc.name }}
          </button>
        </template>
      </nav>

      <div class="flex items-center gap-4">
        <div class="relative hidden sm:block">
          <input v-model="searchQuery" type="text" placeholder="搜索..." class="w-48 lg:w-64 bg-gray-100 dark:bg-gray-900 border-none rounded-full py-1.5 pl-9 pr-4 text-xs focus:ring-2 focus:ring-blue-500/30">
          <i class="fas fa-search absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 text-[10px]"></i>
        </div>
        <button @click="toggleTheme" class="w-9 h-9 rounded-full flex items-center justify-center bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 transition-colors">
          <i class="fas" :class="theme === 'light' ? 'fa-moon' : 'fa-sun'"></i>
        </button>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative">
      
      <!-- Left Sidebar (Desktop) -->
      <aside :class="{'translate-x-0': isMobileSidebarOpen, '-translate-x-full lg:translate-x-0': !isMobileSidebarOpen}" 
             class="fixed lg:relative inset-y-0 left-0 w-72 bg-white dark:bg-[#0d1117] border-r border-gray-200 dark:border-gray-800 z-[60] lg:z-0 transition-transform duration-300 flex flex-col shadow-xl lg:shadow-none pt-14 lg:pt-0">
        <div class="p-4 flex items-center justify-between border-b border-gray-100 dark:border-gray-800 lg:hidden">
          <span class="font-bold">文件夹导航</span>
          <button @click="isMobileSidebarOpen = false" class="p-2"><i class="fas fa-times"></i></button>
        </div>
        <div class="flex-1 overflow-y-auto p-3 custom-scrollbar scroll-smooth">
          <component :is="TreeItem" v-for="root in foldersTree" :key="root.FullPath" :node="root" :current-path="currentPath" :navigate-to="navigateTo" />
        </div>
      </aside>

      <!-- Main Browser Area -->
      <main class="flex-1 flex flex-col bg-white dark:bg-[#0d1117] overflow-hidden min-w-0">
        
        <!-- Toolbar -->
        <div class="h-12 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between px-4 shrink-0 bg-gray-50/30 dark:bg-gray-800/10">
          <div class="flex items-center gap-3">
            <button @click="navigateTo(currentPath.split('/').slice(0,-1).join('/'))" :disabled="!currentPath" class="p-2 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg disabled:opacity-30">
              <i class="fas fa-arrow-left text-sm"></i>
            </button>
            <span class="text-[10px] font-black uppercase tracking-widest text-gray-400">{{ filteredFiles.length }} 项内容</span>
          </div>
          <div class="flex items-center gap-2">
            <button class="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-bold transition-all shadow-lg shadow-blue-500/20">
              <i class="fas fa-plus mr-1.5"></i>新建
            </button>
            <button class="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg text-gray-500"><i class="fas fa-sync-alt"></i></button>
          </div>
        </div>

        <!-- File List -->
        <div class="flex-1 overflow-y-auto relative custom-scrollbar">
          <div v-if="isLoading" class="absolute inset-0 bg-white/50 dark:bg-black/50 z-20 flex items-center justify-center backdrop-blur-sm">
            <div class="flex flex-col items-center">
              <div class="w-10 h-10 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
              <span class="mt-3 text-xs font-bold text-blue-500">正在获取文件...</span>
            </div>
          </div>

          <table class="w-full text-left border-collapse">
            <thead class="sticky top-0 bg-white/95 dark:bg-[#0d1117]/95 backdrop-blur-md border-b border-gray-200 dark:border-gray-800 z-10">
              <tr class="text-[10px] font-black text-gray-400 uppercase tracking-wider">
                <th class="px-4 py-3 w-10 text-center"><input type="checkbox" class="rounded border-gray-300 dark:border-gray-600"></th>
                <th class="px-4 py-3">名称</th>
                <th class="px-4 py-3 hidden sm:table-cell w-32 text-right">大小</th>
                <th class="px-4 py-3 hidden md:table-cell w-44 text-right">修改时间</th>
                <th class="px-4 py-3 w-12"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800/50">
              <tr v-for="file in filteredFiles" :key="file.FullPath" 
                  @dblclick="file.IsDir ? navigateTo(file.FullPath) : null"
                  class="group hover:bg-blue-50/40 dark:hover:bg-blue-900/10 transition-all cursor-pointer">
                <td class="px-4 py-3 text-center"><input type="checkbox" class="rounded accent-blue-600"></td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-3">
                    <div class="w-10 h-10 flex items-center justify-center rounded-xl transition-transform group-hover:scale-105"
                         :class="file.IsDir ? 'bg-amber-100 dark:bg-amber-900/30 text-amber-500' : 'bg-blue-50 dark:bg-blue-900/20 text-blue-500'">
                      <i :class="file.IsDir ? 'fas fa-folder text-xl' : 'fas fa-file-alt text-xl'"></i>
                    </div>
                    <div class="flex flex-col min-w-0">
                      <span class="text-sm font-semibold truncate max-w-xs md:max-w-md group-hover:text-blue-600">{{ file.Name }}</span>
                      <span class="text-[10px] text-gray-400 sm:hidden">{{ formatSize(file.Size) }} • {{ formatDate(file.ModTime) }}</span>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-3 hidden sm:table-cell text-right text-xs font-mono text-gray-500">{{ formatSize(file.Size) }}</td>
                <td class="px-4 py-3 hidden md:table-cell text-right text-xs text-gray-400">{{ formatDate(file.ModTime) }}</td>
                <td class="px-4 py-3 text-right">
                  <button class="p-2 opacity-0 group-hover:opacity-100 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-all">
                    <i class="fas fa-ellipsis-v text-gray-400"></i>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>

          <!-- Empty State -->
          <div v-if="!isLoading && filteredFiles.length === 0" class="flex flex-col items-center justify-center py-40 opacity-20">
            <i class="fas fa-box-open text-7xl mb-4"></i>
            <p class="font-black text-lg">此处空空如也</p>
          </div>
        </div>
      </main>

      <!-- Mobile Overlay -->
      <div v-if="isMobileSidebarOpen" @click="isMobileSidebarOpen = false" 
           class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[55] lg:hidden"></div>
    </div>
  </div>
</template>

<style>
@import url('https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css');

.custom-scrollbar::-webkit-scrollbar { width: 5px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #d1d5db; border-radius: 10px; }
.dark .custom-scrollbar::-webkit-scrollbar-thumb { background: #374151; }

input:focus { outline: none; }
</style>
