<template>
  <div>
    <el-card>
      <h3>问卷结果</h3>
      <p>总参与人数：{{ totalRespondents }}</p>
      <div v-for="q in questions" :key="q.id" style="margin:16px 0;padding:12px;background:#f9fafb;border-radius:8px">
        <strong>{{ q.title }}</strong>
        <div v-if="q.type!=='text'">
          <div v-for="opt in q.options" :key="opt.id">{{ opt.label }} - {{ opt.count }} 人</div>
        </div>
        <div v-else style="color:#999">简答题</div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { adminApi } from '@/api/admin'

const route = useRoute()
const questions = ref<any[]>([]); const totalRespondents = ref(0)

onMounted(async () => {
  try {
    const res = await adminApi.getSurveyResults(Number(route.params.id))
    questions.value = res.data?.survey?.questions || []
    totalRespondents.value = res.data?.total_respondents || 0
  } catch (e) {}
})
</script>
