# Base 底座安全开发指南

本文档说明 Base 后端在**接口级权限鉴权**和**租户数据隔离**方面的实现规范，供后续开发参考。

---

## 1. 接口级权限鉴权

### 1.1 中间件位置

[internal/middleware/permission.go](../internal/middleware/permission.go)

### 1.2 设计目标

- 所有需要登录的接口，在 JWT 认证之后，额外校验当前用户是否拥有该接口的访问权限。
- 权限数据来自 `base_permission` 表，通过 `base_role_permission` 和 `base_user_role` 关联到用户。
- 前端返回的权限码仅用于 UI 控制，真正的鉴权在后端完成。

### 1.3 匹配规则

中间件使用当前请求的 **HTTP Method + 路由模板** 去匹配权限表的 `method` 和 `path` 字段：

- 路由模板示例：`/base/api/v1/users/:id`
- 权限表 `path` 可配置完整路径 `/base/api/v1/users/:id`，也可配置相对路径 `/users/:id`。
- 支持 `:param` 通配符匹配，暂不支持 `*any` 通配符。

### 1.4 放行策略

中间件按以下顺序判定：

1. **平台超级管理员**：`tenantID == 0` 的用户直接放行。
2. **租户管理员**：`base_user.is_admin = 1` 的用户在本租户内直接放行（与「管理员可见全部菜单」保持同一口径；数据仍由 service 层的 `tenant_id` 条件隔离）。
3. **白名单接口**：用户基础信息、菜单、权限、改密、仪表盘、我的应用、未读消息数等基础接口。
   - `GET /base/api/v1/auth/info`
   - `GET /base/api/v1/auth/menus`
   - `GET /base/api/v1/auth/permissions`
   - `POST /base/api/v1/auth/change-password`
   - `GET /base/api/v1/dashboard/stats`
   - `GET /base/api/v1/app-instances/my`
   - `GET /base/api/v1/messages/unread-count`
   - 工作流「属于自己的数据」类接口（归属校验均在 service 层，详见 1.4.1）：
     - `GET /base/api/v1/workflows/options`（发起流程需先选流程定义）
     - `POST|GET /base/api/v1/workflow-instances`
     - `GET /base/api/v1/workflow-instances/:id`
     - `POST /base/api/v1/workflow-instances/:id/cancel`
     - `GET /base/api/v1/workflow-tasks`
     - `POST /base/api/v1/workflow-tasks/:id/approve` / `reject`
4. **未分配任何权限的普通用户**：除白名单接口外一律返回 403「未分配任何接口权限，请联系管理员在「角色管理 → 分配权限」中授权」。
   （旧行为是「只读放行所有 `GET`」，会让零权限用户读到 `/settings`（含 SMTP 密码）、`/users`、`/operation-logs` 等敏感数据，已收窄。）
5. 其余请求按 `base_permission` 的 `method + path` 匹配，未命中返回 403「无权限访问该接口」。

> ⚠️ 历史行为（已取消）：早期版本在「用户没有任何权限记录」时 **整体放行**，导致全新部署下所有登录用户可调用全部写接口。

#### 1.4.1 工作流为何部分接口在白名单里

审批是「用户自己的事」：任何被分配了待办的普通用户都必须能查看与处理，而不应依赖管理员再额外分配接口权限。因此这些接口不做接口级校验，改用**归属校验**保证安全：

| 接口 | 归属校验（service 层） |
|------|------------------------|
| 发起流程 | 只能发起本租户（平台超管例外）已启用的流程 |
| 我的申请 / 实例列表 | 非管理员强制 `initiator_id = 当前用户` |
| 实例详情 | 仅发起人、该实例的审批参与人、管理员可见 |
| 撤销 | 仅发起人本人或管理员，且实例必须处于「审批中」 |
| 待办列表 / 审批 / 驳回 | 按 `approver_id = 当前用户` 过滤，且校验任务归属与状态（不能重复处理、不能处理已结束流程） |

而**流程定义（含节点编排）**属于管理动作，仍走权限点 `base:workflow-def:*`；**删除流程实例**属于管理动作，走 `base:workflow-instance:delete`（不在白名单）。

### 1.5 挂载方式

在 [internal/routes/routes.go](../internal/routes/routes.go) 中，对已登录接口统一挂载：

```go
authorized := api.Group("", middleware.JWTAuth())
authorized.Use(middleware.OperationLog())
authorized.Use(middleware.PermissionAuth())
```

