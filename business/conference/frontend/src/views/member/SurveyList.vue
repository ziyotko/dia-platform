<template>
  <div>
    <el-card>
      <el-table :data="list" stripe>
        <el-table-column prop="title" label="问卷标题" />
        <el-table-column label="状态" width="100"><template #default="{row}"><el-tag :type="row.status==='open'?'success':'info'">{{ row.status==='open'?'开放中':'已关闭' }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="120"><template #default="{row}"><el-button size="small" type="primary" @click="openSurvey(row)">填写问卷</el-button></template></el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="visible" :title="currentSurvey?.title" width="600px" :close-on-click-modal="false">
      <div v-for="q in currentSurvey?.questions" :key="q.id" style="margin-bottom:16px">
        <p><strong>{{ q.sort }}. {{ q.title }}</strong><span v-if="q.required" style="color:red">*</span></p>
        <el-radio-group v-model="answers[q.id]" v-if="q.type==='single'">
          <el-radio v-for="opt in q.options" :key="opt.id" :value="String(opt.id)">{{ opt.label }}</el-radio>
        </el-radio-group>
        <el-checkbox-group v-model="multiAnswers[q.id]" v-else-if="q.type==='multi'">
          <el-checkbox v-for="opt in q.options" :key="opt.id" :value="String(opt.id)">{{ opt.label }}</el-checkbox>
        </el-checkbox-group>
        <el-input v-model="answers[q.id]" type="textarea" v-else placeholder="请输入您的回答" />
      </div>
      <template #footer><el-button @click="visible=false">取消</el-button><el-button type="primary" @click="submit">提交</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { memberApi } from '@/api/member'

const list = ref<any[]>([])
const visible = ref(false); const currentSurvey = ref<any>(null)
const answers = ref<Record<number, string>>({}); const multiAnswers = ref<Record<number, string[]>>({})

async function load() { try { const res = await memberApi.getAvailableSurveys(); list.value = res.data || [] } catch (e) {} }

function openSurvey(s: any) { currentSurvey.value = s; answers.value = {}; multiAnswers.value = {}; visible.value = true }

async function submit() {
  try {
    const ans = Object.keys(answers.value).map(qid => ({
      questionId: Number(qid),
      answer: multiAnswers.value[Number(qid)] ? JSON.stringify(multiAnswers.value[Number(qid)]) : answers.value[Number(qid)]
    }))
    await memberApi.submitSurvey(currentSurvey.value.id, { answers: ans })
    ElMessage.success('提交成功'); visible.value = false; load()
  } catch (e: any) {}
}

load()
</script>
