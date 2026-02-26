<script setup>
import { ref, onMounted, computed, watch } from 'vue'

/**
 * ModernFM - Enterprise Grade UI
 * Optimized for Stability and Speed
 */

// --- Constants ---
const THEME_KEY = 'fm-theme'
const PATH_KEY = 'fm-last-path'

// --- State ---
const theme = ref(localStorage.getItem(THEME_KEY) || 'light')
const currentPath = ref(localStorage.getItem(PATH_KEY) || '')
const files = ref([])
const foldersTree = ref([])
const isLoading = ref(false)
const searchQuery = ref('')
const isMobileSidebarOpen = ref(false)

// --- Theme Logic ---
const applyTheme = (val) => {
  document.documentElement.classList.toggle('dark', val === 'dark')
}
const toggleTheme = () => {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
  localStorage.setItem(THEME_KEY, theme.value)
  applyTheme(theme.value)
}

// --- Data Fetching ---
const fetchFiles = async (path = '') => {
  isLoading.value = true
  try {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`)
    if (!res.ok) throw new Error('API_ERROR')
    const data = await res.json()
    
    // Safety check: ensure data is array
    if (!Array.isArray(data)) {
        files.value = []
        return
    }

    const prefix = path ? (path.endsWith('/') ? path : path + '/') : ''
    
    files.value = data.filter(f => {
      if (f.FullPath === path) return false
      const rel = f.FullPath.startsWith(prefix) ? f.FullPath.slice(prefix.length) : f.FullPath
      return rel && !rel.includes('/')
    }).sort((a, b) => (b.IsDir - a.IsDir) || a.Name.localeCompare(b.Name))

    currentPath.value = path
    localStorage.setItem(PATH_KEY, path)
  } catch (err) {
    console.error('[ModernFM] Fetch failed:', err)
  } finally {
    isLoading.value = false
  }
}

const fetchFolderTree = async () => {
  try {
    const res = await fetch('/api/files/list?path=')
    const data = await res.json()
    if (!Array.isArray(data)) return

    const folders = data.filter(f => f.IsDir).sort((a, b) => a.FullPath.localeCompare(b.FullPath))
    
    const build = (list, p = '') => list.filter(f => {
      const parts = f.FullPath.split('/')
      return parts.slice(0, -1).join('/') === p
    }).map(f => ({ ...f, children: build(list, f.FullPath), isOpen: false }))

    foldersTree.value = [{ 
        Name: '我的文件', 
        FullPath: '', 
        children: build(folders, ''), 
        isOpen: true 
    }]
  } catch (err) {
    console.error('[ModernFM] Tree failed:', err)
  }
}

// --- Navigation ---
const navigateTo = (p) => {
  fetchFiles(p)
  isMobileSidebarOpen.value = false
}

const breadcrumbs = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  let acc = ''
  return [
    { name: '首页', path: '' },
    ...parts.map(p => ({ 
        name: p, 
        path: acc = acc ? `${acc}/${p}` : p 
    }))
  ]
})

const filteredFiles = computed(() => {
  if (!searchQuery.value) return files.value
  const q = searchQuery.value.toLowerCase()
  return files.value.filter(f => f.Name.toLowerCase().includes(q))
})

// --- Utils ---
const formatSize = (b) => {
  if (!b) return '—'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(1024))
  return (b / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}

const formatDate = (d) => new Date(d).toLocaleDateString('zh-CN')

onMounted(() => {
  applyTheme(theme.value)
  fetchFiles(currentPath.value)
  fetchFolderTree()
})
</script>

<template>
  <div class="fixed inset-0 flex flex-col bg-gray-50 dark:bg-[#0d1117] text-gray-900 dark:text-gray-100 transition-colors duration-300 font-sans antialiased overflow-hidden">
    
    <!-- Top Navigation -->
    <header class="h-16 border-b border-gray-200 dark:border-gray-800 bg-white/90 dark:bg-[#161b22]/90 backdrop-blur-md flex items-center justify-between px-6 z-50 shrink-0">
      <div class="flex items-center gap-4">
        <button @click="isMobileSidebarOpen = true" class="lg:hidden p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg">
          <i class="fas fa-bars text-xl"></i>
        </button>
        <div class="flex items-center gap-2 cursor-pointer group" @click="navigateTo('')">
          <div class="w-10 h-10 bg-blue-600 rounded-2xl flex items-center justify-center text-white shadow-xl shadow-blue-500/20 group-hover:scale-110 transition-transform">
            <i class="fas fa-cube text-xl"></i>
          </div>
          <span class="text-2xl font-black tracking-tighter hidden sm:block">ModernFM</span>
        </div>
      </div>

      <!-- Desktop Breadcrumbs -->
      <nav class="hidden md:flex items-center bg-gray-100 dark:bg-gray-900 px-6 py-2.5 rounded-full border border-gray-200 dark:border-gray-800 flex-1 mx-12 overflow-hidden shadow-inner">
        <template v-for="(bc, i) in breadcrumbs" :key="bc.path">
          <i v-if="i > 0" class="fas fa-chevron-right text-[10px] mx-4 opacity-30 text-gray-400"></i>
          <button @click="navigateTo(bc.path)" class="text-sm hover:text-blue-500 font-bold transition-colors truncate" :class="{'text-blue-600': i === breadcrumbs.length - 1}">
            {{ bc.name }}
          </button>
        </template>
      </nav>

      <!-- Actions -->
      <div class="flex items-center gap-5">
        <div class="relative hidden lg:block">
          <input v-model="searchQuery" type="text" placeholder="搜索内容..." class="w-72 bg-gray-100 dark:bg-gray-900 border-none rounded-full py-2.5 pl-12 pr-4 text-sm focus:ring-4 focus:ring-blue-500/10 transition-all">
          <i class="fas fa-search absolute left-5 top-1/2 -translate-y-1/2 text-gray-400"></i>
        </div>
        <button @click="toggleTheme" class="w-11 h-11 rounded-full flex items-center justify-center bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 transition-all border border-gray-200 dark:border-gray-700 shadow-sm">
          <i class="fas" :class="theme === 'light' ? 'fa-moon' : 'fa-sun'"></i>
        </button>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative">
      <!-- Left Sidebar (Tree) -->
      <aside :class="{'translate-x-0': isMobileSidebarOpen, '-translate-x-full lg:translate-x-0': !isMobileSidebarOpen}" 
             class="fixed lg:relative inset-y-0 left-0 w-80 bg-white dark:bg-[#161b22] border-r border-gray-200 dark:border-gray-800 z-[60] lg:z-0 transition-transform duration-500 ease-out flex flex-col shadow-2xl lg:shadow-none pt-16 lg:pt-0">
        <div class="p-6 border-b border-gray-100 dark:border-gray-800 lg:hidden flex justify-between items-center bg-gray-50 dark:bg-gray-900/50">
          <span class="font-black text-xl tracking-tighter">目录导航</span>
          <button @click="isMobileSidebarOpen = false" class="p-2"><i class="fas fa-times text-xl text-gray-400"></i></button>
        </div>
        <div class="flex-1 overflow-y-auto p-4 custom-scrollbar bg-white dark:bg-[#161b22]">
          <div v-for="node in foldersTree" :key="node.FullPath" class="mb-4">
            <div @click="node.isOpen = !node.isOpen" 
                 class="flex items-center gap-3 p-3 rounded-2xl cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 transition-all" 
                 :class="{'bg-blue-50 dark:bg-blue-900/20 text-blue-600 shadow-sm': currentPath === node.FullPath}">
              <i class="fas fa-caret-right transition-transform duration-300" :class="{'rotate-90': node.isOpen, 'opacity-0': !node.children.length}"></i>
              <i class="fas text-lg" :class="node.FullPath === '' ? 'fa-home' : (node.isOpen ? 'fa-folder-open text-amber-500' : 'fa-folder text-amber-500')"></i>
              <span @click.stop="navigateTo(node.FullPath)" class="truncate font-bold text-sm">{{ node.Name }}</span>
            </div>
            <div v-if="node.isOpen" class="ml-6 border-l-2 border-gray-100 dark:border-gray-800 pl-3 mt-1 space-y-1">
              <div v-for="child in node.children" :key="child.FullPath" 
                   @click="child.isOpen = !child.isOpen" 
                   class="flex items-center gap-2.5 p-2.5 rounded-xl cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 transition-all text-sm group" 
                   :class="{'text-blue-600 font-black bg-blue-50 dark:bg-blue-900/20': currentPath === child.FullPath}">
                <i class="fas fa-caret-right text-[10px] transition-transform" :class="{'rotate-90': child.isOpen, 'opacity-0': !child.children.length}"></i>
                <i class="fas fa-folder text-amber-500 opacity-60 group-hover:opacity-100 transition-opacity"></i>
                <span @click.stop="navigateTo(child.FullPath)" class="truncate">{{ child.Name }}</span>
              </div>
            </div>
          </div>
        </div>
      </aside>

      <!-- Main Panel (File Grid/List) -->
      <main class="flex-1 flex flex-col bg-white dark:bg-[#0d1117] min-w-0 overflow-hidden relative shadow-inner">
        <!-- Control Bar -->
        <div class="h-16 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between px-8 bg-white dark:bg-[#0d1117] shrink-0">
          <div class="flex items-center gap-6">
            <button @click="navigateTo(currentPath.split('/').slice(0,-1).join('/'))" :disabled="!currentPath" 
                    class="w-10 h-10 flex items-center justify-center bg-gray-100 dark:bg-gray-800 hover:bg-blue-600 hover:text-white dark:hover:bg-blue-600 rounded-full disabled:opacity-20 transition-all shadow-sm">
              <i class="fas fa-chevron-left"></i>
            </button>
            <div class="flex flex-col">
                <span class="text-xs font-black text-gray-400 uppercase tracking-widest leading-none mb-1">文件管理器</span>
                <span class="text-lg font-black tracking-tight">{{ breadcrumbs[breadcrumbs.length - 1].name }}</span>
            </div>
          </div>
          <div class="flex items-center gap-3">
            <button class="px-6 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-2xl text-xs font-black shadow-xl shadow-blue-600/30 active:scale-95 transition-all">上传新文件</button>
            <button class="w-10 h-10 flex items-center justify-center text-gray-400 hover:text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-xl transition-all"><i class="fas fa-redo-alt"></i></button>
          </div>
        </div>

        <!-- File Content -->
        <div class="flex-1 overflow-y-auto custom-scrollbar p-6">
          <div v-if="isLoading" class="absolute inset-0 bg-white/80 dark:bg-black/80 z-[100] flex items-center justify-center backdrop-blur-md transition-all">
             <div class="flex flex-col items-center">
                 <div class="w-12 h-12 border-[5px] border-blue-100 border-t-blue-600 rounded-full animate-spin mb-4"></div>
                 <span class="text-sm font-black text-blue-600 tracking-tighter uppercase">读取数据中...</span>
             </div>
          </div>
          
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
             <div v-for="file in filteredFiles" :key="file.FullPath" 
                  @dblclick="file.IsDir ? navigateTo(file.FullPath) : null" 
                  class="group relative bg-gray-50 dark:bg-[#161b22] border border-gray-100 dark:border-gray-800 p-4 rounded-3xl cursor-pointer hover:bg-white dark:hover:bg-[#1c2128] hover:shadow-2xl hover:shadow-blue-500/10 hover:-translate-y-1 transition-all duration-300">
                
                <div class="flex items-start justify-between mb-4">
                    <div class="w-14 h-14 flex items-center justify-center rounded-2xl transition-all duration-500 shadow-sm group-hover:rotate-6" 
                         :class="file.IsDir ? 'bg-amber-100 dark:bg-amber-900/30 text-amber-500' : 'bg-blue-100 dark:bg-blue-900/30 text-blue-500'">
                      <i class="fas text-2xl" :class="file.IsDir ? 'fa-folder' : 'fa-file-alt'"></i>
                    </div>
                    <button class="p-2 opacity-0 group-hover:opacity-100 text-gray-300 hover:text-blue-600 transition-all"><i class="fas fa-ellipsis-h text-xl"></i></button>
                </div>
                
                <div class="min-w-0">
                    <p class="font-black text-sm truncate group-hover:text-blue-600 transition-colors">{{ file.Name }}</p>
                    <p class="text-[11px] font-bold text-gray-400 mt-1 uppercase tracking-tighter">
                        {{ file.IsDir ? '文件夹' : formatSize(file.Size) }} • {{ formatDate(file.ModTime) }}
                    </p>
                </div>
                
                <!-- Hover Glow -->
                <div class="absolute inset-0 rounded-3xl border-2 border-transparent group-hover:border-blue-500/30 transition-all pointer-events-none"></div>
             </div>
          </div>

          <!-- Empty Screen -->
          <div v-if="!isLoading && !filteredFiles.length" class="flex flex-col items-center justify-center py-48 opacity-10 grayscale">
            <i class="fas fa-box-open text-[120px] mb-8"></i>
            <p class="text-4xl font-black tracking-tighter">这里什么都没有</p>
          </div>
        </div>
      </main>

      <!-- Overlay -->
      <div v-if="isMobileSidebarOpen" @click="isMobileSidebarOpen = false" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[55] lg:hidden animate-fade-in"></div>
    </div>
  </div>
</template>

<style>
@import url('https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css');

.custom-scrollbar::-webkit-scrollbar { width: 6px; height: 6px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 20px; border: 1px solid white; }
.dark .custom-scrollbar::-webkit-scrollbar-thumb { background: #21262d; border-color: #0d1117; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }

/* Grid Layout Refinement */
@media (min-width: 1024px) {
  .grid { grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); }
}

/* Animations */
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
.animate-fade-in { animation: fadeIn 0.3s ease-out; }

/* Global overrides for modern feel */
::selection { background: #3b82f6; color: white; }
</style>
