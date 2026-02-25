<script setup>
import { ref, onMounted, watchEffect } from 'vue'

// 主题状态：'light', 'dark', 'auto'
const themeMode = ref(localStorage.getItem('fm-theme') || 'auto')

const applyTheme = (mode) => {
  const root = document.documentElement
  if (mode === 'dark' || (mode === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    root.classList.add('dark')
  } else {
    root.classList.remove('dark')
  }
}

watchEffect(() => {
  localStorage.setItem('fm-theme', themeMode.value)
  applyTheme(themeMode.value)
})

// 监听系统主题变化
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
  if (themeMode.value === 'auto') applyTheme('auto')
})

const toggleTheme = (mode) => {
  themeMode.value = mode
}
</script>

<template>
  <div class="app-container min-h-screen transition-colors duration-300 bg-white dark:bg-[#121212] text-slate-900 dark:text-slate-200">
    <!-- 主体布局参考 Alist + 肥牛 -->
    <div class="flex h-screen overflow-hidden">
      
      <!-- 左侧边栏：融合 Alist 目录树与肥牛侧栏 -->
      <aside class="w-64 border-r border-slate-200 dark:border-slate-800 bg-slate-50 dark:bg-[#1a1a1a] flex flex-col">
        <div class="p-4 flex items-center justify-between border-b border-slate-200 dark:border-slate-800">
          <h1 class="text-xl font-bold text-blue-600 dark:text-blue-400 italic">ModernFM</h1>
          <!-- 主题切换器 -->
          <div class="flex bg-slate-200 dark:bg-slate-700 rounded-full p-1 text-[10px]">
            <button @click="toggleTheme('light')" :class="{'bg-white dark:bg-slate-500 shadow-sm': themeMode === 'light'}" class="px-2 py-1 rounded-full"><i class="fas fa-sun"></i></button>
            <button @click="toggleTheme('dark')" :class="{'bg-white dark:bg-slate-500 shadow-sm': themeMode === 'dark'}" class="px-2 py-1 rounded-full ml-1"><i class="fas fa-moon"></i></button>
            <button @click="toggleTheme('auto')" :class="{'bg-white dark:bg-slate-500 shadow-sm': themeMode === 'auto'}" class="px-2 py-1 rounded-full ml-1">Auto</button>
          </div>
        </div>

        <!-- 目录树内容 -->
        <div class="flex-1 overflow-y-auto p-2">
           <div class="tree-item active bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 p-2 rounded-lg mb-1 flex items-center cursor-pointer">
             <i class="fas fa-folder-open mr-2"></i> <span>Root</span>
           </div>
           <div class="tree-item p-2 hover:bg-slate-200 dark:hover:bg-slate-800 rounded-lg flex items-center cursor-pointer transition-colors" v-for="i in 5" :key="i">
             <i class="fas fa-chevron-right text-[10px] mr-2 opacity-50"></i>
             <i class="fas fa-folder text-yellow-500 mr-2"></i>
             <span class="text-sm">Shared Data {{i}}</span>
           </div>
        </div>
      </aside>

      <!-- 右侧主内容 -->
      <main class="flex-1 flex flex-col bg-white dark:bg-[#121212]">
        <!-- 顶部工具栏：复刻 Alist 风格 -->
        <header class="h-14 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between px-4">
          <div class="flex items-center space-x-2 text-sm">
             <i class="fas fa-home opacity-50"></i>
             <span class="opacity-30">/</span>
             <span class="font-medium">Backup</span>
          </div>
          <div class="flex items-center space-x-4">
            <div class="flex space-x-2">
               <button class="px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-xs hover:bg-slate-100 dark:hover:bg-slate-800 transition-all"><i class="fas fa-plus mr-1"></i> 添加</button>
               <button class="px-3 py-1.5 rounded-lg bg-blue-600 text-white text-xs hover:bg-blue-700 shadow-lg shadow-blue-500/20"><i class="fas fa-play mr-1"></i> Play Folder</button>
            </div>
            <div class="w-px h-6 bg-slate-200 dark:bg-slate-800"></div>
            <div class="flex items-center space-x-3 text-slate-400">
               <i class="fas fa-th-large cursor-pointer hover:text-blue-500"></i>
               <i class="fas fa-list cursor-pointer text-blue-500"></i>
               <i class="fas fa-cog cursor-pointer hover:text-blue-500"></i>
            </div>
          </div>
        </header>

        <!-- 文件展示区：高密度列表模式 -->
        <div class="flex-1 overflow-y-auto">
          <table class="w-full border-collapse">
            <thead class="sticky top-0 bg-slate-50 dark:bg-[#181818] border-b border-slate-200 dark:border-slate-800 text-[11px] font-semibold text-slate-500 dark:text-slate-400 uppercase">
              <tr>
                <th class="p-3 w-10 text-center"><input type="checkbox" class="accent-blue-500"></th>
                <th class="p-3 text-left">名称</th>
                <th class="p-3 text-left w-32">大小</th>
                <th class="p-3 text-left w-32">类型</th>
                <th class="p-3 text-left w-48">修改时间</th>
              </tr>
            </thead>
            <tbody class="text-sm">
              <tr v-for="file in mockFiles" :key="file.name" class="group border-b border-slate-100 dark:border-slate-800/50 hover:bg-blue-50 dark:hover:bg-blue-900/10 transition-colors">
                <td class="p-3 text-center"><input type="checkbox" class="accent-blue-500"></td>
                <td class="p-3 flex items-center space-x-3">
                   <div class="text-2xl w-8 h-8 flex items-center justify-center">
                     <i v-if="file.isDir" class="fas fa-folder text-yellow-500"></i>
                     <i v-else-if="file.type === 'video'" class="fas fa-file-video text-purple-500"></i>
                     <i v-else class="fas fa-file text-slate-400"></i>
                   </div>
                   <span class="font-medium truncate max-w-md">{{ file.name }}</span>
                </td>
                <td class="p-3 text-xs text-slate-500 font-mono">{{ file.isDir ? '—' : file.size }}</td>
                <td class="p-3 text-xs text-slate-400">{{ file.type }}</td>
                <td class="p-3 text-xs text-slate-400 font-mono">{{ file.modTime }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </main>
    </div>
  </div>
</template>

<style>
/* 注入 Tailwind 不支持的自定义滚动条 */
::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 10px; }
.dark ::-webkit-scrollbar-thumb { background: #334155; }
::-webkit-scrollbar-track { background: transparent; }

.tree-item { font-size: 13px; font-weight: 500; }
</style>
