<script setup>
import { ref, onMounted, computed, watch, onUnmounted, defineComponent, h } from 'vue'

/**
 * ModernFM - Full Feature Enterprise Edition
 * Full Implementation: Upload, Download, Context Menu, Tab Favicon
 */

// --- 0. Recursive Component Definition ---
// 使用 defineComponent 并在 setup 中定义递归组件
const RecursiveTree = defineComponent({
    name: 'RecursiveTree',
    props: {
        node: Object,
        currentPath: String,
        navigateTo: Function,
        toggleFolder: Function
    },
    template: `
        <div class="tree-node mb-1">
            <div @click="toggleFolder(node)" class="flex items-center gap-2.5 p-2 rounded-xl cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800" :class="{'bg-blue-50 dark:bg-blue-900/20 text-blue-600 font-bold shadow-sm': currentPath === node.FullPath}">
                <div class="w-4 h-4 flex items-center justify-center shrink-0">
                    <i v-if="node.loading" class="fas fa-circle-notch fa-spin text-[10px]"></i>
                    <i v-else-if="node.IsDir" class="fas fa-caret-right transition-transform" :class="{'rotate-90': node.isOpen, 'opacity-0': node.loaded && !node.children.length}"></i>
                </div>
                <i class="fas text-sm" :class="node.FullPath === '' ? 'fa-home' : (node.isOpen ? 'fa-folder-open text-amber-500' : 'fa-folder text-amber-500')"></i>
                <span @click.stop="navigateTo(node.FullPath)" class="truncate text-sm flex-1">{{ node.Name }}</span>
            </div>
            <div v-if="node.isOpen && node.children && node.children.length" class="ml-5 border-l border-slate-200 dark:border-slate-800 pl-2 mt-1 space-y-1">
                <RecursiveTree v-for="child in node.children" :key="child.FullPath" :node="child" :currentPath="currentPath" :navigateTo="navigateTo" :toggleFolder="toggleFolder" />
            </div>
        </div>
    `
})

/**
 * ModernFM - Full Feature Enterprise Edition
 * Full Implementation: Upload, Download, Context Menu, Tab Favicon
 */

const THEME_KEY = 'fm-theme'
const PATH_KEY = 'fm-last-path'

const theme = ref(localStorage.getItem(THEME_KEY) || 'auto')
const currentPath = ref(localStorage.getItem(PATH_KEY) || '')
const history = ref([currentPath.value])
const historyIndex = ref(0)
const files = ref([])
const foldersTree = ref([])
const isLoading = ref(false)
const isSettingsOpen = ref(false)
const isMobileSidebarOpen = ref(false)
const selectedPaths = ref(new Set())
const contextMenu = ref({ show: false, x: 0, y: 0, target: null })

// --- 1. Tab 图标支持 (动态修改 Favicon) ---
const updateTabIcon = () => {
    let link = document.querySelector("link[rel~='icon']");
    if (!link) {
        link = document.createElement('link');
        link.rel = 'icon';
        document.getElementsByTagName('head')[0].appendChild(link);
    }
    // 使用 FontAwesome 的 SVG 转换或简单使用 Emoji 字符作为图标
    link.href = "data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22><text y=%22.9em%22 font-size=%2290%22>🍀</text></svg>";
}

// --- 2. 主题管理 ---
const applyTheme = (v) => {
    const isDark = v === 'dark' || (v === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)
    document.documentElement.classList.toggle('dark', isDark)
}
watch(theme, (v) => { localStorage.setItem(THEME_KEY, v); applyTheme(v) })

