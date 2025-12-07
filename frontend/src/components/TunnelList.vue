<template>
  <el-main class="main-panel">
    <div class="app-header" style="--wails-draggable:drag">

      <div class="header-info-wrapper">
        <div v-if="currentServer" class="header-info">
          <div class="title-row">
            <h2>{{ currentServer.server_name }}</h2>
            <el-tag size="small" type="info" class="server-tag">{{ currentServer.username }}@{{ currentServer.server_host }}:{{ currentServer.server_port }}</el-tag>
          </div>
          <div v-if="currentServer.notes" class="server-notes">
            {{ currentServer.notes }}
          </div>
        </div>
      </div>

      <div class="header-actions" style="--wails-draggable:no-drag">
        <el-button v-if="currentServer" type="primary" :icon="Plus" @click="$emit('add-link')">{{ t.tunnel.add }}</el-button>

        <div class="window-controls">
          <el-button link class="win-btn" @click="handleMinimize" :title="t.tunnel.minimize">
            <el-icon><Minus /></el-icon>
          </el-button>
          <el-button link class="win-btn" @click="handleToggleFullscreen" :title="t.tunnel.fullscreen">
            <el-icon><FullScreen /></el-icon>
          </el-button>
          <el-button link class="win-btn close-btn" @click="handleClose" :title="t.tunnel.close">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </div>
    </div>

    <div v-if="currentServer" class="content-body">
      <el-table :data="currentServer.link_group || []" style="width: 100%" row-key="id">
        <el-table-column :label="t.tunnel.status" width="80" align="center">
          <template #default="{ row }">
            <div class="status-dot-wrapper">
              <el-tooltip :content="isActive(row.id) ? t.tunnel.statusRun : t.tunnel.statusStop" placement="top">
                <div :class="['status-dot', isActive(row.id) ? 'active' : 'inactive']"></div>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="name" :label="t.tunnel.name" min-width="120" />

        <el-table-column :label="t.tunnel.type" width="100">
          <template #default="{ row }">
            <!-- 修改：内网穿透显示为 warning (黄色)，端口转发保持 success (绿色) -->
            <el-tag
                v-if="row.is_penetrate"
                effect="plain"
                size="small"
            >
              {{ t.tunnel.typeRev }}
            </el-tag>
            <el-tag
                v-else
                type="success"
                effect="plain"
                size="small"
            >
              {{ t.tunnel.typeLoc }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t.tunnel.mapping" min-width="220">
          <template #default="{ row }">
            <div class="mapping-column">
              <div class="mapping-row">
                <span class="mapping-label local">L</span>
                <span class="port-text">{{ row.local_host }}:{{ row.local_port }}</span>
              </div>
              <div class="mapping-row">
                <span class="mapping-label remote">R</span>
                <span class="port-text">{{ row.remote_host }}:{{ row.remote_port }}</span>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column :label="t.tunnel.switch" width="80">
          <template #default="{ row }">
            <el-switch
                :model-value="row.is_open"
                size="small"
                :loading="loadingState[row.id]"
                @change="(val) => $emit('toggle-link', currentServer.id, row.id, val)"
            />
          </template>
        </el-table-column>

        <el-table-column :label="t.tunnel.action" width="160" align="right">
          <template #default="{ row }">
            <el-button
                type="primary"
                link
                :icon="CopyDocument"
                @click="$emit('copy-link', row)"
                :title="t.tunnel.copy"
            />

            <el-button type="primary" link :icon="Edit" @click="$emit('edit-link', row)" :title="t.tunnel.edit" />
            <el-popconfirm :title="t.tunnel.deleteConfirm" @confirm="$emit('delete-link', currentServer.id, row.id)">
              <template #reference>
                <el-button type="danger" link :icon="Delete" :title="t.tunnel.delete" />
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div v-else class="empty-state">
      <el-empty :description="t.tunnel.empty" />
    </div>
  </el-main>
</template>

<script lang="ts" setup>
import { inject } from 'vue'
import { Plus, Edit, Delete, CopyDocument, Minus, FullScreen, Close } from '@element-plus/icons-vue'
import { WindowMinimise, WindowFullscreen, WindowUnfullscreen, WindowIsFullscreen, WindowHide } from '../../wailsjs/runtime/runtime'

const props = defineProps({
  currentServer: { type: Object, default: null },
  activeTunnelIds: { type: Array, default: () => [] },
  loadingState: { type: Object, default: () => ({}) }
})

defineEmits(['add-link', 'toggle-link', 'edit-link', 'delete-link', 'copy-link'])

const t: any = inject('t')

const isActive = (linkId: string): boolean => {
  if (!props.currentServer) return false
  const key = `${props.currentServer.id}_${linkId}`
  return props.activeTunnelIds.includes(key)
}

const handleMinimize = () => { WindowMinimise() }
const handleToggleFullscreen = async () => {
  const isFull = await WindowIsFullscreen()
  if (isFull) { WindowUnfullscreen() } else { WindowFullscreen() }
}
const handleClose = () => { WindowHide() }
</script>

<style scoped>
.main-panel {
  padding: 0;
  background-color: var(--el-bg-color);
  display: flex;
  flex-direction: column;
  height: 100%;
}
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 24px 24px 0 24px;
  flex-shrink: 0;
  cursor: default;
  min-height: 40px;
}
.header-info-wrapper { flex: 1; margin-right: 16px; }
.header-info { display: flex; flex-direction: column; align-items: flex-start; gap: 8px; }
.title-row { display: flex; align-items: center; gap: 12px; }
.header-info h2 { margin: 0; font-size: 24px; color: var(--el-text-color-primary); line-height: 1.2; }
.server-tag { font-family: monospace; }
.server-notes {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  max-width: 600px;
  line-height: 1.5;
  background: var(--el-fill-color-light);
  padding: 4px 8px;
  border-radius: 4px;
}
.header-actions { display: flex; align-items: center; gap: 16px; padding-top: 4px; }
.window-controls {
  display: flex;
  align-items: center;
  margin-left: 12px;
  padding-left: 12px;
  border-left: 1px solid var(--el-border-color-lighter);
}
.win-btn { padding: 8px; margin: 0 2px; color: var(--el-text-color-regular); font-size: 16px; }
.win-btn:hover { color: var(--el-color-primary); background-color: var(--el-fill-color-light); }
.win-btn.close-btn:hover { color: white; background-color: #f56c6c; }
.content-body { padding: 24px; flex: 1; overflow: auto; }
.mapping-column { display: flex; flex-direction: column; gap: 4px; }
.mapping-row { display: flex; align-items: center; font-family: monospace; font-size: 12px; line-height: 1.2; }
.mapping-label {
  display: inline-block; width: 16px; height: 16px; line-height: 16px; text-align: center;
  border-radius: 4px; margin-right: 6px; font-weight: bold; font-size: 10px;
}
.mapping-label.local { background-color: var(--el-color-success-light-9); color: var(--el-color-success); border: 1px solid var(--el-color-success-light-5); }
.mapping-label.remote { background-color: var(--el-color-warning-light-9); color: var(--el-color-warning); border: 1px solid var(--el-color-warning-light-5); }
.port-text { color: var(--el-text-color-regular); }
.status-dot-wrapper { display: flex; justify-content: center; align-items: center; height: 100%; }
.status-dot { width: 10px; height: 10px; border-radius: 50%; background-color: var(--el-border-color-darker); transition: all 0.3s ease; }
.status-dot.active { background-color: #67c23a; box-shadow: 0 0 6px rgba(103, 194, 58, 0.6); }
.status-dot.inactive { background-color: #909399; }
.empty-state { display: flex; justify-content: center; align-items: center; height: 100%; }
</style>