<script setup>
import { onMounted } from 'vue'
import ThemedMainView from './views/ThemedMainView.vue'

onMounted(() => {
  // 确保 body 撑满全屏且不滚动
  document.body.style.margin = '0'
  document.body.style.height = '100vh'
  document.body.style.overflow = 'hidden'
})
</script>

<template>
  <ThemedMainView />
</template>

<style>
/* 全局样式覆盖 */
:root {
  color-scheme: light dark;
}

html, body, #app {
  height: 100%;
  width: 100%;
}

/* 适配移动端的点击反馈 */
* {
  -webkit-tap-highlight-color: transparent;
}
</style>
