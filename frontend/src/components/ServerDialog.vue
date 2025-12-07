<template>
  <el-dialog
      :model-value="visible"
      :title="isEdit ? t.serverDialog.editTitle : t.serverDialog.addTitle"
      width="500px"
      @update:model-value="(val) => $emit('update:visible', val)"
      @closed="resetForm"
  >
    <!-- 添加 ref 和 rules -->
    <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
      <!-- 添加 prop 以启用校验 -->
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
        <!-- 核心修改：移除 show-password，确保密码不可查看 -->
        <!-- 当处于编辑模式时，v-model将是 PASSWORD_MASK，显示为 16 个圆点 -->
        <el-input v-model="form.password" type="password" />
      </el-form-item>
      <!-- 备注无需 prop，因为不需要校验 -->
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

// 定义一个常量用于表示未更改的密码（且用于占位显示）
const PASSWORD_MASK = '****************'; // 16个星号，用于占位显示

const formRef = ref<FormInstance>()
const form = reactive({
  id: '',
  server_name: '', server_host: '', server_port: 22,
  username: 'root', password: '', notes: '', is_open: true
})

// 定义校验规则
const rules = reactive<FormRules>({
  server_name: [{ required: true, message: 'Required', trigger: 'blur' }],
  server_host: [{ required: true, message: 'Required', trigger: 'blur' }],
  server_port: [{ required: true, message: 'Required', trigger: 'blur' }],
  username: [{ required: true, message: 'Required', trigger: 'blur' }],
  // 密码校验：新增时必须，修改时如果用户输入了新值也必须是有效值
  password: [{ required: true, message: 'Required', trigger: 'blur' }]
})

// 监听打开时传入的数据
watch(() => props.visible, (val) => {
  if (val && props.isEdit && props.initialData) {
    // 保持之前修复的防污染逻辑
    form.id = props.initialData.id
    form.server_name = props.initialData.server_name
    form.server_host = props.initialData.server_host
    form.server_port = props.initialData.server_port
    form.username = props.initialData.username
    // 核心修改：在编辑模式下，将密码设置为占位符，原始密码数据不会加载
    form.password = PASSWORD_MASK
    form.notes = props.initialData.notes
    form.is_open = props.initialData.is_open
  } else if (val && !props.isEdit) {
    resetForm()
  }
})

const resetForm = () => {
  if (formRef.value) formRef.value.clearValidate() // 清除校验红字
  form.id = ''
  form.server_name = ''
  form.server_host = ''
  form.server_port = 22
  form.username = 'root'
  form.password = '' // 新增时密码为空
  form.notes = ''
  form.is_open = true
}

const handleSave = async () => {
  if (!formRef.value) return

  // 执行校验
  await formRef.value.validate((valid) => {
    if (valid) {
      const payload = { ...form }
      payload.server_port = parseInt(payload.server_port as any)

      // 核心修改：如果密码未更改（仍是占位符），则不发送密码字段
      if (props.isEdit && payload.password === PASSWORD_MASK) {
        delete payload.password // 后端通过缺失该字段来判断不更新密码
      }

      emit('save', payload)
    } else {
      ElMessage.warning(t.value.serverDialog.validFail)
    }
  })
}
</script>