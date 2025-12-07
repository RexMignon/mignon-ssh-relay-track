<template>
  <el-dialog
      :model-value="visible"
      :title="isEdit ? t.serverDialog.editTitle : t.serverDialog.addTitle"
      width="500px"
      @update:model-value="(val) => $emit('update:visible', val)"
      @closed="resetForm"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item :label="t.serverDialog.name" prop="server_name">
        <el-input v-model="form.server_name" :placeholder="t.serverDialog.namePh" />
      </el-form-item>
      <el-row :gutter="20">
        <el-col :span="16">
          <el-form-item :label="t.serverDialog.host" prop="server_host">
            <el-input v-model="form.server_host" :placeholder="t.serverDialog.hostPh" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item :label="t.serverDialog.port" prop="server_port" label-width="auto">
            <el-input-number
                v-model="form.server_port"
                :min="1"
                :max="65535"
                style="width: 100%"
                :controls="false"
                :placeholder="t.serverDialog.portPh"
            />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item :label="t.serverDialog.username" prop="username">
        <el-input v-model="form.username" :placeholder="t.serverDialog.usernamePh" />
      </el-form-item>
      <el-form-item :label="t.serverDialog.password" prop="password">
        <el-input v-model="form.password" type="password" show-password />
      </el-form-item>
      <el-form-item :label="t.serverDialog.notes">
        <el-input v-model="form.notes" type="textarea" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:visible', false)">{{ t.serverDialog.cancel }}</el-button>
      <el-button type="primary" @click="handleSave">{{ t.serverDialog.save }}</el-button>
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
  server_name: '', server_host: '', server_port: 22,
  username: 'root', password: '', notes: '', is_open: true
})

const rules = reactive<FormRules>({
  server_name: [{ required: true, message: 'Required', trigger: 'blur' }],
  server_host: [{ required: true, message: 'Required', trigger: 'blur' }],
  server_port: [{ required: true, message: 'Required', trigger: 'blur' }],
  username: [{ required: true, message: 'Required', trigger: 'blur' }],
  password: [{ required: true, message: 'Required', trigger: 'blur' }]
})

watch(() => props.visible, (val) => {
  if (val && props.isEdit && props.initialData) {
    form.id = props.initialData.id
    form.server_name = props.initialData.server_name
    form.server_host = props.initialData.server_host
    form.server_port = props.initialData.server_port
    form.username = props.initialData.username
    form.password = props.initialData.password
    form.notes = props.initialData.notes
    form.is_open = props.initialData.is_open
  } else if (val && !props.isEdit) {
    resetForm()
  }
})

const resetForm = () => {
  if (formRef.value) formRef.value.clearValidate()
  form.id = ''
  form.server_name = ''
  form.server_host = ''
  form.server_port = 22
  form.username = 'root'
  form.password = ''
  form.notes = ''
  form.is_open = true
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (valid) {
      const payload = { ...form }
      payload.server_port = parseInt(payload.server_port as any)
      emit('save', payload)
    } else {
      ElMessage.warning(t.value.serverDialog.validFail)
    }
  })
}
</script>