<script setup>
import { ref, onMounted, computed, watch, onUnmounted } from 'vue'

/**
 * ModernFM - Advanced File Manager UI
 * Features: Lazy Loading Tree, Context Menu, Multi-selection, Multi-view
 */

// --- Constants & Persistence ---
const THEME_KEY = 'fm-theme'
const PATH_KEY = 'fm-last-path'
const VIEW_MODE_KEY = 'fm-view-mode'

// --- State Management ---
const theme = ref(localStorage.getItem(THEME_KEY) || 'light')
const currentPath = ref(localStorage.getItem(PATH_KEY) || '')
const viewMode = ref(localStorage.getItem(VIEW_MODE_KEY) || 'grid') // grid, list, small-grid
const sortKey = ref('Name')
const sortOrder = ref(1) // 1: asc, -1: desc

const files = ref([])
const foldersTree = ref([]) // Lazy loaded tree
const isLoading = ref(false)
const searchQuery = ref('')
const isMobileSidebarOpen = ref(false)

// Selection State
const selectedPaths = ref(new Set())
const lastSelectedIndex = ref(-1)

// Context Menu State
const contextMenu = ref({ show: false, x: 0, y: 0, target: null })

// --- Theme Logic ---
const applyTheme = (val) => document.documentElement.classList.toggle('dark', val === 'dark')
const toggleTheme = () => {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
  localStorage.setItem(THEME_KEY, theme.value)
  applyTheme(theme.value)
}

