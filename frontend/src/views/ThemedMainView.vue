<script setup>
import { ref, onMounted, computed, watch } from 'vue'

/**
 * ModernFM - Ultra Logic Edition
 * Strategy: Incremental Scanning + Settings Menu
 */

const THEME_KEY = 'fm-theme'
const PATH_KEY = 'fm-last-path'

const theme = ref(localStorage.getItem(THEME_KEY) || 'auto')
const currentPath = ref(localStorage.getItem(PATH_KEY) || '')
const files = ref([])
const foldersTree = ref([])
const isLoading = ref(false)
const isSettingsOpen = ref(false)
const isMobileSidebarOpen = ref(false)

// --- Theme ---
const updateTheme = () => {
    const root = document.documentElement
    if (theme.value === 'auto') {
        root.classList.toggle('dark', window.matchMedia('(prefers-color-scheme: dark)').matches)
    } else {
        root.classList.toggle('dark', theme.value === 'dark')
    }
    localStorage.setItem(THEME_KEY, theme.value)
}
watch(theme, updateTheme)

// --- File Logic (Incremental) ---
const fetchFiles = async (path = '') => {
  isLoading.value = true
  try {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`)
    const data = await res.json()
    files.value = Array.isArray(data) ? data.sort((a, b) => (b.IsDir - a.IsDir) || a.Name.localeCompare(b.Name)) : []
    currentPath.value = path
    localStorage.setItem(PATH_KEY, path)
  } catch (e) { console.error(e) } finally { isLoading.value = false }
}

const fetchSubFolders = async (parentPath) => {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(parentPath)}`)
    const data = await res.json()
    return data.filter(f => f.IsDir).sort((a,b) => a.Name.localeCompare(b.Name))
}

const initializeTree = async () => {
    const rootFolders = await fetchSubFolders('')
    foldersTree.value = [{ Name: '资源库', FullPath: '', children: rootFolders.map(f => ({ ...f, children: [], isOpen: false, loaded: false })), isOpen: true, loaded: true }]
}

const toggleFolder = async (node) => {
    node.isOpen = !node.isOpen
    if (node.isOpen && !node.loaded) {
        node.loading = true
        const subs = await fetchSubFolders(node.FullPath)
        node.children = subs.map(f => ({ ...f, children: [], isOpen: false, loaded: false }))
        node.loaded = true
        node.loading = false
    }
}

const startFullScan = async () => {
    if (confirm('全量扫描占用大量资源，系统有可能变慢，是否启动？')) {
        await fetch('/api/system/rescan', { method: 'POST' })
        alert('后台全量扫描已启动。')
        isSettingsOpen.value = false
    }
}

const navigateTo = (p) => { fetchFiles(p); isMobileSidebarOpen.value = false }

onMounted(() => {
    updateTheme()
    fetchFiles(currentPath.value)
    initializeTree()
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', updateTheme)
})
</script>