### 1.6 权限配置示例

底座已内置权限种子 `internal/seed/permission_seed.go`（幂等，按 `code` 写入/更新），启动时自动写入：

- 各模块的分组节点（`type=menu`，仅用于权限树展示，不参与接口匹配）；
- 每个接口的权限点（`type=api`，`method` + 相对路径，如 `POST /users`、`PUT /users/:id`）；
- 全部底座权限点会自动授予平台超级管理员角色（`tenant_id=0, code=super_admin`），保证升级后平台管理员不被拦截。

新增接口时，在 `basePermissionSeeds` 中补一条即可，无需手工建数据；普通用户角色的权限请在「角色管理 → 分配权限」中勾选。

手工新增权限记录时，字段含义如下：

| app_code | code | name | type | method | path | status |
|----------|------|------|------|--------|------|--------|
| base | user:create | 创建用户 | api | POST | /users | 1 |
| base | user:update | 编辑用户 | api | PUT | /users/:id | 1 |

然后在角色管理中把权限分配给角色，拥有该角色的用户即可访问对应接口。

---

## 2. 租户数据隔离

### 2.1 核心原则

- 每个租户只能操作自己租户下的数据。
- 超级管理员（`tenantID == 0`）可以跨租户管理所有数据。
- 按 ID 操作的 `Update`、`Delete`、`GetByID` 接口必须在 SQL 中追加 `WHERE tenant_id = ?` 条件。

### 2.2 实现模式

#### Service 层

对于带 `tenant_id` 字段的模型，相关 service 方法签名增加 `tenantID uint64` 参数：

```go
func (s UserService) Update(u *models.User, tenantID uint64) error
func (s UserService) Delete(id uint64, tenantID uint64) error
func (s UserService) GetByID(id uint64, tenantID uint64) (*models.User, error)
```

方法内部判断：

```go
db := db.DB.Model(u)
if tenantID > 0 {
    db = db.Where("tenant_id = ?", tenantID)
}
return db.Updates(updates).Error
```

- 当 `tenantID > 0`（普通租户管理员），SQL 自动带 `tenant_id = ?` 过滤。
- 当 `tenantID == 0`（超级管理员），不加过滤，可管理全量数据。

#### 写操作必须校验命中（`internal/service/helper.go`）

GORM 的 `Updates` / `Delete` 在 `WHERE` 命中 0 行时**不返回错误**，若直接 `return res.Error`，接口会提示
「保存成功 / 删除成功」而数据没有任何变化。菜单、字典、机构、消息模板这些模块的读取口径包含平台内置数据
（`tenant_id = 0`，租户可见但不可改），这个问题在租户管理员操作平台内置项时必然出现，因此：

- 更新前用 `ensureRecordExists(...)` 先做一次存在性 + 归属校验（`COUNT`）；
  **不要用 `RowsAffected` 判断更新**：MySQL 默认返回「实际变更行数」，原样保存时会得到 0，会被误判为失败。
- 删除后用 `ensureDeleteAffected(...)` 校验 `RowsAffected > 0`（软删除一定会改 `deleted_at`，因此可靠）。

#### Controller 层

从 gin 上下文获取当前租户 ID 并传入 service：

```go
func (ctl *UserController) Update(c *gin.Context) {
    var u models.User
    c.ShouldBindJSON(&u)
    u.ID = uint64(parseID(c))
    tenantID := c.GetUint64("tenantID")
    if err := ctl.service.Update(&u, tenantID); err != nil {
        response.Fail(c, err.Error())
        return
    }
    response.Ok(c, u)
}
```

### 2.3 已改造模块

以下模块的 `Update` / `Delete` / `GetByID` 已按租户隔离：

| 模块 | Service 方法 | Controller 方法 |
|------|-------------|----------------|
| User | `Update/Delete/GetByID` | `Update/Delete/Get` |
| Role | `Update/Delete/GetByID` | `Update/Delete/Get` |
| WorkflowRole | `Update/Delete/GetByID` | `Update/Delete/Get` |
| Workflow | `Update/Delete/GetByID` | `Update/Delete/Get/SaveNodes` |
| WorkflowInstance | `Delete`（按租户） | `Delete`（仅管理员）、`Detail`（发起人/参与人/管理员） |
| Menu | `Update/Delete/GetByID` | `Update/Delete` |
| Organization | `Update/Delete/GetByID` | `Update/Delete/Get` |
| AppInstance | `Update/Delete/GetByID` | `Update/Delete/List` |
| Message | `Update/Delete/GetByID` | `Delete/Get`（无 Update 接口） |
| MessageTemplate | `Update/Delete/GetByID` | `Update/Delete/Get` |
| Dict | `Update/Delete/GetByID` | `Update/Delete/Get` |
| File | `Delete`（按租户列表） | `Delete/List/Upload`（上传记录当前租户） |

