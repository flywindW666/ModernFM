<script setup>
import { ref, onMounted, computed, watch, onUnmounted } from 'vue'

/**
 * ModernFM - Premium Enterprise Edition
 * Features: Auto-Scan detection, Adaptive Theme, Navigation History
 */

const THEME_KEY = 'fm-theme'
const PATH_KEY = 'fm-last-path'
const VIEW_MODE_KEY = 'fm-view-mode'

const theme = ref(localStorage.getItem(THEME_KEY) || 'auto')
const currentPath = ref(localStorage.getItem(PATH_KEY) || '')
const history = ref([currentPath.value])
const historyIndex = ref(0)

const files = ref([])
const foldersTree = ref([])
const isLoading = ref(false)
const isScanning = ref(false) // 扫描状态
const searchQuery = ref('')
const isMobileSidebarOpen = ref(false)
const viewMode = ref(localStorage.getItem(VIEW_MODE_KEY) || 'grid')
const sortKey = ref('Name')
const sortOrder = ref(1)

const selectedPaths = ref(new Set())
const contextMenu = ref({ show: false, x: 0, y: 0, target: null })

// --- 主题逻辑 ---
const updateSystemTheme = () => {
    if (theme.value === 'auto') {
        const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches
        document.documentElement.classList.toggle('dark', isDark)
    }
}

watch(theme, (v) => {
    localStorage.setItem(THEME_KEY, v)
    if (v === 'auto') updateSystemTheme()
    else document.documentElement.classList.toggle('dark', v === 'dark')
}, { immediate: true })

// --- 状态轮询 (检查扫描进度) ---
const checkScanStatus = async () => {
    try {
        const res = await fetch('/api/files/list?path=')
        const data = await res.json()
        // 如果数据还在快速增长或后端标记忙碌
        if (data.length === 0) isScanning.value = true
        else isScanning.value = false
    } catch (e) { isScanning.value = false }
}

// --- 数据获取 ---
const fetchFiles = async (path = '', skipHistory = false) => {
  isLoading.value = true
  try {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`)
    const data = await res.json()
    
    // 核心逻辑：如果正在扫描且数据不全，先显示文件夹，过滤掉文件以保证流畅
    if (isScanning.value) {
        files.value = data.filter(f => f.IsDir)
    } else {
        files.value = Array.isArray(data) ? data : []
    }

    currentPath.value = path
    localStorage.setItem(PATH_KEY, path)
    selectedPaths.value.clear()

    if (!skipHistory) {
        history.value = history.value.slice(0, historyIndex.value + 1)
        history.value.push(path)
        historyIndex.value++
    }
  } catch (e) { console.error(e) } finally { isLoading.value = false }
}

// --- 导航历史 ---
const goBack = () => {
    if (historyIndex.value > 0) {
        historyIndex.value--
        fetchFiles(history.value[historyIndex.value], true)
    }
}
const goForward = () => {
    if (historyIndex.value < history.value.length - 1) {
        historyIndex.value++
        fetchFiles(history.value[historyIndex.value], true)
    }
}

const navigateTo = (p) => fetchFiles(p)

// --- 目录树 ---
const fetchSubFolders = async (parentPath) => {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(parentPath)}`)
    const data = await res.json()
    return data.filter(f => f.IsDir).sort((a,b) => a.Name.localeCompare(b.Name))
}

