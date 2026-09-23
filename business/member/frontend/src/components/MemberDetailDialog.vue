<template>
  <el-dialog
    :model-value="modelValue"
    width="800px"
    class="member-detail-dialog"
    :show-close="false"
    :close-on-click-modal="true"
    append-to-body
    @update:model-value="onVisibleChange"
  >
    <template #header>
      <div class="detail-head" v-if="detail">
        <el-avatar :size="60" :src="fileUrl(detail.avatar)" class="detail-avatar">{{ initial(detail) }}</el-avatar>
        <div class="detail-title">
          <div class="name-line">
            <span class="name">{{ displayName(detail) }}</span>
            <el-tag :type="statusTag(detail.status)" size="large">{{ statusLabel(detail.status) }}</el-tag>
          </div>
          <div class="sub">{{ detail.username }} · {{ typeText(detail.member_type) }}<template v-if="detail.member_level"> · 等级 {{ levelText(detail.member_level) }}</template></div>
        </div>
      </div>
    </template>

    <div v-loading="loading" class="detail-body">
      <template v-if="detail">
        <div class="section">
          <div class="section-title">基本资料</div>
          <div class="grid">
            <div class="item"><span class="label">会员类型</span><span class="value">{{ typeText(detail.member_type) }}</span></div>
            <div class="item"><span class="label">入会机构</span><span class="value">{{ detail.org_name || '-' }}</span></div>
            <div class="item"><span class="label">会员等级</span><span class="value">{{ levelText(detail.member_level) }}</span></div>
            <div class="item"><span class="label">注册时间</span><span class="value">{{ fmt(detail.created_at) }}</span></div>
            <div class="item"><span class="label">最近更新</span><span class="value">{{ fmt(detail.updated_at) }}</span></div>
            <div class="item"><span class="label">手机号</span><span class="value">{{ detail.mobile || '-' }}</span></div>
            <div class="item"><span class="label">邮箱</span><span class="value">{{ detail.email || '-' }}</span></div>
          </div>
        </div>

        <div class="section" v-if="detail.member_type === 'unit'">
          <div class="section-title">单位信息</div>
          <div class="grid">
            <div class="item"><span class="label">公司名称</span><span class="value">{{ detail.company_name || '-' }}</span></div>
            <div class="item"><span class="label">统一社会信用代码</span><span class="value">{{ detail.credit_code || '-' }}</span></div>
            <div class="item"><span class="label">法定代表人</span><span class="value">{{ detail.legal_person || '-' }}</span></div>
            <div class="item"><span class="label">所属行业</span><span class="value">{{ detail.industry || '-' }}</span></div>
            <div class="item"><span class="label">联系人</span><span class="value">{{ detail.contact_person || '-' }}<template v-if="detail.contact_title">（{{ detail.contact_title }}）</template></span></div>
            <div class="item"><span class="label">联系电话</span><span class="value">{{ detail.contact_mobile || '-' }}</span></div>
            <div class="item"><span class="label">成立日期</span><span class="value">{{ detail.founded_date || '-' }}</span></div>
            <div class="item"><span class="label">注册资本</span><span class="value">{{ detail.registered_capital || '-' }}</span></div>
            <div class="item"><span class="label">员工人数</span><span class="value">{{ detail.employee_count != null ? detail.employee_count + ' 人' : '-' }}</span></div>
            <div class="item"><span class="label">邮编</span><span class="value">{{ detail.postal_code || '-' }}</span></div>
            <div class="item span2"><span class="label">地址</span><span class="value">{{ detail.address || '-' }}</span></div>
            <div class="item"><span class="label">网站</span><span class="value"><el-link v-if="detail.website" :href="detail.website" target="_blank" type="primary">{{ detail.website }}</el-link><template v-else>-</template></span></div>
            <div class="item"><span class="label">营业执照</span><span class="value"><el-link v-if="detail.cert_file" :href="fileUrl(detail.cert_file)" target="_blank" type="primary">查看 / 下载</el-link><template v-else>-</template></span></div>
            <div class="item span2"><span class="label">经营范围</span><span class="value">{{ detail.business_scope || '-' }}</span></div>
            <div class="item span2"><span class="label">公司简介</span><span class="value">{{ detail.description || '-' }}</span></div>
          </div>
        </div>

        <div class="section" v-if="detail.member_type === 'personal'">
          <div class="section-title">个人信息</div>
          <div class="grid">
            <div class="item"><span class="label">姓名</span><span class="value">{{ detail.name || '-' }}</span></div>
            <div class="item"><span class="label">身份证号</span><span class="value">{{ detail.id_card || '-' }}</span></div>
          </div>
        </div>
      </template>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { fileUrl } from '@/utils/fileUrl'

