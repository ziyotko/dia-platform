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

中间件对以下情况直接放行：

1. **超级管理员**：`tenantID == 0` 的用户直接放行。
2. **白名单接口**：用户基础信息、菜单、权限、改密、仪表盘、我的应用等基础接口。
   - `GET /base/api/v1/auth/info`
   - `GET /base/api/v1/auth/menus`
   - `GET /base/api/v1/auth/permissions`
   - `POST /base/api/v1/auth/change-password`
   - `GET /base/api/v1/dashboard/stats`
   - `GET /base/api/v1/app-instances/my`
3. **空权限表兼容**：若用户没有任何角色关联的有效权限记录，直接放行，避免新系统或历史数据导致全部接口不可用。

### 1.5 挂载方式

在 [internal/routes/routes.go](../internal/routes/routes.go) 中，对已登录接口统一挂载：

```go
authorized := api.Group("", middleware.JWTAuth())
authorized.Use(middleware.OperationLog())
authorized.Use(middleware.PermissionAuth())
```

### 1.6 权限配置示例

在 `base_permission` 表中新增一条记录：

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
- 按 ID 操作的 `Update`、`Delete`、`Get` 接口必须在 SQL 中追加 `WHERE tenant_id = ?` 条件。

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

| 模块 | Service 方法 | Controller 方法 |
|------|-------------|----------------|
| User | `Update/Delete/GetByID` | `Update/Delete/Get` |
| Role | `Update/Delete` | `Update/Delete` |
| Menu | `Update/Delete` | `Update/Delete` |
| Organization | `Update/Delete` | `Update/Delete` |
| AppInstance | `Update/Delete` | `Update/Delete`（Create 也已强制当前租户） |
| Message | `Update/Delete/GetByID` | `Delete/Get`（无 Update 接口） |

### 2.4 List/Tree 查询

列表和树形接口本身已按 `tenantID` 过滤：

```go
query := db.DB.Model(&models.User{}).Where("tenant_id = ?", tenantID)
```

### 2.5 待改造模块

以下模块仍存在按 ID 操作不校验租户归属的风险，建议按相同模式补齐：

- `MessageTemplate`：Update/Delete/GetByID
- `Tenant`：Update/Delete/GetByID（当前仅 SuperAdminOnly 控制路由，但 Get 仍可查任意租户）

### 2.6 平台级资源说明

以下模型本身没有 `tenant_id` 字段，属于平台级资源，不在租户隔离范围内：

- `App`：应用定义全平台共享。
- `Permission`：权限点定义全平台共享。

---

## 3. 开发checklist

新增一个需要按 ID 操作的接口时，请确认：

- [ ] 模型是否包含 `tenant_id` 字段。
- [ ] `Update`/`Delete`/`GetByID` 方法是否接收并使用了 `tenantID` 参数。
- [ ] Controller 是否从 `c.GetUint64("tenantID")` 获取当前租户 ID 并传入 service。
- [ ] 超级管理员（`tenantID == 0`）行为是否符合预期。
- [ ] 跨租户请求是否返回"资源不存在"或 403，而不是泄露其他租户数据。
- [ ] 该接口是否已在 `base_permission` 表中配置，并分配给需要访问的角色。