const initializeTree = async () => {
    const rootFolders = await fetchSubFolders('')
    foldersTree.value = [{ Name: '根目录', FullPath: '', children: rootFolders.map(f => ({ ...f, children: [], isOpen: false, loaded: false })), isOpen: true, loaded: true }]
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

onMounted(() => {
    checkScanStatus()
    fetchFiles(currentPath.value)
    initializeTree()
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', updateSystemTheme)
})
</script>

<template>
  <div class="fixed inset-0 flex flex-col bg-slate-50 dark:bg-[#0d1117] text-slate-900 dark:text-slate-100 transition-colors duration-300 font-sans antialiased overflow-hidden select-none">
    
    <!-- Header -->
    <header class="h-16 border-b border-slate-200 dark:border-slate-800 bg-white/95 dark:bg-[#161b22]/95 backdrop-blur-md flex items-center justify-between px-6 z-50 shrink-0 shadow-sm">
      <div class="flex items-center gap-4">
        <button @click="isMobileSidebarOpen = true" class="lg:hidden p-2 text-slate-500"><i class="fas fa-bars text-xl"></i></button>
        <div class="flex items-center gap-2 cursor-pointer group" @click="navigateTo('')">
          <div class="w-10 h-10 bg-gradient-to-tr from-blue-600 to-indigo-600 rounded-xl flex items-center justify-center text-white shadow-lg group-hover:rotate-6 transition-transform">
            <i class="fas fa-database text-xl"></i>
          </div>
          <span class="text-2xl font-black tracking-tighter hidden sm:block">ModernFM</span>
        </div>
      </div>

      <!-- Center: Breadcrumbs -->
      <nav class="hidden md:flex items-center bg-slate-100/80 dark:bg-slate-900/80 px-6 py-2 rounded-full border border-slate-200 dark:border-slate-700 flex-1 mx-12 overflow-hidden shadow-inner">
        <div v-for="(bc, i) in breadcrumbs" :key="bc.path" class="flex items-center shrink-0">
          <i v-if="i > 0" class="fas fa-chevron-right text-[9px] mx-3 opacity-30 text-slate-400"></i>
          <button @click="navigateTo(bc.path)" class="text-sm hover:text-blue-500 font-bold transition-colors truncate max-w-[120px]" :class="{'text-blue-600': i === breadcrumbs.length - 1}">{{ bc.name }}</button>
        </div>
      </nav>

      <!-- Right: System Theme & Search -->
      <div class="flex items-center gap-3">
        <div class="flex bg-slate-100 dark:bg-slate-800 p-1 rounded-full shadow-inner border border-slate-200 dark:border-slate-700">
          <button @click="theme = 'light'" :class="{'bg-white dark:bg-slate-600 shadow-sm text-blue-600': theme === 'light'}" class="w-8 h-8 rounded-full flex items-center justify-center transition-all"><i class="fas fa-sun text-xs"></i></button>
          <button @click="theme = 'dark'" :class="{'bg-white dark:bg-slate-600 shadow-sm text-blue-600': theme === 'dark'}" class="w-8 h-8 rounded-full flex items-center justify-center transition-all ml-1"><i class="fas fa-moon text-xs"></i></button>
          <button @click="theme = 'auto'" :class="{'bg-white dark:bg-slate-600 shadow-sm text-blue-600': theme === 'auto'}" class="w-8 h-8 rounded-full flex items-center justify-center transition-all ml-1"><i class="fas fa-desktop text-xs"></i></button>
        </div>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative">
      <!-- Left Tree Sidebar -->
      <aside :class="{'translate-x-0': isMobileSidebarOpen, '-translate-x-full lg:translate-x-0': !isMobileSidebarOpen}" 
             class="fixed lg:relative inset-y-0 left-0 w-80 bg-white dark:bg-[#161b22] border-r border-slate-200 dark:border-slate-800 z-[60] lg:z-0 transition-transform duration-500 flex flex-col pt-16 lg:pt-0">
        <div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
           <div class="px-2 mb-4 text-[10px] font-black text-slate-400 uppercase tracking-widest flex items-center justify-between">
               <span>目录导航</span>
               <i v-if="isScanning" class="fas fa-circle-notch fa-spin text-blue-500"></i>
           </div>
           <div v-for="node in foldersTree" :key="node.FullPath">
              <RecursiveTree :node="node" :currentPath="currentPath" :navigateTo="navigateTo" :toggleFolder="toggleFolder" />
           </div>
        </div>
      </aside>

      <!-- Main Panel -->
      <main class="flex-1 flex flex-col bg-white dark:bg-[#0d1117] min-w-0 overflow-hidden relative shadow-inner">
        <!-- History & Mode Control Bar -->
        <div class="h-14 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between px-6 bg-slate-50/50 dark:bg-slate-800/10 shrink-0">
          <div class="flex items-center gap-3">
            <div class="flex items-center bg-white dark:bg-slate-800 rounded-lg shadow-sm border border-slate-200 dark:border-slate-700 p-0.5">
                <button @click="goBack" :disabled="historyIndex <= 0" class="w-8 h-8 flex items-center justify-center hover:text-blue-500 disabled:opacity-20 transition-all"><i class="fas fa-arrow-left text-sm"></i></button>
                <button @click="goForward" :disabled="historyIndex >= history.length - 1" class="w-8 h-8 flex items-center justify-center hover:text-blue-500 disabled:opacity-20 transition-all border-l border-slate-100 dark:border-slate-700"><i class="fas fa-arrow-right text-sm"></i></button>
            </div>
            <div v-if="isScanning" class="flex items-center gap-2 px-3 py-1.5 bg-blue-50 dark:bg-blue-900/30 border border-blue-100 dark:border-blue-800 rounded-full animate-pulse">
                <i class="fas fa-search-location text-blue-500 text-xs"></i>
                <span class="text-[10px] font-black text-blue-600 uppercase tracking-tighter">系统正在扫描库中，请耐心等待...</span>
            </div>
          </div>

          <div class="flex items-center gap-3">
             <div class="flex bg-slate-100 dark:bg-slate-900 p-1 rounded-lg">
                <button @click="viewMode = 'grid'" :class="{'bg-white dark:bg-slate-700 text-blue-600': viewMode === 'grid'}" class="w-8 h-8 rounded-md flex items-center justify-center transition-all"><i class="fas fa-th-large text-xs"></i></button>
                <button @click="viewMode = 'list'" :class="{'bg-white dark:bg-slate-700 text-blue-600': viewMode === 'list'}" class="w-8 h-8 rounded-md flex items-center justify-center transition-all ml-1"><i class="fas fa-list text-xs"></i></button>
             </div>
          </div>
        </div>

        <!-- File Grid/List -->
        <div class="flex-1 overflow-y-auto custom-scrollbar p-6">
          <div v-if="isLoading" class="absolute inset-0 bg-white/60 dark:bg-black/60 z-30 flex items-center justify-center backdrop-blur-sm"><i class="fas fa-circle-notch fa-spin text-4xl text-blue-500"></i></div>

          <div v-if="viewMode === 'grid'" class="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-6 2xl:grid-cols-8 gap-5">
             <div v-for="file in files" :key="file.FullPath" @dblclick="file.IsDir ? navigateTo(file.FullPath) : null" 
                  class="group relative bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-800 p-4 rounded-3xl cursor-pointer hover:bg-white dark:hover:bg-slate-800 hover:shadow-2xl hover:shadow-blue-500/10 hover:-translate-y-1 transition-all duration-300">
                <div class="aspect-square flex items-center justify-center mb-3">
                    <i class="fas text-4xl group-hover:scale-110 transition-transform duration-500" :class="file.IsDir ? 'fa-folder text-amber-500' : 'fa-file-alt text-blue-400'"></i>
                </div>
                <p class="text-[11px] font-black text-center truncate px-2 dark:text-slate-300">{{ file.Name }}</p>
             </div>
          </div>

          <!-- List View with sorting -->
          <table v-else class="w-full text-left border-collapse">
            <thead class="sticky top-0 bg-white dark:bg-[#0d1117] z-20 border-b border-slate-200 dark:border-slate-800">
               <tr class="text-[10px] font-black text-slate-400 uppercase tracking-widest">
                  <th class="px-4 py-4 cursor-pointer hover:text-blue-500 transition-colors" @click="sortKey='Name'; sortOrder*=-1">文件名</th>
                  <th class="px-4 py-4 text-right cursor-pointer hover:text-blue-500 transition-colors" @click="sortKey='Size'; sortOrder*=-1">大小</th>
                  <th class="px-4 py-4 text-right hidden md:table-cell">修改日期</th>
               </tr>
            </thead>
            <tbody class="divide-y divide-slate-50 dark:divide-slate-800/50">
               <tr v-for="file in files" :key="file.FullPath" @dblclick="file.IsDir ? navigateTo(file.FullPath) : null" class="hover:bg-blue-50/40 dark:hover:bg-blue-900/10 transition-all cursor-pointer group">
                  <td class="px-4 py-4 flex items-center gap-3">
                    <i class="fas text-lg" :class="file.IsDir ? 'fa-folder text-amber-500' : 'fa-file-alt text-blue-400'"></i>
                    <span class="text-sm font-bold group-hover:text-blue-600 transition-colors">{{ file.Name }}</span>
                  </td>
                  <td class="px-4 py-4 text-right text-xs font-mono text-slate-400">{{ file.IsDir ? '—' : (file.Size / 1024 / 1024).toFixed(1) + ' MB' }}</td>
                  <td class="px-4 py-4 text-right text-xs text-slate-500 hidden md:table-cell">{{ new Date(file.ModTime).toLocaleDateString() }}</td>
               </tr>
            </tbody>
          </table>

          <div v-if="!isLoading && !files.length" class="flex flex-col items-center py-48 opacity-10 grayscale"><i class="fas fa-box-open text-8xl mb-6"></i><p class="text-2xl font-black">暂无数据</p></div>
        </div>
      </main>

      <!-- Overlay for mobile drawer -->
      <div v-if="isMobileSidebarOpen" @click="isMobileSidebarOpen = false" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[55] lg:hidden animate-fade-in"></div>
    </div>
  </div>
</template>

<script>
// --- Recursive Tree Component ---
const RecursiveTree = {
    name: 'RecursiveTree',
    props: ['node', 'currentPath', 'navigateTo', 'toggleFolder'],
    template: `
        <div class="tree-node mb-1">
            <div @click="toggleFolder(node)" 
                 class="flex items-center gap-2.5 p-2 rounded-xl cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800 transition-all"
                 :class="{'bg-blue-50 dark:bg-blue-900/20 text-blue-600 font-bold': currentPath === node.FullPath}">
                <div class="w-4 h-4 flex items-center justify-center shrink-0">
                    <i v-if="node.loading" class="fas fa-circle-notch fa-spin text-[10px]"></i>
                    <i v-else class="fas fa-caret-right transition-transform duration-300" :class="{'rotate-90': node.isOpen, 'opacity-0': node.loaded && !node.children.length}"></i>
                </div>
                <i class="fas text-sm shrink-0" :class="node.FullPath === '' ? 'fa-home' : (node.isOpen ? 'fa-folder-open text-amber-500' : 'fa-folder text-amber-500')"></i>
                <span @click.stop="navigateTo(node.FullPath)" class="truncate text-sm flex-1">{{ node.Name }}</span>
            </div>
            <div v-if="node.isOpen" class="ml-5 border-l border-slate-100 dark:border-slate-800 pl-2 mt-1 space-y-1">
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
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
.animate-fade-in { animation: fadeIn 0.3s ease-out; }
</style>