<template>
  <div class="fixed inset-0 flex flex-col bg-slate-50 dark:bg-[#0d1117] text-slate-900 dark:text-slate-100 transition-colors duration-300 font-sans antialiased overflow-hidden select-none">
    
    <!-- Navbar -->
    <header class="h-16 border-b border-slate-200 dark:border-slate-800 bg-white/95 dark:bg-[#161b22]/95 backdrop-blur-md flex items-center justify-between px-6 z-[100] shrink-0 shadow-sm">
      <div class="flex items-center gap-4">
        <button @click="isMobileSidebarOpen = true" class="lg:hidden p-2 text-slate-500"><i class="fas fa-bars text-xl"></i></button>
        <div @click="navigateTo('')" class="flex items-center gap-2 cursor-pointer group">
          <div class="w-10 h-10 bg-blue-600 rounded-2xl flex items-center justify-center text-white shadow-lg transition-transform group-hover:scale-105">
            <i class="fas fa-microchip text-xl"></i>
          </div>
          <span class="text-2xl font-black tracking-tighter hidden sm:block">ModernFM</span>
        </div>
      </div>

      <nav class="hidden md:flex items-center bg-slate-100 dark:bg-slate-900 px-6 py-2.5 rounded-full border border-slate-200 dark:border-slate-800 flex-1 mx-12 overflow-hidden shadow-inner">
        <div v-for="(part, i) in currentPath.split('/').filter(Boolean)" :key="i" class="flex items-center">
            <i class="fas fa-chevron-right text-[9px] mx-3 opacity-30 text-gray-400"></i>
            <span class="text-sm font-bold text-blue-600">{{ part }}</span>
        </div>
        <span v-if="!currentPath" class="text-sm font-bold opacity-30">根目录</span>
      </nav>

      <div class="flex items-center gap-3 relative">
        <button @click="isSettingsOpen = !isSettingsOpen" class="w-11 h-11 rounded-full flex items-center justify-center bg-slate-100 dark:bg-slate-800 hover:bg-white dark:hover:bg-slate-700 transition-all border border-slate-200 dark:border-slate-700 shadow-sm">
          <i class="fas fa-cog text-gray-500 group-hover:rotate-90 transition-transform"></i>
        </button>

        <!-- Settings Dropdown -->
        <div v-if="isSettingsOpen" class="absolute top-14 right-0 w-64 bg-white dark:bg-[#1c2128] border border-slate-200 dark:border-slate-700 rounded-3xl shadow-2xl p-2 z-[110] animate-fade-in">
           <div class="px-4 py-3 border-b border-slate-100 dark:border-slate-800 mb-2">
               <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">系统偏好设置</span>
           </div>
           
           <div class="space-y-1">
               <div class="px-3 py-2 text-xs font-bold text-slate-500">外观模式</div>
               <button @click="theme = 'light'" :class="{'bg-blue-50 dark:bg-blue-900/30 text-blue-600': theme === 'light'}" class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 transition-all text-sm">
                   <i class="fas fa-sun w-5"></i> 明亮模式
               </button>
               <button @click="theme = 'dark'" :class="{'bg-blue-50 dark:bg-blue-900/30 text-blue-600': theme === 'dark'}" class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 transition-all text-sm">
                   <i class="fas fa-moon w-5"></i> 暗黑模式
               </button>
               <button @click="theme = 'auto'" :class="{'bg-blue-50 dark:bg-blue-900/30 text-blue-600': theme === 'auto'}" class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 transition-all text-sm">
                   <i class="fas fa-desktop w-5"></i> 跟随系统
               </button>
           </div>

           <div class="h-px bg-slate-100 dark:bg-slate-800 my-2 mx-2"></div>
           
           <div class="space-y-1">
               <div class="px-3 py-2 text-xs font-bold text-slate-500">数据维护</div>
               <button @click="startFullScan" class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-red-50 dark:hover:bg-red-900/20 text-red-500 transition-all text-sm font-bold">
                   <i class="fas fa-sync-alt w-5"></i> 启动全量扫描
               </button>
           </div>
        </div>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative">
      <!-- Left Sidebar -->
      <aside :class="{'translate-x-0': isMobileSidebarOpen, '-translate-x-full lg:translate-x-0': !isMobileSidebarOpen}" 
             class="fixed lg:relative inset-y-0 left-0 w-80 bg-white dark:bg-[#161b22] border-r border-slate-200 dark:border-slate-800 z-[60] lg:z-0 transition-transform duration-500 flex flex-col pt-16 lg:pt-0">
        <div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
           <div v-for="node in foldersTree" :key="node.FullPath">
              <RecursiveTree :node="node" :currentPath="currentPath" :navigateTo="navigateTo" :toggleFolder="toggleFolder" />
           </div>
        </div>
      </aside>

      <!-- Main Browser -->
      <main class="flex-1 flex flex-col bg-white dark:bg-[#0d1117] min-w-0 overflow-hidden relative shadow-inner">
        <div class="h-14 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between px-6 bg-slate-50/50 dark:bg-slate-800/10 shrink-0">
          <div class="flex items-center gap-4">
            <button @click="navigateTo(currentPath.split('/').slice(0,-1).join('/'))" :disabled="!currentPath" class="w-9 h-9 flex items-center justify-center bg-white dark:bg-slate-800 rounded-full border border-slate-200 dark:border-slate-700 shadow-sm disabled:opacity-30"><i class="fas fa-chevron-left text-sm"></i></button>
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">{{ files.length }} 项内容</span>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto custom-scrollbar p-6">
          <div v-if="isLoading" class="absolute inset-0 bg-white/60 dark:bg-black/60 z-[90] flex items-center justify-center backdrop-blur-sm"><i class="fas fa-spinner fa-spin text-4xl text-blue-600"></i></div>
          
          <div class="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-6 2xl:grid-cols-8 gap-5">
             <div v-for="file in files" :key="file.FullPath" @dblclick="file.IsDir ? navigateTo(file.FullPath) : null" 
                  class="group relative bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-800 p-4 rounded-3xl cursor-pointer hover:bg-white dark:hover:bg-slate-800 hover:shadow-2xl transition-all duration-300">
                <div class="aspect-square flex items-center justify-center mb-3">
                    <i class="fas text-4xl transition-transform duration-500 group-hover:scale-110" :class="file.IsDir ? 'fa-folder text-amber-500' : 'fa-file-alt text-blue-400'"></i>
                </div>
                <p class="text-[11px] font-black text-center truncate px-2 text-slate-600 dark:text-slate-300 group-hover:text-blue-600 transition-colors">{{ file.Name }}</p>
             </div>
          </div>

          <div v-if="!isLoading && !files.length" class="flex flex-col items-center py-48 opacity-10 grayscale"><i class="fas fa-box-open text-8xl mb-6"></i><p class="text-2xl font-black uppercase tracking-tighter">此目录空无一物</p></div>
        </div>
      </main>

      <div v-if="isMobileSidebarOpen || isSettingsOpen" @click="isMobileSidebarOpen = false; isSettingsOpen = false" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[55] animate-fade-in"></div>
    </div>
  </div>