> 注：Role 的 `AssignMenus`、`AssignPermissions`，User 的 `AssignRoles`、`ResetPassword`，
> WorkflowRole 的 `AssignUsers`（成员必须与角色同租户，否则报错）与 `ListUserOptions`，
> 以及各模块的 `List/Tree` 查询均已按当前租户过滤。

### 2.4 List 查询与跨租户管理

列表接口的租户口径统一为：

```go
if tenantID > 0 {
    query = query.Where("tenant_id = ?", tenantID)      // 普通租户：只能看本租户
} else if filterTenantID > 0 {
    query = query.Where("tenant_id = ?", filterTenantID) // 平台超管：按 tenantId 查询参数过滤
}                                                        // 平台超管不传 tenantId：查看全部租户
```

### 2.5 创建时的目标租户（重要）

带 `tenant_id` 的资源在创建时统一使用 `controllers.resolveTenantID(c, requested)`：

- 平台超级管理员（`tenantID == 0`）：可用请求体中的 `tenantId` 指定目标租户（0 表示平台级）；
- 普通租户用户：**忽略请求中的 `tenantId`**，强制写入自身租户，防止跨租户写入。

涉及的创建接口：`POST /users`、`POST /roles`、`POST /app-instances`、`POST /dicts`、`POST /organizations`、`POST /message-templates`、`POST /workflow-roles`、`POST /workflows`。

### 2.6 不需要租户隔离的模块

#### Tenant 租户管理

`Tenant` 模块本身即为租户管理入口，且路由已绑定 `middleware.SuperAdminOnly()`，只有平台超级管理员能访问。其职责就是管理所有租户，因此 `Update/Delete/Get` 不需要按 `tenant_id` 过滤，保持可直接操作任意租户。

#### 平台级资源

以下模型本身没有 `tenant_id` 字段，属于平台级资源，不在租户隔离范围内：

- `App`：应用定义全平台共享。
- `Permission`：权限点定义全平台共享。
- `Setting`：系统设置全平台共享。

---

## 3. 登录验证码

验证码实现与 portal / member / application 保持一致（`internal/service/captcha_service.go`）：

- **样式**：`base64Captcha.NewDriverString`，5 位数字 + 大写字母，字符集 `234679ACDEFGHJKMNPQRTUVWXY`
  （已剔除 0/O、1/I/L、2/Z、5/S、8/B 等易混字符），带空心线/粘连线/正弦线干扰；
  图片尺寸 100x300（宽:高 = 3:1），对应前端 150x50 容器 + `object-fit: cover`，避免裁切。
- **存储**：仅 Redis（key 为 `captcha:<id>`，TTL 5 分钟），已去除内存 store，多实例部署校验一致。
- **校验**：一次性（无论对错校验后立即删除）、大小写不敏感、忽畸首尾空格；`id` 或 `code` 为空直接失败。
- **接口字段**：`GET /base/api/v1/auth/captcha` 返回 `{ captcha_id, captcha_img }`；
  登录请求体使用 `captcha_id` + `captcha_code`（与 portal / member / application 一致）。
- **开关**：是否启用验证码由「系统设置 → 安全策略 → 登录验证码」控制（`settings.security.captchaEnabled`）。
  登录页先调 `GET /base/api/v1/site-info` 获取 `captchaEnabled`，关闭时不渲染验证码输入框、不提交验证码字段，后端同样不校验。
- **失败处理**：验证码错误或密码错误均计入 `login_fail:<tenantID>:<username>`，窗口与阈值由安全策略配置（默认 5 次 / 30 分钟）。

---

## 4. 安全策略（系统设置）

「系统设置 → 安全策略」写入 `base_setting`（`category = security`，值以字符串存储），由 `internal/service/settings_service.go` 统一读取并带默认值与范围限制：

