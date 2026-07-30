<template>
  <span ref="elRef">{{ displayValue }}</span>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

const props = withDefaults(defineProps<{
  num: number
  duration?: number
}>(), {
  duration: 800
})

const elRef = ref<HTMLElement>()
const displayValue = ref(0)
let animId = 0

function animate(from: number, to: number) {
  cancelAnimationFrame(animId)
  const start = performance.now()

  function tick(now: number) {
    const elapsed = now - start
    const progress = Math.min(elapsed / props.duration, 1)
    // ease-out cubic
    const eased = 1 - Math.pow(1 - progress, 3)
    displayValue.value = Math.round(from + (to - from) * eased)
    if (progress < 1) {
      animId = requestAnimationFrame(tick)
    }
  }

  animId = requestAnimationFrame(tick)
}

watch(() => props.num, (val, old) => {
  animate(old || 0, val)
})

onMounted(() => {
  displayValue.value = props.num
})
</script>