/**
 * 会员详情弹窗（管理后台「会员管理 / 会籍记录 / 资料记录」共用）。
 * dialog 使用 append-to-body teleport 到 body，故正文样式写在下方非 scoped 的
 * `.member-detail-dialog` 作用域里（scoped 样式不会作用到 teleport 出去的内容）。
 */
const props = withDefaults(defineProps<{
  /** 是否显示（v-model） */
  modelValue: boolean
  /** 会员详情数据（GET /admin/members/:id） */
  detail: any
  /** 详情加载中 */
  loading?: boolean
  /** 会员类型文案，默认「单位会员 / 个人会员」 */
  typeLabel?: (t: string) => string
  /** 会员等级名称解析（各页用自己的 levels 列表查表），默认原样显示 */
  levelName?: (id: any) => string
}>(), { loading: false })

const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void }>()

function onVisibleChange(v: boolean) {
  emit('update:modelValue', v)
}

const statusMap: Record<string, { l: string; t: string }> = {
  registering: { l: '注册中', t: 'info' },
  pending_review: { l: '待审核', t: 'warning' },
  pending_payment: { l: '待缴费', t: 'danger' },
  active: { l: '正式会员', t: 'success' },
  rejected: { l: '已拒绝', t: 'danger' },
  expired: { l: '已过期', t: 'info' }
}

function statusLabel(s?: string) { return (s && statusMap[s]?.l) || s || '-' }
function statusTag(s?: string) { return (s && statusMap[s]?.t) || 'info' }
function displayName(m: any) {
  if (!m) return '-'
  return m.member_type === 'unit' ? (m.company_name || m.username || '-') : (m.name || m.username || '-')
}
function initial(m: any) { return ((displayName(m) || '?').trim()[0] || '?').toUpperCase() }
function fmt(d?: string) { return d ? d.replace('T', ' ').slice(0, 16) : '-' }
function typeText(t?: string) {
  if (props.typeLabel) return props.typeLabel(t || '')
  return t === 'unit' ? '单位会员' : t === 'personal' ? '个人会员' : (t || '-')
}
function levelText(id: any) {
  if (props.levelName) return props.levelName(id)
  return id === null || id === undefined || id === '' ? '-' : String(id)
}
</script>

<!-- 详情弹窗样式（dialog 使用 teleport，故使用非 scoped 样式并以类名限定作用域） -->
<style lang="scss">
.member-detail-dialog {
  border-radius: 14px;
  overflow: hidden;

  .el-dialog__header {
    padding: 22px 24px 16px;
    background: linear-gradient(135deg, #f4f7ff 0%, #eef1ff 100%);
    border-bottom: 1px solid #eef0f3;
    margin-right: 0;
  }
  .el-dialog__body {
    padding: 24px;
  }

  .detail-head {
    display: flex;
    align-items: center;
    gap: 16px;
    padding-right: 40px;
  }
  .detail-avatar {
    background: linear-gradient(135deg, #002fa7, #002686);
    color: #fff;
    font-size: 24px;
    font-weight: 600;
    flex-shrink: 0;
  }
  .detail-title {
    .name-line {
      display: flex;
      align-items: center;
      gap: 12px;
      .name { font-size: 20px; font-weight: 600; color: #1d2739; }
    }
    .sub { margin-top: 6px; color: #909399; font-size: 13px; }
  }

  .detail-body {
    padding-top: 4px;
    .section { margin-bottom: 26px; &:last-child { margin-bottom: 0; } }
    .section-title {
      font-size: 15px;
      font-weight: 600;
      color: #1d2739;
      padding-left: 10px;
      border-left: 4px solid #002fa7;
      margin-bottom: 14px;
      line-height: 1.2;
    }
    .grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 14px 28px;
      background: #fafbfc;
      border: 1px solid #eef0f3;
      border-radius: 10px;
      padding: 18px 22px;
      .item {
        display: flex;
        flex-direction: column;
        gap: 4px;
        &.span2 { grid-column: 1 / -1; }
      }
      .label { font-size: 12px; color: #909399; }
      .value { font-size: 14px; color: #303133; line-height: 1.6; word-break: break-word; }
    }
  }
}
</style>