// --- Data Fetching (Main View) ---
const fetchFiles = async (path = '') => {
  isLoading.value = true
  try {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`)
    const data = await res.json()
    files.value = Array.isArray(data) ? data.map(f => ({ ...f, selected: false })) : []
    currentPath.value = path
    localStorage.setItem(PATH_KEY, path)
    selectedPaths.value.clear() // Clear selection on navigate
  } catch (e) { console.error(e) } finally { isLoading.value = false }
}

// --- Tree Logic (Lazy Loading) ---
const fetchSubFolders = async (parentPath) => {
    try {
        const res = await fetch(`/api/files/list?path=${encodeURIComponent(parentPath)}`)
        const data = await res.json()
        return Array.isArray(data) ? data.filter(f => f.IsDir).sort((a,b) => a.Name.localeCompare(b.Name)) : []
    } catch (e) { return [] }
}

const initializeTree = async () => {
    const rootFolders = await fetchSubFolders('')
    foldersTree.value = [{
        Name: '根目录',
        FullPath: '',
        children: rootFolders.map(f => ({ ...f, children: [], isOpen: false, loaded: false })),
        isOpen: true,
        loaded: true
    }]
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

// --- Selection Logic ---
const toggleSelect = (file, index, event) => {
    if (event.shiftKey && lastSelectedIndex.value !== -1) {
        const start = Math.min(lastSelectedIndex.value, index)
        const end = Math.max(lastSelectedIndex.value, index)
        for (let i = start; i <= end; i++) selectedPaths.value.add(sortedFiles.value[i].FullPath)
    } else if (event.ctrlKey || event.metaKey) {
        if (selectedPaths.value.has(file.FullPath)) selectedPaths.value.delete(file.FullPath)
        else selectedPaths.value.add(file.FullPath)
    } else {
        selectedPaths.value.clear()
        selectedPaths.value.add(file.FullPath)
    }
    lastSelectedIndex.value = index
}

// --- Context Menu ---
const openContextMenu = (e, target = null) => {
    e.preventDefault()
    if (target && !selectedPaths.value.has(target.FullPath)) {
        selectedPaths.value.clear()
        selectedPaths.value.add(target.FullPath)
    }
    contextMenu.value = { show: true, x: e.clientX, y: e.clientY, target }
}

const closeContextMenu = () => { contextMenu.value.show = false }

// --- File Operations (Stubs for now) ---
const executeAction = (action) => {
    const targets = Array.from(selectedPaths.value)
    alert(`执行操作: ${action} 对目标: ${targets.length} 项`)
    closeContextMenu()
}

// --- Navigation ---
const navigateTo = (p) => { fetchFiles(p); isMobileSidebarOpen.value = false }

const breadcrumbs = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  let acc = ''
  return [{ name: '首页', path: '' }, ...parts.map(p => ({ name: p, path: acc = acc ? `${acc}/${p}` : p }))]
})

const sortedFiles = computed(() => {
  const q = searchQuery.value.toLowerCase()
  let list = files.value.filter(f => f.Name.toLowerCase().includes(q))
  
  return list.sort((a, b) => {
    if (a.IsDir !== b.IsDir) return b.IsDir ? 1 : -1
    let valA = a[sortKey.value], valB = b[sortKey.value]
    if (typeof valA === 'string') return valA.localeCompare(valB) * sortOrder.value
    return (valA - valB) * sortOrder.value
  })
})

const setSort = (key) => {
    if (sortKey.value === key) sortOrder.value *= -1
    else { sortKey.value = key; sortOrder.value = 1 }
}

const switchView = (mode) => {
    viewMode.value = mode
    localStorage.setItem(VIEW_MODE_KEY, mode)
}

onMounted(() => {
  applyTheme(theme.value)
  fetchFiles(currentPath.value)
  initializeTree()
  window.addEventListener('click', closeContextMenu)
})
onUnmounted(() => window.removeEventListener('click', closeContextMenu))

const formatSize = (b) => {
  if (!b) return '—'
  const i = Math.floor(Math.log(b) / Math.log(1024))
  return (b / Math.pow(1024, i)).toFixed(1) + ' ' + ['B', 'KB', 'MB', 'GB', 'TB'][i]
}
</script>

<template>
  <div class="fixed inset-0 flex flex-col bg-gray-50 dark:bg-[#0d1117] text-gray-900 dark:text-gray-100 transition-colors duration-300 font-sans antialiased overflow-hidden select-none">
    
    <!-- Header -->
    <header class="h-16 border-b border-gray-200 dark:border-gray-800 bg-white/95 dark:bg-[#161b22]/95 backdrop-blur-md flex items-center justify-between px-6 z-50 shrink-0 shadow-sm">
      <div class="flex items-center gap-4">
        <button @click="isMobileSidebarOpen = true" class="lg:hidden p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg"><i class="fas fa-bars text-xl"></i></button>
        <div class="flex items-center gap-2 cursor-pointer group" @click="navigateTo('')">
          <div class="w-10 h-10 bg-blue-600 rounded-2xl flex items-center justify-center text-white shadow-lg shadow-blue-600/20 group-hover:scale-105 transition-transform"><i class="fas fa-folder-tree text-lg"></i></div>
          <span class="text-2xl font-black tracking-tighter hidden sm:block">ModernFM</span>
        </div>
      </div>

      <nav class="hidden md:flex items-center bg-gray-100/50 dark:bg-gray-900/50 px-6 py-2 rounded-full border border-gray-200 dark:border-gray-800 flex-1 mx-12 overflow-hidden shadow-inner">
        <template v-for="(bc, i) in breadcrumbs" :key="bc.path">
          <i v-if="i > 0" class="fas fa-chevron-right text-[10px] mx-3 opacity-30 text-gray-400"></i>
          <button @click="navigateTo(bc.path)" class="text-sm hover:text-blue-500 font-bold transition-colors truncate" :class="{'text-blue-600': i === breadcrumbs.length - 1}">{{ bc.name }}</button>
        </template>
      </nav>

      <div class="flex items-center gap-3">
        <div class="relative hidden lg:block mr-2">
          <input v-model="searchQuery" type="text" placeholder="快速查找..." class="w-64 bg-gray-100 dark:bg-gray-900 border-none rounded-full py-2.5 pl-12 pr-4 text-xs focus:ring-4 focus:ring-blue-500/10 transition-all">
          <i class="fas fa-search absolute left-5 top-1/2 -translate-y-1/2 text-gray-400"></i>
        </div>
        <button @click="toggleTheme" class="w-10 h-10 rounded-full flex items-center justify-center bg-white dark:bg-gray-800 hover:shadow-md border border-gray-200 dark:border-gray-700 transition-all"><i class="fas" :class="theme === 'light' ? 'fa-moon text-gray-600' : 'fa-sun text-amber-400'"></i></button>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative" @contextmenu.self="openContextMenu">
      <!-- Sidebar Tree (Lazy) -->
      <aside :class="{'translate-x-0': isMobileSidebarOpen, '-translate-x-full lg:translate-x-0': !isMobileSidebarOpen}" 
             class="fixed lg:relative inset-y-0 left-0 w-80 bg-white dark:bg-[#161b22] border-r border-gray-200 dark:border-gray-800 z-[60] lg:z-0 transition-transform duration-500 ease-out flex flex-col pt-16 lg:pt-0">
        <div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
           <div class="px-2 mb-4 text-[10px] font-black text-gray-400 uppercase tracking-widest">资源目录树</div>
           <div v-for="node in foldersTree" :key="node.FullPath">
              <RecursiveTree :node="node" :currentPath="currentPath" :navigateTo="navigateTo" :toggleFolder="toggleFolder" />
           </div>
        </div>
      </aside>

      <!-- Main Container -->
      <main class="flex-1 flex flex-col bg-white dark:bg-[#0d1117] min-w-0 overflow-hidden relative shadow-inner">
        <!-- View Control Bar -->
        <div class="h-14 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between px-6 bg-white dark:bg-[#0d1117] shrink-0">
          <div class="flex items-center gap-4">
             <div class="flex items-center bg-gray-100 dark:bg-gray-800 p-1 rounded-xl shadow-inner">
                <button @click="switchView('grid')" :class="{'bg-white dark:bg-gray-700 shadow-sm text-blue-600': viewMode === 'grid'}" class="w-8 h-8 rounded-lg flex items-center justify-center transition-all"><i class="fas fa-th-large"></i></button>
                <button @click="switchView('small-grid')" :class="{'bg-white dark:bg-gray-700 shadow-sm text-blue-600': viewMode === 'small-grid'}" class="w-8 h-8 rounded-lg flex items-center justify-center transition-all"><i class="fas fa-th"></i></button>
                <button @click="switchView('list')" :class="{'bg-white dark:bg-gray-700 shadow-sm text-blue-600': viewMode === 'list'}" class="w-8 h-8 rounded-lg flex items-center justify-center transition-all"><i class="fas fa-list"></i></button>
             </div>
             <div class="h-6 w-px bg-gray-200 dark:border-gray-700 mx-2"></div>
             <span class="text-[10px] font-black text-gray-400 uppercase tracking-widest">{{ selectedPaths.size }} / {{ sortedFiles.length }} 已选择</span>
          </div>
          <div class="flex items-center gap-2">
             <button @click="setSort('Name')" class="px-3 py-1.5 text-[10px] font-bold rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800" :class="{'text-blue-600': sortKey === 'Name'}">按名称</button>
             <button @click="setSort('Size')" class="px-3 py-1.5 text-[10px] font-bold rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800" :class="{'text-blue-600': sortKey === 'Size'}">按大小</button>
             <button @click="fetchFiles(currentPath)" class="w-9 h-9 flex items-center justify-center text-gray-400 hover:text-blue-500 rounded-xl"><i class="fas fa-sync-alt"></i></button>
          </div>
        </div>

        <!-- Scrollable Content -->
        <div class="flex-1 overflow-y-auto custom-scrollbar p-6" @mousedown.self="selectedPaths.clear()">
          <div v-if="isLoading" class="absolute inset-0 bg-white/60 dark:bg-black/60 z-30 flex items-center justify-center backdrop-blur-sm"><i class="fas fa-spinner fa-spin text-4xl text-blue-500"></i></div>

          <!-- Grid View -->
          <div v-if="viewMode === 'grid' || viewMode === 'small-grid'" class="grid gap-4" :class="viewMode === 'grid' ? 'grid-cols-2 md:grid-cols-3 xl:grid-cols-5' : 'grid-cols-3 md:grid-cols-5 xl:grid-cols-8'">
            <div v-for="(file, idx) in sortedFiles" :key="file.FullPath" 
                 @click="toggleSelect(file, idx, $event)" @dblclick="file.IsDir ? navigateTo(file.FullPath) : executeAction('preview')"
                 @contextmenu="openContextMenu($event, file)"
                 class="group relative p-4 rounded-3xl border transition-all duration-300 cursor-pointer"
                 :class="selectedPaths.has(file.FullPath) ? 'bg-blue-600/10 border-blue-600 shadow-lg shadow-blue-500/10' : 'bg-gray-50/50 dark:bg-[#161b22]/50 border-transparent hover:border-gray-200 dark:hover:border-gray-700'">
                <div class="aspect-square flex items-center justify-center mb-4 transition-transform duration-500 group-hover:scale-110">
                   <i class="fas text-4xl" :class="file.IsDir ? 'fa-folder text-amber-500' : 'fa-file-alt text-blue-400'"></i>
                </div>
                <div class="text-center">
                    <p class="text-xs font-bold truncate px-1" :class="selectedPaths.has(file.FullPath) ? 'text-blue-600' : ''">{{ file.Name }}</p>
                    <p v-if="viewMode === 'grid'" class="text-[9px] font-black text-gray-400 uppercase mt-1 tracking-tighter">{{ file.IsDir ? '目录' : formatSize(file.Size) }}</p>
                </div>
                <div v-if="selectedPaths.has(file.FullPath)" class="absolute top-3 right-3 w-5 h-5 bg-blue-600 rounded-full flex items-center justify-center shadow-lg"><i class="fas fa-check text-[10px] text-white"></i></div>
            </div>
          </div>

          <!-- List View -->
          <table v-if="viewMode === 'list'" class="w-full text-left">
            <thead class="sticky top-0 bg-white/90 dark:bg-[#0d1117]/90 z-20 border-b border-gray-200 dark:border-gray-800">
               <tr class="text-[10px] font-black text-gray-400 uppercase tracking-widest">
                  <th class="px-6 py-4 w-12 text-center"></th>
                  <th class="px-4 py-4 cursor-pointer" @click="setSort('Name')">名称 <i v-if="sortKey === 'Name'" class="fas" :class="sortOrder > 0 ? 'fa-arrow-up' : 'fa-arrow-down'"></i></th>
                  <th class="px-4 py-4 text-right cursor-pointer" @click="setSort('Size')">大小 <i v-if="sortKey === 'Size'" class="fas" :class="sortOrder > 0 ? 'fa-arrow-up' : 'fa-arrow-down'"></i></th>
                  <th class="px-4 py-4 text-right cursor-pointer" @click="setSort('ModTime')">修改日期</th>
               </tr>
            </thead>
            <tbody class="divide-y divide-gray-50 dark:divide-gray-800/50">
               <tr v-for="(file, idx) in sortedFiles" :key="file.FullPath" 
                   @click="toggleSelect(file, idx, $event)" @dblclick="file.IsDir ? navigateTo(file.FullPath) : executeAction('preview')"
                   @contextmenu="openContextMenu($event, file)"
                   class="group hover:bg-gray-100/50 dark:hover:bg-blue-900/5 transition-colors cursor-pointer"
                   :class="selectedPaths.has(file.FullPath) ? 'bg-blue-50/50 dark:bg-blue-900/20' : ''">
                  <td class="px-6 py-4 text-center">
                    <div class="w-5 h-5 border-2 rounded transition-all flex items-center justify-center" :class="selectedPaths.has(file.FullPath) ? 'bg-blue-600 border-blue-600' : 'border-gray-200 dark:border-gray-700'">
                       <i v-if="selectedPaths.has(file.FullPath)" class="fas fa-check text-white text-[10px]"></i>
                    </div>
                  </td>
                  <td class="px-4 py-4">
                    <div class="flex items-center gap-3">
                       <i class="fas text-lg" :class="file.IsDir ? 'fa-folder text-amber-500' : 'fa-file-alt text-blue-400'"></i>
                       <span class="text-sm font-bold truncate max-w-xl">{{ file.Name }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-4 text-right text-xs font-mono text-gray-400">{{ formatSize(file.Size) }}</td>
                  <td class="px-4 py-4 text-right text-xs text-gray-400">{{ new Date(file.ModTime).toLocaleDateString() }}</td>
               </tr>
            </tbody>
          </table>

          <div v-if="!isLoading && !sortedFiles.length" class="flex flex-col items-center py-48 opacity-10"><i class="fas fa-box-open text-8xl mb-6"></i><p class="text-2xl font-black">空空如也</p></div>
        </div>
      </main>

      <!-- Context Menu -->
      <div v-if="contextMenu.show" :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }" 
           class="fixed w-56 bg-white dark:bg-[#1c2128] border border-gray-200 dark:border-gray-700 rounded-2xl shadow-2xl z-[100] py-2 px-1 backdrop-blur-lg animate-fade-in">
          <div class="px-3 py-2 text-[10px] font-black text-gray-400 uppercase tracking-tighter border-b border-gray-100 dark:border-gray-800 mb-1">已选择 {{ selectedPaths.size }} 项</div>
          <button @click="executeAction('preview')" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm hover:bg-blue-600 hover:text-white transition-all"><i class="fas fa-eye w-5"></i> 快速预览</button>
          <button @click="executeAction('download')" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm hover:bg-blue-600 hover:text-white transition-all"><i class="fas fa-download w-5"></i> 下载资源</button>
          <div class="h-px bg-gray-100 dark:bg-gray-800 my-1 mx-2"></div>
          <button @click="executeAction('rename')" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-all"><i class="fas fa-edit w-5"></i> 重命名</button>
          <button @click="executeAction('copy')" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-all"><i class="fas fa-copy w-5"></i> 复制到...</button>
          <button @click="executeAction('move')" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-all"><i class="fas fa-file-export w-5"></i> 移动到...</button>
          <div class="h-px bg-gray-100 dark:bg-gray-800 my-1 mx-2"></div>
          <button @click="executeAction('compress')" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-all"><i class="fas fa-file-archive w-5"></i> ZIP 压缩</button>
          <button @click="executeAction('info')" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-all"><i class="fas fa-info-circle w-5"></i> 详细信息</button>
          <div class="h-px bg-gray-100 dark:bg-gray-800 my-1 mx-2"></div>
          <button @click="executeAction('delete')" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl text-sm hover:bg-red-600 hover:text-white text-red-500 transition-all"><i class="fas fa-trash-alt w-5"></i> 永久删除</button>
      </div>

      <div v-if="isMobileSidebarOpen" @click="isMobileSidebarOpen = false" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[55] lg:hidden"></div>
    </div>
  </div>
</template>

<script>
// --- Recursive Tree Component ---
const RecursiveTree = {
    name: 'RecursiveTree',
    props: ['node', 'currentPath', 'navigateTo', 'toggleFolder'],
    template: `
        <div class="tree-node">
            <div @click="node.children && toggleFolder(node)" 
                 class="flex items-center gap-2.5 p-2.5 rounded-2xl cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 transition-all"
                 :class="{'bg-blue-600/10 text-blue-600 font-bold': currentPath === node.FullPath}">
                <div class="w-4 h-4 flex items-center justify-center">
                    <i v-if="node.loading" class="fas fa-circle-notch fa-spin text-[10px]"></i>
                    <i v-else class="fas fa-caret-right transition-transform duration-300" :class="{'rotate-90': node.isOpen, 'opacity-0': node.loaded && !node.children.length}"></i>
                </div>
                <i class="fas text-base" :class="node.FullPath === '' ? 'fa-home' : (node.isOpen ? 'fa-folder-open text-amber-500' : 'fa-folder text-amber-500')"></i>
                <span @click.stop="navigateTo(node.FullPath)" class="truncate text-sm flex-1">{{ node.Name }}</span>
            </div>
            <div v-if="node.isOpen" class="ml-6 border-l-2 border-gray-50 dark:border-gray-800/50 pl-2 space-y-1">
                <RecursiveTree v-for="child in node.children" :key="child.FullPath" :node="child" :currentPath="currentPath" :navigateTo="navigateTo" :toggleFolder="toggleFolder" />
            </div>
        </div>
    `
}
</script>

<style>
@import url('https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css');
.custom-scrollbar::-webkit-scrollbar { width: 5px; height: 5px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #e5e7eb; border-radius: 20px; }
.dark .custom-scrollbar::-webkit-scrollbar-thumb { background: #30363d; }
@keyframes fadeIn { from { opacity: 0; transform: translateY(-5px); } to { opacity: 1; transform: translateY(0); } }
.animate-fade-in { animation: fadeIn 0.15s ease-out; }
</style>