| 设置项 | Key | 默认值 | 作用 |
|--------|-----|--------|------|
| 登录验证码 | `captchaEnabled` | `true` | 关闭后登录接口不再校验验证码，登录页隐藏验证码框 |
| 登录失败锁定 | `loginLock` | `true` | 是否启用连续失败锁定 |
| 最大失败次数 | `maxFailCount` | `5` | 达到该次数即锁定（范围 3~20） |
| 锁定时长(分钟) | `lockDuration` | `30` | 锁定时间，同时是失败计数 Redis key 的 TTL（范围 1~1440） |
| 密码最小长度 | `pwdMinLength` | `8` | 新增用户 / 重置密码 / 修改密码时服务端强校验（范围 6~32） |

**要点**：

- 读取入口统一为 `service.SettingsService{}.GetSecuritySettings()`，登录、验证码锁定、用户创建三个流程共用一份配置，避免口径不一。
- **密码不存在隐式默认值**：新增用户必须显式填写初始密码，`POST /users/:id/reset-password` 也必须由管理员传入新密码。
  早期版本会隐式使用 `123456`（6 位），与 `pwdMinLength`（默认 8）自相矛盾——不填密码必然报错，
  而重置密码又不做长度校验、绕过策略；现在两者统一走 `service.validatePassword`。
- 设置为空或非法时回退到默认值（数值会被 clamp 到合法区间），因此旧环境不需要预先写入任何记录。
- 前端「安全策略」表单必须与后端 key 对齐（`api/setting.ts` + `views/base/setting/index.vue` 的 `handleSaveSecurity`）。

---

## 5. 公开接口限流

实现：[internal/middleware/rate_limit.go](../internal/middleware/rate_limit.go)，基于 Redis 固定窗口（借鉴 portal 的同名中间件）。

**已挂载限流的接口**（参数见 `config.yaml` 的 `server.*_rate_limit`）：

| 接口 | 限制 | 配置项 |
|------|------|--------|
| `POST /base/api/v1/auth/login` | 10 次/分钟 | `login_rate_limit` / `login_rate_window_seconds` |
| `GET /base/api/v1/auth/captcha` | 30 次/分钟 | `captcha_rate_limit` / `captcha_rate_window_seconds` |
| `POST /base/api/v1/auth/init` | 5 次/分钟 | `init_rate_limit` / `init_rate_window_seconds` |

**要点**：

- 计数 key 为 `ratelimit:<路由模板>:<客户端IP>:<窗口段>`，**每个接口各自独立配额**；
  （portal 当前实现不含路由段，同 IP 下各接口共用一个计数，验证码刷多了会把登录额度一并吃掉，base 未采纳）
- 超限返回 `{ code: 429, message: "请求过于频繁，请稍后再试" }`。
- Redis 不可用时退化为**进程内固定窗口限流**（fail-closed），避免限流组件故障时被无限刷量。
- 客户端 IP 由 gin 的 `ClientIP()` 解析：仅当请求来自 `server.trusted_proxies` 配置的可信代理时才信任
  `X-Forwarded-For` / `X-Real-IP`，否则一律使用 `RemoteAddr`，防止伪造头绕过限流。
  部署在 Nginx 等反向代理后时，**必须**把代理真实地址填入 `trusted_proxies`，否则限流会把所有请求算到代理 IP 上。

---

## 6. 开发 checklist

新增一个需要按 ID 操作的接口时，请确认：

- [ ] 模型是否包含 `tenant_id` 字段。
- [ ] `Update`/`Delete`/`GetByID` 方法是否接收并使用了 `tenantID` 参数。
- [ ] Controller 是否从 `c.GetUint64("tenantID")` 获取当前租户 ID 并传入 service。
- [ ] 创建接口是否使用 `resolveTenantID(c, requested)` 而不是直接赋值 `c.GetUint64("tenantID")`。
- [ ] 判断「平台超级管理员」是否统一使用 `models.IsPlatformTenant(tenantID)`，而不要内联 `tenantID == 0`。
- [ ] 读取可见性是否符合约定：字典/机构/消息模板/菜单/消息为 `tenant_id = 自身 OR tenant_id = 0`，用户/角色/文件/应用实例严格等于自身。
- [ ] 列表接口：租户用户只看自己，超管默认全部并支持 `tenantId` 查询参数过滤。
- [ ] 跨租户请求是否返回"资源不存在"或 403，而不是泄露其他租户数据。
- [ ] 该接口是否已在 `internal/seed/permission_seed.go` 的 `basePermissionSeeds` 中登记。
- [ ] 若前端需要控制按钮显隐，权限点的 `code` 是否与 `userStore.can('base:xxx:yyy')` 中使用的字符串一致（超管/租户管理员由 `can()` 自动放行）。
