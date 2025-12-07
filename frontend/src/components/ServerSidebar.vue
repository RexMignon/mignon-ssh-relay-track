<template>
  <el-aside :width="asideWidth" class="aside-panel" :class="{ 'is-collapsed': isCollapsed }">
    <div v-if="!isCollapsed" class="resize-handle" @mousedown="startResize"></div>

    <div class="aside-header" style="--wails-draggable:drag">
      <div class="brand-container">
        <img v-if="!isCollapsed" :src="Logo" alt="logo" class="app-logo" />
        <span v-if="!isCollapsed" class="app-title">{{ t.app.title }}</span>
      </div>

      <div class="header-actions" style="--wails-draggable:no-drag">
        <el-button
            v-if="!isCollapsed"
            type="primary"
            size="small"
            :icon="Plus"
            circle
            @click="$emit('add-server')"
            :title="t.sidebar.add"
        />
        <el-button
            link
            class="collapse-btn"
            @click="toggleCollapse"
            :title="isCollapsed ? t.sidebar.expand : t.sidebar.collapse"
        >
          <el-icon><component :is="isCollapsed ? Expand : Fold" /></el-icon>
        </el-button>
      </div>
    </div>

    <div v-if="!isCollapsed" class="search-wrapper">
      <el-input
          v-model="searchKeyword"
          :placeholder="t.sidebar.search"
          :prefix-icon="Search"
          clearable
          size="small"
      />
    </div>

    <el-scrollbar class="menu-scrollbar">
      <el-menu
          :default-active="currentServerId"
          class="server-menu"
          :collapse="isCollapsed"
          :collapse-transition="false"
          @select="(idx) => $emit('select-server', idx)"
      >
        <el-menu-item v-for="server in filteredServerList" :key="server.id" :index="server.id">
          <el-icon><Monitor /></el-icon>
          <template #title>
            <span class="server-name-text" :title="server.server_name">{{ server.server_name }}</span>
            <div class="server-controls" @click.stop>
              <el-tooltip :content="t.sidebar.toggleGroup" placement="top" :enterable="false">
                <el-switch
                    :model-value="server.is_open"
                    size="small"
                    :loading="server.loading"
                    @change="(val) => $emit('toggle-server', server, val)"
                    style="margin-right: 8px;"
                />
              </el-tooltip>
              <el-button
                  type="primary"
                  link
                  size="small"
                  :icon="Edit"
                  class="action-btn"
                  @click="$emit('edit-server', server)"
                  :title="t.sidebar.edit"
              />
              <el-popconfirm :title="t.sidebar.deleteConfirm" @confirm="$emit('delete-server', server.id)">
                <template #reference>
                  <el-button
                      type="danger"
                      link
                      size="small"
                      :icon="Delete"
                      class="action-btn"
                      :title="t.sidebar.delete"
                  />
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-menu-item>

        <div v-if="filteredServerList.length === 0 && !isCollapsed" class="empty-search">
          {{ t.sidebar.empty }}
        </div>
      </el-menu>
    </el-scrollbar>

    <!-- Footer -->
    <div class="aside-footer">
      <div v-if="!isCollapsed" class="footer-content">
        <div class="theme-switch-wrapper">
          <el-switch
              :model-value="isDark"
              inline-prompt
              :active-icon="Moon"
              :inactive-icon="Sunny"
              @change="(val) => $emit('toggle-theme', val)"
          />
          <span class="footer-text">{{ t.app.theme }}</span>
        </div>

        <!-- 语言切换: 显示 En/文 -->
        <div class="lang-switch-group" @click="$emit('toggle-language')" :title="t.app.switchLang">
          <!-- En -->
          <span :class="['lang-text', language === 'en' ? 'active' : 'inactive']">En</span>
          <span class="lang-divider">/</span>
          <!-- 文 -->
          <span :class="['lang-text', language === 'zh' ? 'active' : 'inactive']">文</span>

          <el-icon class="lang-icon"><Switch /></el-icon>
        </div>
      </div>

      <div v-else class="footer-mini">
        <el-icon class="clickable-icon" @click="$emit('toggle-theme', !isDark)" :title="t.app.theme">
          <component :is="isDark ? Moon : Sunny" />
        </el-icon>
      </div>
    </div>
  </el-aside>
</template>

<script lang="ts" setup>
import { ref, computed, onUnmounted, inject } from 'vue'
import { Monitor, Plus, Delete, Edit, Moon, Sunny, Fold, Expand, Search, Switch } from '@element-plus/icons-vue'
import Logo from '../assets/images/logo-universal.png'

const props = defineProps({
  serverList: { type: Array as () => any[], default: () => [] },
  currentServerId: { type: String, default: '' },
  isDark: { type: Boolean, default: false },
  language: { type: String, default: 'zh' }
})

defineEmits(['select-server', 'add-server', 'edit-server', 'delete-server', 'toggle-theme', 'toggle-server', 'toggle-language'])

// 注入翻译对象
const t: any = inject('t')

const isCollapsed = ref(false)
const sidebarWidth = ref(280)
const searchKeyword = ref('')

