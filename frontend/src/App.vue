<template>
  <el-container class="layout-container">
    <!-- 左侧侧边栏 -->
    <ServerSidebar
        :server-list="config.config"
        :current-server-id="currentServerId"
        :is-dark="config.is_dark"
        :language="language"
        @select-server="(id) => currentServerId = id"
        @toggle-theme="handleThemeToggle"
        @add-server="openServerDialog(null)"
        @edit-server="openServerDialog"
        @delete-server="handleDeleteServer"
        @toggle-server="handleToggleServer"
        @toggle-language="handleLanguageToggle"
    />

    <!-- 右侧内容区 -->
    <TunnelList
        :current-server="currentServer"
        :active-tunnel-ids="activeTunnelIds"
        :loading-state="loadingState"
        @refresh="refreshData"
        @add-link="openLinkDialog(null)"
        @edit-link="openLinkDialog"
        @delete-link="handleDeleteLink"
        @toggle-link="handleToggleLink"
        @copy-link="handleCopyLink"
    />

    <!-- 弹窗组件 -->
    <ServerDialog
        v-model:visible="serverDialog.visible"
        :is-edit="serverDialog.isEdit"
        :initial-data="serverDialog.data"
        @save="onServerSave"
    />

    <LinkDialog
        v-model:visible="linkDialog.visible"
        :is-edit="linkDialog.isEdit"
        :initial-data="linkDialog.data"
        @save="onLinkSave"
    />

  </el-container>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, nextTick, provide } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'

// 子组件
import ServerSidebar from './components/ServerSidebar.vue'
import TunnelList from './components/TunnelList.vue'
import ServerDialog from './components/ServerDialog.vue'
import LinkDialog from './components/LinkDialog.vue'

// i18n
import { i18n } from './i18n'

// Wails Imports
import {
  GetConfig, GetActiveTunnelIds, AddServer, ModifyServer, ModifyServers, DeleteServer,
  AddLink, ModifyLink, DeleteLink, ToggleLinkStatus, ThemeSwitch, SetLanguage
} from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

interface Link {
  id: string;
  name: string;
  local_host: string;
  local_port: number;
  remote_host: string;
  remote_port: number;
  is_penetrate: boolean;
  is_open: boolean;
  notes: string;
}
interface ServerConfig {
  id: string;
  server_name: string;
  server_host: string;
  server_port: number;
  username: string;
  link_group: Link[];
}
interface ConfigState {
  config: ServerConfig[];
  is_dark: boolean;
  is_english: boolean; // 新增：从配置中读取是否为英文
}

// === State ===
const config = ref<ConfigState>({ config: [], is_dark: false, is_english: false })
const activeTunnelIds = ref<string[]>([])
const currentServerId = ref<string>('')
const loadingState = reactive<Record<string, boolean>>({})
const language = ref<string>('zh')

// === i18n Provider ===
// 计算当前语言的字典
const t = computed(() => {
  return language.value === 'zh' ? i18n.zh : i18n.en
})
// 向所有子组件注入 't'
provide('t', t)

// Dialog States
const serverDialog = reactive({ visible: false, isEdit: false, data: null })
const linkDialog = reactive({ visible: false, isEdit: false, data: null })

// === Computed ===
const currentServer = computed(() => {
  return config.value.config.find(s => s.id === currentServerId.value)
})

// === Lifecycle ===
onMounted(async () => {
  await refreshData()

  EventsOn("tunnel_event", (event: { Error: string, LinkName: string, ID: string, IsStopped: boolean }) => {
    if (event.Error) {
      ElNotification({
        title: 'Tunnel Error',
        message: `[${event.LinkName}] ${event.Error}`,
        type: 'error',
        duration: 5000
      })
      removeActiveId(event.ID)
    } else {
      if (event.IsStopped) {
        removeActiveId(event.ID)
      } else {
        addActiveId(event.ID)
      }
    }
    refreshActiveIds()
  })

  if (config.value.config.length > 0) {
    currentServerId.value = config.value.config[0].id
  }
  // 此时 config 已经加载，主题和语言都已应用
})

// === Data Helpers ===
const refreshData = async () => {
  try {
    const cfg = await GetConfig()
    config.value = { ...cfg } as ConfigState

    // 同步语言状态
    if (config.value.is_english) {
      language.value = 'en'
    } else {
      language.value = 'zh'
    }

    // 同步主题状态
    applyTheme(config.value.is_dark)

    await refreshActiveIds()
  } catch (err) {
    ElMessage.error("Config Load Failed: " + err)
  }
}

const refreshActiveIds = async () => {
  const ids = await GetActiveTunnelIds()
  activeTunnelIds.value = ids || []
}

