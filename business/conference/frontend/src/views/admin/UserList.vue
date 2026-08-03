<template>
  <div>
    <el-card>
      <div style="margin-bottom:16px"><el-button type="primary" @click="createVisible=true">添加用户</el-button></div>
      <el-table :data="list" stripe>
        <el-table-column prop="realName" label="姓名" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="phone" label="手机号" />
        <el-table-column prop="branch" label="分会" />
        <el-table-column prop="memberLevel" label="等级" />
        <el-table-column label="有效" width="80"><template #default="{row}"><el-tag :type="row.isValid?'success':'danger'" size="small">{{ row.isValid?'有效':'无效' }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="180"><template #default="{row}">
          <el-button size="small" @click="toggleValid(row)">{{ row.isValid?'设为无效':'设为有效' }}</el-button>
          <el-button size="small" type="danger" @click="del(row.id)">删除</el-button>
        </template></el-table-column>
      </el-table>
      <el-pagination v-if="total>size" v-model:current-page="page" :page-size="size" :total="total" layout="prev,pager,next" @current-change="load" style="margin-top:16px" />
    </el-card>

    <el-dialog v-model="createVisible" title="添加用户" width="400px">
      <el-form :model="createForm">
        <el-form-item label="用户名"><el-input v-model="createForm.username" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="createForm.password" type="password" /></el-form-item>
        <el-form-item label="姓名"><el-input v-model="createForm.realName" /></el-form-item>
        <el-form-item label="手机号"><el-input v-model="createForm.phone" /></el-form-item>
        <el-form-item label="分会"><el-input v-model="createForm.branch" /></el-form-item>
        <el-form-item label="等级"><el-input v-model="createForm.memberLevel" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="createVisible=false">取消</el-button><el-button type="primary" @click="doCreate">确认</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { adminApi } from '@/api/admin'

const list = ref<any[]>([]); const page = ref(1); const size = 10; const total = ref(0)
const createVisible = ref(false)
const createForm = reactive({ username: '', password: '', realName: '', phone: '', branch: '', memberLevel: '' })

async function load() { try { const res = await adminApi.getUsers({ page: page.value, pageSize: size }); list.value = res.data?.list || []; total.value = res.data?.total || 0 } catch (e) {} }

async function toggleValid(row: any) { try { await adminApi.setValidity(row.id, { isValid: !row.isValid }); load() } catch (e: any) {} }
async function del(id: number) { try { await ElMessageBox.confirm('确认删除？'); await adminApi.deleteUser(id); load() } catch (e) {} }
async function doCreate() { try { await adminApi.createUser(createForm); ElMessage.success('创建成功'); createVisible.value = false; load() } catch (e: any) {} }

load()
</script>
