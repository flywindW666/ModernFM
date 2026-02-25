<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const props = defineProps(['targetFile'])
const emit = defineEmits(['action'])

const visible = ref(false)
const position = ref({ x: 0, y: 0 })

const show = (e) => {
  e.preventDefault()
  position.value = { x: e.clientX, y: e.clientY }
  visible.value = true
}

const hide = () => {
  visible.value = false
}

const handleAction = (type) => {
  emit('action', { type, file: props.targetFile })
  hide()
}

onMounted(() => {
  window.addEventListener('click', hide)
})

onUnmounted(() => {
  window.removeEventListener('click', hide)
})

defineExpose({ show })
</script>

<template>
  <div v-if="visible" 
       :style="{ top: position.y + 'px', left: position.x + 'px' }"
       class="fixed z-[3000] w-48 bg-white dark:bg-zinc-800 shadow-2xl border border-slate-200 dark:border-zinc-700 rounded-xl py-2 text-sm overflow-hidden animate-in fade-in zoom-in duration-100">
    
    <div @click="handleAction('open')" class="item"><i class="fas fa-folder-open mr-3 text-blue-500"></i> 打开</div>
    <div @click="handleAction('download')" class="item"><i class="fas fa-download mr-3 text-green-500"></i> 下载</div>
    
    <div class="divider"></div>
    
    <div @click="handleAction('share')" class="item"><i class="fas fa-share-alt mr-3 text-indigo-500"></i> 生成分享链接</div>
    <div @click="handleAction('rename')" class="item"><i class="fas fa-edit mr-3 text-amber-500"></i> 重命名</div>
    
    <div class="divider"></div>
    
    <!-- 压缩与解压菜单项 -->
    <div @click="handleAction('compress')" class="item"><i class="fas fa-file-archive mr-3 text-orange-500"></i> 压缩为 ZIP</div>
    <div v-if="targetFile.name.endsWith('.zip')" @click="handleAction('extract')" class="item"><i class="fas fa-box-open mr-3 text-cyan-500"></i> 解压到当前目录</div>
    
    <div class="divider"></div>
    
    <div @click="handleAction('delete')" class="item text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20"><i class="fas fa-trash-alt mr-3"></i> 删除</div>
  </div>
</template>

<style scoped>
.item { @apply px-4 py-2 flex items-center cursor-pointer transition-colors hover:bg-slate-100 dark:hover:bg-zinc-700 text-slate-700 dark:text-slate-300; }
.divider { @apply h-px bg-slate-100 dark:bg-zinc-700 my-1; }
</style>