const addActiveId = (key: string) => { if (!activeTunnelIds.value.includes(key)) activeTunnelIds.value.push(key) }
const removeActiveId = (key: string) => { activeTunnelIds.value = activeTunnelIds.value.filter(k => k !== key) }

// === Theme ===
const handleThemeToggle = async (isDark: boolean) => {
  config.value.is_dark = isDark
  applyTheme(isDark)
  try {
    await ThemeSwitch(isDark)
  } catch (e) {
    console.error("Theme switch sync failed:", e)
  }
}

const applyTheme = (isDark: boolean) => {
  const html = document.documentElement
  if (isDark) html.classList.add('dark')
  else html.classList.remove('dark')
}

// === Language ===
const handleLanguageToggle = async () => {
  // 如果当前是 zh，点击后变成 en (is_english = true)
  // 如果当前是 en，点击后变成 zh (is_english = false)
  const nextIsEnglish = language.value === 'zh'

  try {
    // 调用后端接口
    await SetLanguage(nextIsEnglish)

    // 更新本地状态
    language.value = nextIsEnglish ? 'en' : 'zh'
    config.value.is_english = nextIsEnglish
  } catch (e) {
    ElMessage.error("Language Switch Failed: " + e)
  }
}

// === Server Actions ===
const openServerDialog = (data: any = null) => {
  serverDialog.data = data
  serverDialog.isEdit = !!data
  serverDialog.visible = true
}

const onServerSave = async (payload: any) => {
  const finalPayload = { ...payload }
  if (!serverDialog.isEdit) finalPayload.id = generateUUID()

  try {
    if (serverDialog.isEdit) {
      await ModifyServer(finalPayload.id, finalPayload)
    } else {
      await AddServer(finalPayload)
    }
    serverDialog.visible = false
    await refreshData()
    ElMessage.success(t.value.serverDialog.saveSuccess)
  } catch (e) {
    ElMessage.error("Save Failed: " + e)
  }
}

const handleDeleteServer = async (id: string) => {
  try {
    await DeleteServer(id)
    if (currentServerId.value === id) currentServerId.value = ''
    await refreshData()
    ElMessage.success("Server Deleted")
  } catch(e) {
    ElMessage.error(e as string)
  }
}

const handleToggleServer = async (server: any, isOpen: boolean) => {
  if (!server) return
  try {
    await ModifyServers(server.id, isOpen)
    ElMessage.success(isOpen ? 'Starting Group...' : 'Stopping Group...')
  } catch (e) {
    ElMessage.error("Operation Failed: " + e)
  } finally {
    await refreshData()
  }
}

// === Link Actions ===
const openLinkDialog = (data: any = null) => {
  if (!currentServerId.value) return ElMessage.warning(t.value.linkDialog.warnServer)
  linkDialog.data = data
  linkDialog.isEdit = !!data
  linkDialog.visible = true
}

const onLinkSave = async (payload: any) => {
  const finalPayload = { ...payload }
  if (!linkDialog.isEdit) finalPayload.id = generateUUID()

  try {
    if (linkDialog.isEdit) {
      await ModifyLink(currentServerId.value, finalPayload.id, finalPayload)
    } else {
      await AddLink(currentServerId.value, finalPayload)
    }
    linkDialog.visible = false
    await refreshData()
    ElMessage.success(t.value.linkDialog.saveSuccess)
  } catch (e) {
    ElMessage.error("Save Failed: " + e)
  }
}

const handleToggleLink = async (serverId: string, linkId: string, isOpen: boolean) => {
  loadingState[linkId] = true
  try {
    await ToggleLinkStatus(serverId, linkId, isOpen)
    await nextTick()
    await refreshData()
  } catch (e) {
    ElMessage.error("Toggle Failed")
    await refreshData()
  } finally {
    loadingState[linkId] = false
  }
}

const handleDeleteLink = async (serverId: string, linkId: string) => {
  await DeleteLink(serverId, linkId)
  await refreshData()
  ElMessage.success("Tunnel Deleted")
}

// === Copy Logic ===
const handleCopyLink = (link: Link) => {
  let textToCopy = '';
  if (link.is_penetrate) {
    const server = currentServer.value;
    if (!server) return;
    textToCopy = `${server.server_host}:${link.remote_port}`;
  } else {
    textToCopy = `127.0.0.1:${link.local_port}`;
  }

  navigator.clipboard.writeText(textToCopy).then(() => {
    ElMessage.success(t.value.tunnel.copySuccess + textToCopy);
  }).catch(err => {
    ElMessage.error(t.value.tunnel.copyFail);
  });
}

const generateUUID = (): string => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    var r = Math.random() * 16 | 0, v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  overflow: hidden;
  background-color: var(--el-bg-color);
  color: var(--el-text-color-primary);
}
</style>