const asideWidth = computed(() => isCollapsed.value ? '64px' : `${sidebarWidth.value}px`)

const filteredServerList = computed(() => {
  if (!searchKeyword.value) return props.serverList
  const keyword = searchKeyword.value.toLowerCase()
  return props.serverList.filter(server => {
    if (server.server_name && server.server_name.toLowerCase().includes(keyword)) return true
    if (server.notes && server.notes.toLowerCase().includes(keyword)) return true
    if (server.link_group && Array.isArray(server.link_group)) {
      if (server.link_group.some((link: any) =>
          (link.name && link.name.toLowerCase().includes(keyword)) ||
          (link.notes && link.notes.toLowerCase().includes(keyword))
      )) return true
    }
    return false
  })
})

const toggleCollapse = () => isCollapsed.value = !isCollapsed.value

const MIN_WIDTH = 200
const MAX_WIDTH = 500
const startResize = (e: MouseEvent) => {
  e.preventDefault()
  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}
const onMouseMove = (e: MouseEvent) => {
  let newWidth = e.clientX
  if (newWidth < MIN_WIDTH) newWidth = MIN_WIDTH
  if (newWidth > MAX_WIDTH) newWidth = MAX_WIDTH
  sidebarWidth.value = newWidth
}
const onMouseUp = () => {
  document.removeEventListener('mousemove', onMouseMove)
  document.removeEventListener('mouseup', onMouseUp)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
}
onUnmounted(() => {
  document.removeEventListener('mousemove', onMouseMove)
  document.removeEventListener('mouseup', onMouseUp)
})
</script>

<style scoped>
.aside-panel {
  background-color: var(--el-bg-color-overlay);
  border-right: 1px solid var(--el-border-color-light);
  display: flex;
  flex-direction: column;
  height: 100%;
  position: relative;
  transition: width 0.1s ease-out;
  overflow: hidden;
}

.resize-handle {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 5px;
  cursor: col-resize;
  z-index: 100;
  background-color: transparent;
  transition: background-color 0.2s;
}
.resize-handle:hover,
.resize-handle:active {
  background-color: var(--el-color-primary-light-8);
}

.aside-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
  flex-shrink: 0;
  cursor: default;
}
.aside-panel.is-collapsed .aside-header {
  justify-content: center;
  padding: 0;
}

.brand-container {
  display: flex;
  align-items: center;
  gap: 10px;
  overflow: hidden;
}
.app-logo {
  width: 28px;
  height: 28px;
  object-fit: contain;
  flex-shrink: 0;
}
.app-title {
  font-weight: bold;
  font-size: 16px;
  color: var(--el-text-color-primary);
  white-space: nowrap;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}
.collapse-btn {
  font-size: 18px;
  padding: 4px;
  margin-left: 4px;
  color: var(--el-text-color-regular);
}
.aside-panel.is-collapsed .collapse-btn { margin-left: 0; }
.collapse-btn:hover { color: var(--el-color-primary); }

.search-wrapper {
  padding: 10px 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.menu-scrollbar { flex: 1; }
.server-menu { border-right: none; background-color: transparent; }
:deep(.el-menu-item) { display: flex; align-items: center; padding-right: 12px; }
.server-menu:not(.el-menu--collapse) { width: 100%; }
.server-name-text {
  margin-left: 8px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.server-controls {
  display: flex;
  align-items: center;
  opacity: 0;
  transition: opacity 0.2s;
  margin-left: auto;
}
.el-menu-item:hover .server-controls { opacity: 1; }
.el-menu--collapse .server-controls { display: none; }
.action-btn { padding: 4px; margin-left: 2px; font-size: 14px; }
.empty-search {
  text-align: center;
  color: var(--el-text-color-placeholder);
  font-size: 12px;
  padding: 20px 0;
}

.aside-footer {
  height: 50px;
  border-top: 1px solid var(--el-border-color-light);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  padding: 0 10px;
}

.footer-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}
.theme-switch-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
}
.footer-text { font-size: 12px; color: var(--el-text-color-secondary); }

/* 语言切换样式 */
.lang-switch-group {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: background-color 0.2s;
}
.lang-switch-group:hover {
  background-color: var(--el-fill-color-light);
}

.lang-text {
  font-size: 12px;
  font-weight: bold;
  transition: color 0.3s;
}

/* 激活状态：正常颜色 */
.lang-text.active {
  color: var(--el-text-color-primary);
}

/* 非激活状态：灰色 */
.lang-text.inactive {
  color: var(--el-text-color-placeholder);
}

.lang-divider {
  margin: 0 4px;
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

.lang-icon {
  margin-left: 8px;
  font-size: 16px;
  color: var(--el-text-color-regular);
}
.lang-switch-group:hover .lang-icon {
  color: var(--el-color-primary);
}

/* 折叠态底部 */
.footer-mini {
  display: flex;
  align-items: center;
  justify-content: space-around;
  width: 100%;
  gap: 4px;
}
.clickable-icon {
  cursor: pointer;
  font-size: 18px;
  color: var(--el-text-color-regular);
  padding: 4px;
}
.clickable-icon:hover { color: var(--el-color-primary); }
</style>