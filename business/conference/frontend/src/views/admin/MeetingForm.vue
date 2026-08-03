<template>
  <div>
    <el-card>
      <h3>{{ isEdit ? '编辑会议' : '创建会议' }}</h3>
      <el-form :model="form" label-width="110px">

        <!-- 基本信息 -->
        <el-divider content-position="left">基本信息</el-divider>
        <el-form-item label="会议名称" required>
          <el-input v-model="form.title" placeholder="请输入会议名称" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="会议类型">
              <el-select v-model="form.type" style="width:100%">
                <el-option label="线上会议" value="online" />
                <el-option label="线下会议" value="offline" />
                <el-option label="混合会议" value="hybrid" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="会议状态">
              <el-select v-model="form.status" style="width:100%">
                <el-option label="草稿" value="draft" />
                <el-option label="发布(开放报名)" value="open" />
                <el-option label="已关闭" value="closed" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="会议时间" required>
          <el-date-picker v-model="timeRange" type="datetimerange" range-separator="至"
            start-placeholder="开始时间" end-placeholder="结束时间" value-format="YYYY-MM-DDTHH:mm:ss" style="width:100%" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="名额上限">
              <el-input-number v-model="form.capacity" :min="0" :max="100000" />
              <div class="tip">0 表示不限名额</div>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="会议费用">
              <el-input-number v-model="form.fee" :min="0" :precision="2" />
              <div class="tip">0 表示免费</div>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="会议地点" v-if="form.type !== 'online'">
          <el-input v-model="form.location" placeholder="线下会议地点，线上可留空" />
        </el-form-item>
        <el-form-item label="封面图">
          <el-input v-model="form.coverImage" placeholder="输入封面图片 URL" />
        </el-form-item>
        <el-form-item label="会议描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="会议简介、注意事项等" />
        </el-form-item>

        <!-- 报名与签到时间 -->
        <el-divider content-position="left">报名与签到时间</el-divider>
        <el-form-item label="报名时间">
          <el-date-picker v-model="regRange" type="datetimerange" range-separator="至"
            start-placeholder="报名开始" end-placeholder="报名结束" value-format="YYYY-MM-DDTHH:mm:ss" style="width:100%" />
          <div class="tip">留空表示会议发布后即可报名</div>
        </el-form-item>
        <el-form-item label="签到时间">
          <el-date-picker v-model="signRange" type="datetimerange" range-separator="至"
            start-placeholder="签到开始" end-placeholder="签到结束" value-format="YYYY-MM-DDTHH:mm:ss" style="width:100%" />
          <div class="tip">留空表示会议时间内均可签到</div>
        </el-form-item>
        <el-form-item label="取消报名截止">
          <el-date-picker v-model="form.cancelDeadline" type="datetime"
            placeholder="在此时间后不可取消报名" value-format="YYYY-MM-DDTHH:mm:ss" style="width:100%" />
        </el-form-item>

        <!-- 参会权限 -->
        <el-divider content-position="left">参会权限</el-divider>
        <el-form-item label="会员等级限制">
          <el-select v-model="form.accessLevels" multiple filterable allow-create default-first-option
            placeholder="选择或输入会员等级，留空不限制" style="width:100%">
            <el-option v-for="lv in levelPresets" :key="lv" :label="lv" :value="lv" />
          </el-select>
        </el-form-item>
        <el-form-item label="分会限制">
          <el-select v-model="form.accessBranches" multiple filterable allow-create default-first-option
            placeholder="选择或输入分会，留空不限制" style="width:100%">
            <el-option v-for="b in branchPresets" :key="b" :label="b" :value="b" />
          </el-select>
        </el-form-item>

        <!-- 功能开关 -->
        <el-divider content-position="left">功能开关</el-divider>
        <el-row :gutter="16">
          <el-col :span="8"><el-form-item label="报名审核"><el-switch v-model="form.needApproval" active-text="需审核" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="候补机制"><el-switch v-model="form.allowWaitlist" active-text="满员候补" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="会议收费"><el-switch v-model="form.isPaid" active-text="在线缴费" /></el-form-item></el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8"><el-form-item label="开启直播"><el-switch v-model="form.enableLive" active-text="直播录播" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="开启投票"><el-switch v-model="form.enableVote" active-text="投票表决" /></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="开启问卷"><el-switch v-model="form.enableSurvey" active-text="问卷调研" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="开启学分"><el-switch v-model="form.enableCredit" active-text="签到发放学分" /></el-form-item>

        <el-divider />
        <el-form-item>
          <el-button type="primary" @click="save">保存</el-button>
          <el-button @click="$router.back()">返回</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { adminApi } from '@/api/admin'

const route = useRoute()
const isEdit = ref(false)
const timeRange = ref<any[]>([])
const regRange = ref<any[]>([])
const signRange = ref<any[]>([])

const levelPresets = ['普通会员', '高级会员', '理事会员', '荣誉会员']
const branchPresets = ['第一分会', '第二分会', '第三分会', '第四分会']

const form = reactive<any>({
  title: '', type: 'offline', status: 'draft', capacity: 0, fee: 0, location: '', coverImage: '',
  description: '', cancelDeadline: null,
  accessLevels: [], accessBranches: [],
  needApproval: false, allowWaitlist: false, isPaid: false,
  enableLive: false, enableVote: false, enableSurvey: false, enableCredit: false
})

onMounted(async () => {
  const id = route.params.id
  if (id) {
    isEdit.value = true
    try {
      const res = await adminApi.getMeeting(Number(id))
      Object.assign(form, res.data)
      if (res.data.startTime && res.data.endTime) timeRange.value = [res.data.startTime, res.data.endTime]
      if (res.data.regStartTime && res.data.regEndTime) regRange.value = [res.data.regStartTime, res.data.regEndTime]
      if (res.data.signStartTime && res.data.signEndTime) signRange.value = [res.data.signStartTime, res.data.signEndTime]
    } catch (e) {}
  }
})

async function save() {
  try {
    if (timeRange.value.length === 2) {
      form.startTime = timeRange.value[0]
      form.endTime = timeRange.value[1]
    }
    if (regRange.value.length === 2) {
      form.regStartTime = regRange.value[0]
      form.regEndTime = regRange.value[1]
    }
    if (signRange.value.length === 2) {
      form.signStartTime = signRange.value[0]
      form.signEndTime = signRange.value[1]
    }
    if (isEdit.value) await adminApi.updateMeeting(Number(route.params.id), form)
    else await adminApi.createMeeting(form)
    ElMessage.success('保存成功')
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  }
}
</script>

<style scoped>
.tip { font-size: 12px; color: #999; line-height: 1.4; }
</style>
