<script setup>
import { ref, onMounted, computed, watch, onUnmounted } from 'vue'

/**
 * ModernFM - CloudDrive2 Inspired UI
 * Manual Rewrite for Maximum Precision
 */

const theme = ref(localStorage.getItem('fm-theme') || 'light')
const currentPath = ref(localStorage.getItem('fm-last-path') || '')
const files = ref([])
const foldersTree = ref([])
const isLoading = ref(false)
const searchQuery = ref('')
const isMobileSidebarOpen = ref(false)

// --- 主题 ---
const toggleTheme = () => {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
  localStorage.setItem('fm-theme', theme.value)
}
watch(theme, (v) => document.documentElement.classList.toggle('dark', v === 'dark'), { immediate: true })

// --- 数据获取 ---
const fetchFiles = async (path = '') => {
  isLoading.value = true
  try {
    const res = await fetch(`/api/files/list?path=${encodeURIComponent(path)}`)
    const data = await res.json()
    const prefix = path ? (path.endsWith('/') ? path : path + '/') : ''
    files.value = data.filter(f => {
      if (f.FullPath === path) return false
      const rel = f.FullPath.startsWith(prefix) ? f.FullPath.slice(prefix.length) : f.FullPath
      return rel && !rel.includes('/')
    }).sort((a, b) => (b.IsDir - a.IsDir) || a.Name.localeCompare(b.Name))
    currentPath.value = path
    localStorage.setItem('fm-last-path', path)
  } catch (e) { console.error(e) } finally { isLoading.value = false }
}

const fetchFolderTree = async () => {
  try {
    const res = await fetch('/api/files/list?path=')
    const data = await res.json()
    const folders = data.filter(f => f.IsDir).sort((a, b) => a.FullPath.localeCompare(b.FullPath))
    const build = (list, p = '') => list.filter(f => {
      const parts = f.FullPath.split('/')
      return parts.slice(0, -1).join('/') === p
    }).map(f => ({ ...f, children: build(list, f.FullPath), isOpen: false }))
    foldersTree.value = [{ Name: '根目录', FullPath: '', children: build(folders, ''), isOpen: true }]
  } catch (e) { console.error(e) }
}

const navigateTo = (p) => { fetchFiles(p); isMobileSidebarOpen.value = false }

const breadcrumbs = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  let acc = ''
  return [{ name: '首页', path: '' }, ...parts.map(p => ({ name: p, path: acc = acc ? `${acc}/${p}` : p }))]
})

const filteredFiles = computed(() => {
  if (!searchQuery.value) return files.value
  return files.value.filter(f => f.Name.toLowerCase().includes(searchQuery.value.toLowerCase()))
})

onMounted(() => { fetchFiles(currentPath.value); fetchFolderTree() })

// --- 格式化 ---
const formatSize = (b) => b ? (b / Math.pow(1024, Math.floor(Math.log(b) / Math.log(1024)))).toFixed(1) + ' ' + ['B', 'KB', 'MB', 'GB', 'TB'][Math.floor(Math.log(b) / Math.log(1024))] : '—'
</script>