</template>

<script>
const RecursiveTree = {
    name: 'RecursiveTree',
    props: ['node', 'currentPath', 'navigateTo', 'toggleFolder'],
    template: `
        <div class="tree-node mb-1">
            <div @click="toggleFolder(node)" class="flex items-center gap-2.5 p-2 rounded-xl cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800 transition-all" :class="{'bg-blue-50 dark:bg-blue-900/20 text-blue-600 font-bold shadow-sm': currentPath === node.FullPath}">
                <div class="w-4 h-4 flex items-center justify-center shrink-0">
                    <i v-if="node.loading" class="fas fa-circle-notch fa-spin text-[10px]"></i>
                    <i v-else class="fas fa-caret-right transition-transform duration-300" :class="{'rotate-90': node.isOpen, 'opacity-0': node.loaded && !node.children.length}"></i>
                </div>
                <i class="fas text-sm shrink-0" :class="node.FullPath === '' ? 'fa-home' : (node.isOpen ? 'fa-folder-open text-amber-500' : 'fa-folder text-amber-500')"></i>
                <span @click.stop="navigateTo(node.FullPath)" class="truncate text-sm flex-1">{{ node.Name }}</span>
            </div>
            <div v-if="node.isOpen" class="ml-5 border-l-2 border-slate-50 dark:border-slate-800 pl-2 mt-1 space-y-1">
                <RecursiveTree v-for="child in node.children" :key="child.FullPath" :node="child" :currentPath="currentPath" :navigateTo="navigateTo" :toggleFolder="toggleFolder" />
            </div>
        </div>
    `
}
</script>

<style>
@import url('https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css');
.custom-scrollbar::-webkit-scrollbar { width: 5px; height: 5px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 20px; }
.dark .custom-scrollbar::-webkit-scrollbar-thumb { background: #334155; }
@keyframes fadeIn { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }
.animate-fade-in { animation: fadeIn 0.2s ease-out; }
</style>
