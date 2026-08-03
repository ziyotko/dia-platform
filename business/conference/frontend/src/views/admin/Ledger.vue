<template>
  <div>
    <el-card>
      <el-table :data="ledger" stripe>
        <el-table-column prop="meeting_title" label="会议" />
        <el-table-column label="收入"><template #default="{row}">¥{{ row.income }}</template></el-table-column>
        <el-table-column label="退款"><template #default="{row}">¥{{ row.refunded }}</template></el-table-column>
        <el-table-column label="净收入"><template #default="{row}">¥{{ row.net_income }}</template></el-table-column>
        <el-table-column prop="order_count" label="订单数" width="80" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'

const ledger = ref<any[]>([])

onMounted(async () => { try { const res = await adminApi.getLedger(); ledger.value = res.data || [] } catch (e) {} })
</script>