<template>
  <div class="fixed inset-0 flex flex-col bg-gray-50 dark:bg-[#0d1117] text-gray-900 dark:text-gray-100 transition-colors duration-200 font-sans antialiased">
    <!-- Navbar -->
    <header class="h-16 border-b border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-[#161b22]/80 backdrop-blur-md flex items-center justify-between px-6 z-50">
      <div class="flex items-center gap-4">
        <button @click="isMobileSidebarOpen = true" class="lg:hidden p-2 text-gray-500"><i class="fas fa-bars text-xl"></i></button>
        <div class="flex items-center gap-2 cursor-pointer" @click="navigateTo('')">
          <div class="w-10 h-10 bg-blue-600 rounded-xl flex items-center justify-center text-white shadow-lg shadow-blue-500/30">
            <i class="fas fa-cloud-download-alt text-xl"></i>
          </div>
          <span class="text-2xl font-black tracking-tighter hidden sm:block">ModernFM</span>
        </div>
      </div>

      <nav class="hidden md:flex items-center bg-gray-100 dark:bg-gray-900 px-6 py-2 rounded-full border border-gray-200 dark:border-gray-800 flex-1 mx-10 overflow-hidden">
        <template v-for="(bc, i) in breadcrumbs" :key="bc.path">
          <i v-if="i > 0" class="fas fa-chevron-right text-[10px] mx-3 opacity-30 text-gray-400"></i>
          <button @click="navigateTo(bc.path)" class="text-sm hover:text-blue-500 font-medium truncate" :class="{'text-blue-600 font-bold': i === breadcrumbs.length - 1}">{{ bc.name }}</button>
        </template>
      </nav>

      <div class="flex items-center gap-4">
        <div class="relative hidden sm:block">
          <input v-model="searchQuery" type="text" placeholder="搜索文件..." class="w-64 bg-gray-100 dark:bg-gray-900 border-none rounded-full py-2 pl-10 pr-4 text-sm focus:ring-2 focus:ring-blue-500/20 transition-all">
          <i class="fas fa-search absolute left-4 top-1/2 -translate-y-1/2 text-gray-400"></i>
        </div>
        <button @click="toggleTheme" class="w-10 h-10 rounded-full flex items-center justify-center bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors">
          <i class="fas" :class="theme === 'light' ? 'fa-moon' : 'fa-sun'"></i>
        </button>
      </div>
    </header>

    <div class="flex flex-1 overflow-hidden relative">
      <!-- Sidebar -->
      <aside :class="{'translate-x-0': isMobileSidebarOpen, '-translate-x-full lg:translate-x-0': !isMobileSidebarOpen}" 
             class="fixed lg:relative inset-y-0 left-0 w-80 bg-white dark:bg-[#161b22] border-r border-gray-200 dark:border-gray-800 z-[60] lg:z-0 transition-transform duration-300 flex flex-col pt-16 lg:pt-0">
        <div class="p-6 border-b border-gray-100 dark:border-gray-800 lg:hidden flex justify-between items-center">
          <span class="font-black text-xl">导航菜单</span>
          <button @click="isMobileSidebarOpen = false"><i class="fas fa-times text-xl text-gray-400"></i></button>
        </div>
        <div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
          <div v-for="node in foldersTree" :key="node.FullPath" class="tree-root">
            <div @click="node.isOpen = !node.isOpen" class="flex items-center gap-2 p-2.5 rounded-xl cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 transition-all" :class="{'bg-blue-50 dark:bg-blue-900/20 text-blue-600 font-bold': currentPath === node.FullPath}">
              <i class="fas fa-caret-right transition-transform" :class="{'rotate-90': node.isOpen, 'opacity-0': !node.children.length}"></i>
              <i class="fas" :class="node.FullPath === '' ? 'fa-hdd' : (node.isOpen ? 'fa-folder-open text-amber-500' : 'fa-folder text-amber-500')"></i>
              <span @click.stop="navigateTo(node.FullPath)" class="truncate">{{ node.Name }}</span>
            </div>
            <div v-if="node.isOpen" class="ml-5 border-l border-gray-100 dark:border-gray-800 pl-2 mt-1 space-y-1">
              <div v-for="child in node.children" :key="child.FullPath" @click="child.isOpen = !child.isOpen" class="flex items-center gap-2 p-2 rounded-lg cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 transition-all text-sm" :class="{'text-blue-600 font-bold': currentPath === child.FullPath}">
                <i class="fas fa-caret-right transition-transform" :class="{'rotate-90': child.isOpen, 'opacity-0': !child.children.length}"></i>
                <i class="fas fa-folder text-amber-500 opacity-80"></i>
                <span @click.stop="navigateTo(child.FullPath)" class="truncate">{{ child.Name }}</span>
              </div>
            </div>
          </div>
        </div>
      </aside>

      <!-- Main Panel -->
      <main class="flex-1 flex flex-col bg-white dark:bg-[#0d1117] min-w-0 overflow-hidden relative">
        <div class="h-14 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between px-6 bg-gray-50/30 dark:bg-gray-800/10">
          <div class="flex items-center gap-4">
            <button @click="navigateTo(currentPath.split('/').slice(0,-1).join('/'))" :disabled="!currentPath" class="w-9 h-9 flex items-center justify-center hover:bg-gray-200 dark:hover:bg-gray-800 rounded-full disabled:opacity-20 transition-all"><i class="fas fa-arrow-left"></i></button>
            <span class="text-[11px] font-black text-gray-400 uppercase tracking-widest">{{ filteredFiles.length }} 项资源</span>
          </div>
          <div class="flex items-center gap-2">
            <button class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-xs font-bold shadow-lg shadow-blue-500/20 transition-all">上传文件</button>
            <button class="p-2.5 text-gray-400 hover:text-blue-500"><i class="fas fa-sync-alt"></i></button>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto custom-scrollbar relative">
          <div v-if="isLoading" class="absolute inset-0 bg-white/60 dark:bg-black/60 z-30 flex items-center justify-center backdrop-blur-sm"><i class="fas fa-circle-notch fa-spin text-4xl text-blue-500"></i></div>
          
          <table class="w-full text-left">
            <thead class="sticky top-0 bg-white/95 dark:bg-[#0d1117]/95 backdrop-blur-md border-b border-gray-200 dark:border-gray-800 z-20">
              <tr class="text-[11px] font-black text-gray-400 uppercase tracking-wider">
                <th class="px-6 py-4 w-12 text-center"><input type="checkbox" class="rounded"></th>
                <th class="px-6 py-4">文件名</th>
                <th class="px-6 py-4 hidden sm:table-cell w-32 text-right">大小</th>
                <th class="px-6 py-4 hidden md:table-cell w-40 text-right">修改日期</th>
                <th class="px-6 py-4 w-16"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-50 dark:divide-gray-800/50">
              <tr v-for="file in filteredFiles" :key="file.FullPath" @dblclick="file.IsDir ? navigateTo(file.FullPath) : null" class="group hover:bg-blue-50/40 dark:hover:bg-blue-900/10 transition-all cursor-pointer">
                <td class="px-6 py-4 text-center"><input type="checkbox" class="rounded accent-blue-600"></td>
                <td class="px-6 py-4">
                  <div class="flex items-center gap-4">
                    <div class="w-11 h-11 flex items-center justify-center rounded-2xl shadow-sm transition-transform group-hover:scale-110" :class="file.IsDir ? 'bg-amber-100 dark:bg-amber-900/20 text-amber-500' : 'bg-blue-50 dark:bg-blue-900/20 text-blue-500'">
                      <i class="fas text-xl" :class="file.IsDir ? 'fa-folder' : 'fa-file-alt'"></i>
                    </div>
                    <div class="flex flex-col min-w-0">
                      <span class="text-sm font-bold truncate group-hover:text-blue-600 transition-colors">{{ file.Name }}</span>
                      <span class="text-[10px] text-gray-400 sm:hidden">{{ formatSize(file.Size) }} • {{ formatDate(file.ModTime) }}</span>
                    </div>
                  </div>
                </td>
                <td class="px-6 py-4 hidden sm:table-cell text-right text-xs font-mono text-gray-500">{{ formatSize(file.Size) }}</td>
                <td class="px-6 py-4 hidden md:table-cell text-right text-xs text-gray-400">{{ formatDate(file.ModTime) }}</td>
                <td class="px-6 py-4 text-right"><button class="p-2 opacity-0 group-hover:opacity-100 text-gray-300 hover:text-blue-500 transition-all"><i class="fas fa-ellipsis-v"></i></button></td>
              </tr>
            </tbody>
          </table>

          <div v-if="!isLoading && !filteredFiles.length" class="flex flex-col items-center py-48 opacity-20"><i class="fas fa-box-open text-8xl mb-6"></i><p class="text-2xl font-black">空空如也</p></div>
        </div>
      </main>

      <div v-if="isMobileSidebarOpen" @click="isMobileSidebarOpen = false" class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[55] lg:hidden"></div>
    </div>
  </div>
</template>

<style>
@import url('https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css');
.custom-scrollbar::-webkit-scrollbar { width: 5px; height: 5px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 10px; }
.dark .custom-scrollbar::-webkit-scrollbar-thumb { background: #374151; }
</style>