// --- 3. 核心数据交互 ---
const fetchFiles = async (path = '', skipHistory = false) => {
  isLoading.value = true
  try {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`)
    const data = await res.json()
    files.value = Array.isArray(data) ? data.sort((a, b) => (b.IsDir - a.IsDir) || a.Name.localeCompare(b.Name)) : []
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

const fetchSubFolders = async (p) => {
    try {
        const res = await fetch(`/api/files/list?path=${encodeURIComponent(p)}`)
        if (!res.ok) throw new Error('API Error')
        const data = await res.json()
        return Array.isArray(data) ? data.filter(f => f.IsDir).sort((a,b) => a.Name.localeCompare(b.Name)) : []
    } catch (e) {
        console.error('fetchSubFolders error:', e)
        return []
    }
}

const initializeTree = async () => {
    console.log('Initializing tree...');
    try {
        const rootFolders = await fetchSubFolders('')
        console.log('Root folders fetched:', rootFolders);
        // 关键点：将数据放入一个临时的普通数组，然后再整体赋值给 foldersTree.value
        const rootNode = { 
            Name: '资源库', 
            FullPath: '', 
            IsDir: true,
            children: rootFolders.map(f => ({ ...f, children: [], isOpen: false, loaded: false })), 
            isOpen: true, 
            loaded: true 
        };
        foldersTree.value = [rootNode];
        console.log('Folders tree set successfully:', foldersTree.value);
    } catch (e) {
        console.error('Failed to initialize tree:', e)
        foldersTree.value = [{ Name: '资源库 (加载失败)', FullPath: '', IsDir: true, children: [], isOpen: true, loaded: true }]
    }
}

// 强制刷新树的方法
const refreshTree = () => initializeTree()

const toggleFolder = async (node) => {
    if (!node.IsDir) return
    node.isOpen = !node.isOpen
    if (node.isOpen && !node.loaded) {
        node.loading = true
        try {
            const subs = await fetchSubFolders(node.FullPath)
            node.children = subs.map(f => ({ ...f, children: [], isOpen: false, loaded: false }))
            node.loaded = true
        } catch (e) {
            console.error('toggleFolder error:', e)
        } finally {
            node.loading = false
        }
    }
}

// --- 4. 文件管理功能实现 ---

// 下载
const downloadFile = (file) => {
    if (file.IsDir) return alert('暂不支持文件夹直接下载，请使用压缩功能')
    window.open(`/api/files/download?path=${encodeURIComponent(file.FullPath)}`, '_blank')
}

// 上传 (简单实现)
const triggerUpload = () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.onchange = async (e) => {
        const file = e.target.files[0]
        const formData = new FormData()
        formData.append('file', file)
        formData.append('path', currentPath.value)
        formData.append('filename', file.name)
        isLoading.value = true
        await fetch('/api/files/upload', { method: 'POST', body: formData })
        fetchFiles(currentPath.value, true)
    }
    input.click()
}

// 通用操作 (删除/重命名)
const executeAction = async (action) => {
    const targets = Array.from(selectedPaths.value)
    if (!targets.length) return
    
    if (action === 'delete') {
        if (!confirm(`确定要永久删除这 ${targets.length} 项资源吗？`)) return
    }

    if (action === 'rename') {
        const newName = prompt('请输入新名称:', files.value.find(f => f.FullPath === targets[0]).Name)
        if (!newName) return
        await fetch('/api/files/action', { 
            method: 'POST', 
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ action, paths: targets, newName }) 
        })
    } else {
        await fetch('/api/files/action', { 
            method: 'POST', 
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({ action, paths: targets }) 
        })
    }
    
    closeContextMenu()
    fetchFiles(currentPath.value, true)
}

// --- 右键菜单控制 ---
const openContextMenu = (e, file) => {
    e.preventDefault()
    if (!selectedPaths.value.has(file.FullPath)) {
        selectedPaths.value.clear()
        selectedPaths.value.add(file.FullPath)
    }
    contextMenu.value = { show: true, x: e.clientX, y: e.clientY, target: file }
}
const closeContextMenu = () => { contextMenu.value.show = false }

const navigateTo = (p) => { fetchFiles(p); isMobileSidebarOpen.value = false }

onMounted(() => {
    updateTabIcon()
    fetchFiles(currentPath.value)
    initializeTree()
    window.addEventListener('click', closeContextMenu)
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
            <i class="fas fa-cube text-xl"></i>
          </div>
          <span class="text-2xl font-black tracking-tighter hidden sm:block italic">ModernFM</span>
        </div>
      </div>

      <nav class="hidden md:flex items-center bg-slate-100 dark:bg-slate-900 px-6 py-2 rounded-full border border-slate-200 dark:border-slate-800 flex-1 mx-12 overflow-hidden shadow-inner">
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

        <div v-if="isSettingsOpen" class="absolute top-14 right-0 w-64 bg-white dark:bg-[#1c2128] border border-slate-200 dark:border-slate-700 rounded-3xl shadow-2xl p-2 z-[110] animate-fade-in">
           <button @click="theme = 'light'" class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 text-sm"><i class="fas fa-sun w-5"></i> 明亮模式</button>
           <button @click="theme = 'dark'" class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 text-sm"><i class="fas fa-moon w-5"></i> 暗黑模式</button>
           <button @click="theme = 'auto'" class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 text-sm"><i class="fas fa-desktop w-5"></i> 跟随系统</button>
           <div class="h-px bg-slate-100 dark:bg-slate-800 my-2 mx-2"></div>
           <button @click="refreshTree" class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 text-sm"><i class="fas fa-project-diagram w-5"></i> 重载目录树</button>
           <button @click="fetchFiles(currentPath, true)" class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800 text-sm"><i class="fas fa-sync-alt w-5"></i> 刷新文件库</button>
        </div>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative">
      <!-- Left Tree -->
      <aside :class="{'translate-x-0': isMobileSidebarOpen, '-translate-x-full lg:translate-x-0': !isMobileSidebarOpen}" 
             class="fixed lg:relative inset-y-0 left-0 w-80 bg-white dark:bg-[#161b22] border-r border-slate-200 dark:border-slate-800 z-[60] lg:z-0 transition-transform duration-500 flex flex-col pt-16 lg:pt-0">
        <div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
           <div v-if="foldersTree.length === 0" class="p-4 text-slate-400 text-sm italic">
             加载中...
           </div>
           <div v-for="node in foldersTree" :key="node.FullPath">
              <RecursiveTree :node="node" :currentPath="currentPath" :navigateTo="navigateTo" :toggleFolder="toggleFolder" />
           </div>
        </div>
      </aside>

      <!-- Main Panel -->
      <main class="flex-1 flex flex-col bg-white dark:bg-[#0d1117] min-w-0 overflow-hidden relative shadow-inner">
        <div class="h-14 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between px-6 bg-slate-50/50 dark:bg-slate-800/10 shrink-0">
          <div class="flex items-center gap-4">
            <button @click="historyIndex--; fetchFiles(history[historyIndex], true)" :disabled="historyIndex <= 0" class="w-9 h-9 flex items-center justify-center bg-white dark:bg-slate-800 rounded-full border border-slate-200 dark:border-slate-700 shadow-sm disabled:opacity-20"><i class="fas fa-chevron-left"></i></button>
            <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">{{ files.length }} 项内容</span>
          </div>
          <button @click="triggerUpload" class="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-2xl text-xs font-black shadow-lg shadow-blue-500/20 active:scale-95 transition-all">
            <i class="fas fa-plus mr-2"></i>上传文件
          </button>
        </div>

        <div class="flex-1 overflow-y-auto custom-scrollbar p-6" @contextmenu.self="contextMenu.show = false">
          <div v-if="isLoading" class="absolute inset-0 bg-white/60 dark:bg-black/60 z-[90] flex items-center justify-center backdrop-blur-sm"><i class="fas fa-spinner fa-spin text-4xl text-blue-600"></i></div>
          
          <div class="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-6 2xl:grid-cols-8 gap-5">
             <div v-for="(file, idx) in files" :key="file.FullPath" 
                  @click="selectedPaths.has(file.FullPath) ? selectedPaths.delete(file.FullPath) : selectedPaths.add(file.FullPath)"
                  @dblclick="file.IsDir ? navigateTo(file.FullPath) : downloadFile(file)"
                  @contextmenu="openContextMenu($event, file)"
                  class="group relative bg-slate-50 dark:bg-slate-900/50 border p-4 rounded-3xl cursor-pointer hover:bg-white dark:hover:bg-slate-800 transition-all duration-300"
                  :class="selectedPaths.has(file.FullPath) ? 'border-blue-500 shadow-xl shadow-blue-500/10' : 'border-slate-100 dark:border-slate-800'">
                <div class="aspect-square flex items-center justify-center mb-3">
                    <i class="fas text-4xl" :class="file.IsDir ? 'fa-folder text-amber-500' : 'fa-file-alt text-blue-400'"></i>
                </div>
                <p class="text-[11px] font-black text-center truncate px-2 text-slate-600 dark:text-slate-300 group-hover:text-blue-600 transition-colors">{{ file.Name }}</p>
                <div v-if="selectedPaths.has(file.FullPath)" class="absolute top-3 right-3 w-5 h-5 bg-blue-600 rounded-full flex items-center justify-center"><i class="fas fa-check text-[10px] text-white"></i></div>
             </div>
          </div>
        </div>
      </main>

      <!-- Context Menu -->
      <div v-if="contextMenu.show" :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }" class="fixed w-56 bg-white dark:bg-[#1c2128] border border-slate-200 dark:border-slate-700 rounded-3xl shadow-2xl p-2 z-[200] animate-fade-in shadow-blue-500/10">
         <button @click="contextMenu.target.IsDir ? navigateTo(contextMenu.target.FullPath) : downloadFile(contextMenu.target)" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-blue-600 hover:text-white text-sm transition-all"><i class="fas fa-play w-4"></i> {{ contextMenu.target.IsDir ? '打开文件夹' : '下载/预览' }}</button>
         <button @click="executeAction('rename')" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-700 text-sm transition-all"><i class="fas fa-edit w-4"></i> 重命名</button>
         <div class="h-px bg-slate-100 dark:bg-slate-800 my-1 mx-2"></div>
         <button @click="executeAction('delete')" class="w-full flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-red-600 hover:text-white text-red-500 text-sm transition-all font-bold"><i class="fas fa-trash w-4"></i> 删除</button>
      </div>

      <div v-if="isMobileSidebarOpen" @click="isMobileSidebarOpen = false" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[55] lg:hidden animate-fade-in"></div>
    </div>
  </div>
</template>

<style>
@import url('https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css');
.custom-scrollbar::-webkit-scrollbar { width: 5px; height: 5px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 20px; }
.dark .custom-scrollbar::-webkit-scrollbar-thumb { background: #334155; }
@keyframes fadeIn { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }
.animate-fade-in { animation: fadeIn 0.2s ease-out; }
</style>
