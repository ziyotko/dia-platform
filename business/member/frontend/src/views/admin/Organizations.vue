<template>
  <div class="admin-orgs" v-loading="loading">
    <div class="page-header"><h3>组织机构管理</h3><el-button type="primary" @click="showCreate=true">新增组织</el-button></div>
    <el-card>
      <el-tree :data="tree" :props="{label:'name',children:'children'}" node-key="id" default-expand-all>
        <template #default="{node,data}">
          <span class="tree-node">
            <span>{{ node.label }}</span>
            <span class="actions">
              <el-button text size="small" type="primary" @click.stop="editOrg(data)">编辑</el-button>
              <el-button text size="small" type="danger" @click.stop="delOrg(data)">删除</el-button>
            </span>
          </span>
        </template>
      </el-tree>
    </el-card>

    <el-dialog v-model="showCreate" :title="editingOrg?'编辑组织':'新增组织'" width="480px">
      <el-form :model="orgForm" size="large">
        <el-form-item label="名称" required><el-input v-model="orgForm.name" /></el-form-item>
        <el-form-item label="类型"><el-select v-model="orgForm.type"><el-option label="协会" value="association" /><el-option label="分会" value="branch" /><el-option label="委员会" value="committee" /></el-select></el-form-item>
        <el-form-item label="描述"><el-input v-model="orgForm.description" type="textarea" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="orgForm.sort" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="showCreate=false">取消</el-button><el-button type="primary" @click="saveOrg">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { orgApi } from '@/api/index'
import { ElMessage, ElMessageBox } from 'element-plus'

const tree = ref<any[]>([]); const loading = ref(true); const showCreate = ref(false); const editingOrg = ref<any>(null)
const orgForm = reactive({ name: '', type: 'branch', description: '', sort: 0 })

onMounted(async () => {
  try { const r = await orgApi.getTree(); tree.value = r.data || [] } catch {} finally { loading.value = false }
})

function editOrg(data: any) { editingOrg.value = data; Object.assign(orgForm, data); showCreate.value = true }

async function saveOrg() {
  try {
    if (editingOrg.value) { await adminApi.updateOrg(editingOrg.value.id, orgForm); ElMessage.success('更新成功') }
    else { await adminApi.createOrg({ name: orgForm.name, type: orgForm.type, description: orgForm.description, sort: orgForm.sort }); ElMessage.success('创建成功') }
    showCreate.value = false; editingOrg.value = null
    const r = await orgApi.getTree(); tree.value = r.data || []
  } catch {}
}
async function delOrg(data: any) {
  try { await ElMessageBox.confirm('确认删除？', '警告', { type: 'warning' }); await adminApi.deleteOrg(data.id); ElMessage.success('已删除'); const r = await orgApi.getTree(); tree.value = r.data || [] } catch {}
}
</script>

<style scoped lang="scss">
.admin-orgs { max-width: 700px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.tree-node { display: flex; justify-content: space-between; align-items: center; width: 100%; .actions { display: none; } }
.tree-node:hover .actions { display: inline; }
</style>
