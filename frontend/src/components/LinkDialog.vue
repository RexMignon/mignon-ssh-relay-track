<template>
  <el-dialog
      :model-value="visible"
      :title="isEdit ? t.linkDialog.editTitle : t.linkDialog.addTitle"
      width="600px"
      @update:model-value="(val) => $emit('update:visible', val)"
      @closed="resetForm"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item :label="t.linkDialog.name" prop="name">
        <el-input v-model="form.name" :placeholder="t.linkDialog.namePh" />
      </el-form-item>

      <el-form-item :label="t.linkDialog.mode">
        <el-radio-group v-model="form.is_penetrate">
          <el-radio-button :value="false">{{ t.linkDialog.modeLoc }}</el-radio-button>
          <el-radio-button :value="true">{{ t.linkDialog.modeRev }}</el-radio-button>
        </el-radio-group>
        <div class="form-tip" v-if="!form.is_penetrate">
          {{ t.linkDialog.tipLoc }}
        </div>
        <div class="form-tip" v-else>
          {{ t.linkDialog.tipRev }}
        </div>
      </el-form-item>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-divider content-position="left">{{ t.linkDialog.localEnd }}</el-divider>
          <el-form-item :label="t.linkDialog.host" prop="local_host">
            <el-input v-model="form.local_host" />
          </el-form-item>
          <el-form-item :label="t.linkDialog.port" prop="local_port">
            <el-input-number v-model="form.local_port" :min="1" :max="65535" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-divider content-position="left">{{ t.linkDialog.remoteEnd }}</el-divider>
          <el-form-item :label="t.linkDialog.host" prop="remote_host">
            <el-input v-model="form.remote_host" />
          </el-form-item>
          <el-form-item :label="t.linkDialog.port" prop="remote_port">
            <el-input-number v-model="form.remote_port" :min="1" :max="65535" style="width: 100%" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item :label="t.linkDialog.defaultOpen">
        <el-switch v-model="form.is_open" />
      </el-form-item>

      <el-form-item :label="t.linkDialog.notes">
        <el-input v-model="form.notes" type="textarea" :placeholder="t.linkDialog.notesPh" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:visible', false)">{{ t.linkDialog.cancel }}</el-button>
      <el-button type="primary" @click="handleSave">{{ t.linkDialog.save }}</el-button>
    </template>
  </el-dialog>
</template>

<script lang="ts" setup>
import { reactive, watch, ref, inject } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'

const props = defineProps({
  visible: Boolean,
  isEdit: Boolean,
  initialData: Object
})

const emit = defineEmits(['update:visible', 'save'])
const t: any = inject('t')

const formRef = ref<FormInstance>()
const form = reactive({
  id: '',
  name: '',
  local_host: '127.0.0.1', local_port: 8080,
  remote_host: '127.0.0.1', remote_port: 80,
  is_penetrate: false, is_open: true, notes: ''
})

const rules = reactive<FormRules>({
  name: [{ required: true, message: 'Required', trigger: 'blur' }],
  local_host: [{ required: true, message: 'Required', trigger: 'blur' }],
  local_port: [{ required: true, message: 'Required', trigger: 'blur' }],
  remote_host: [{ required: true, message: 'Required', trigger: 'blur' }],
  remote_port: [{ required: true, message: 'Required', trigger: 'blur' }]
})

watch(() => props.visible, (val) => {
  if (val && props.isEdit && props.initialData) {
    form.id = props.initialData.id
    form.name = props.initialData.name
    form.local_host = props.initialData.local_host
    form.local_port = props.initialData.local_port
    form.remote_host = props.initialData.remote_host
    form.remote_port = props.initialData.remote_port
    form.is_penetrate = props.initialData.is_penetrate
    form.is_open = props.initialData.is_open
    form.notes = props.initialData.notes
  } else if (val && !props.isEdit) {
    resetForm()
  }
})

const resetForm = () => {
  if (formRef.value) formRef.value.clearValidate()
  form.id = ''
  form.name = ''
  form.local_host = '127.0.0.1'
  form.local_port = 8080
  form.remote_host = '127.0.0.1'
  form.remote_port = 80
  form.is_penetrate = false
  form.is_open = true
  form.notes = ''
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (valid) {
      const payload = { ...form }
      payload.local_port = parseInt(payload.local_port as any)
      payload.remote_port = parseInt(payload.remote_port as any)
      emit('save', payload)
    } else {
      ElMessage.warning(t.value.linkDialog.validFail)
    }
  })
}
</script>

<style scoped>
.form-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 5px;
  line-height: 1.4;
  background: var(--el-fill-color-light);
  padding: 8px;
  border-radius: 4px;
}
</style>