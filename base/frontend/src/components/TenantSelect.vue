<template>
  <el-select
    :model-value="modelValue"
    :placeholder="placeholder"
    :clearable="clearable"
    :disabled="disabled"
    filterable
    style="width: 100%"
    @update:model-value="(v: any) => emit('update:modelValue', v)"
  >
    <el-option label="平台级（不属于任何租户）" :value="0" />
    <el-option
      v-for="t in tenantOptions"
      :key="t.id"
      :label="`${t.name}（${t.code}）`"
      :value="t.id"
    />
  </el-select>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { tenantOptions, ensureTenants } from '@/utils/tenantOptions'

withDefaults(
  defineProps<{
    modelValue?: number
    placeholder?: string
    clearable?: boolean
    disabled?: boolean
  }>(),
  {
    modelValue: 0,
    placeholder: '请选择租户',
    clearable: false,
    disabled: false
  }
)

const emit = defineEmits<{ (e: 'update:modelValue', value: number | undefined): void }>()

onMounted(() => {
  ensureTenants()
})
</script>
