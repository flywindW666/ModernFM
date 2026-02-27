<script setup>
import { ChevronRight, ChevronDown, Folder, FolderOpen, Home } from 'lucide-vue-next';

const props = defineProps({
  node: Object,
  currentPath: String,
  navigateTo: Function,
  toggleFolder: Function
});
</script>

<template>
  <div class="tree-node select-none w-full">
    <div 
      class="flex items-center gap-2 px-2 py-1 rounded-lg cursor-pointer transition-colors w-full overflow-hidden"
      :class="[
        currentPath === node.FullPath ? 'bg-blue-500/10 text-blue-500 font-medium' : 'hover:bg-slate-100 dark:hover:bg-slate-800/50 text-slate-600 dark:text-slate-400'
      ]"
      @click="toggleFolder(node)"
    >
      <!-- 展开/收起图标 -->
      <div class="w-4 h-4 flex items-center justify-center shrink-0">
        <template v-if="node.IsDir">
          <component 
            :is="node.isOpen ? ChevronDown : ChevronRight" 
            class="w-3.5 h-3.5"
            :class="{'animate-spin opacity-50': node.loading}"
          />
        </template>
      </div>

      <!-- 文件夹图标 -->
      <component 
        :is="node.FullPath === '' ? Home : (node.isOpen ? FolderOpen : Folder)" 
        class="w-4 h-4 shrink-0"
        :class="node.FullPath === '' ? 'text-blue-500' : 'text-amber-500/80'"
      />

      <!-- 名称 (点击跳转) -->
      <span 
        class="truncate text-[13px] flex-1 min-w-0"
        @click.stop="navigateTo(node.FullPath)"
      >
        {{ node.Name }}
      </span>
    </div>

    <!-- 子目录递归渲染 -->
    <div v-if="node.isOpen && node.children && node.children.length" class="ml-4 border-l border-slate-200 dark:border-slate-800 pl-1 mt-0.5">
      <RecursiveTree 
        v-for="child in node.children" 
        :key="child.FullPath" 
        :node="child" 
        :currentPath="currentPath"
        :navigateTo="navigateTo"
        :toggleFolder="toggleFolder"
      />
    </div>
  </div>
</template>

<script>
// 生产环境必须显式定义组件名以支持递归
export default {
  name: 'RecursiveTree'
}
</script>
