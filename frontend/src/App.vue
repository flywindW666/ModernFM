<script setup>
import { ref, onMounted, watchEffect } from 'vue'
import ThemedMainView from './views/ThemedMainView.vue'

// 全局主题管理逻辑
const theme = ref(localStorage.getItem('fm-theme') || 'auto')

const updateTheme = () => {
  const isDark = theme.value === 'dark' || 
    (theme.value === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)
  
  if (isDark) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

watchEffect(() => {
  localStorage.setItem('fm-theme', theme.value)
  updateTheme()
})

onMounted(() => {
  updateTheme()
  // 监听系统主题变化
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', updateTheme)
})
</script>

<template>
  <div class="modern-fm-app">
    <!-- 注入主题切换控制的全局视图 -->
    <ThemedMainView />
  </div>
</template>

<style>
/* 全局基础样式 */
html, body {
  margin: 0;
  padding: 0;
  height: 100%;
}

/* 适配深色模式的滚动条 */
.dark ::-webkit-scrollbar-thumb {
  background-color: #3f3f46;
}
.dark ::-webkit-scrollbar-track {
  background-color: #18181b;
}
</style>
