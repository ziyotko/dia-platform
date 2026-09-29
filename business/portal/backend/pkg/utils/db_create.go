package utils

import (
	"context"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// CreatePreservingZeroValues 插入记录，并保证「显式传入的零值」真正落库。
//
// 背景（已在 gorm v1.31.2 的 callbacks/create.go 中确认）：
// 字段带 `gorm:"default:1"` 这类「非零默认值」标签时，GORM 会把 Go 侧零值当作「未设置」，
// 于是把 DefaultValueInterface（即 1）写进 INSERT，**并把这个值回填进传入的结构体**。
// 结果是 status（default:1）永远存不进 0 ——「新建时选禁用」被静默存成「启用」，
// 调用方读回结构体还以为是 1，后续逻辑（如模板「同类型只允许启用一个」）也跟着判断错。
// 该行为在 Create 回调内部由 `field.DefaultValueInterface != nil` 决定，
// `Select("*")` / `Omit` 都绕不过去（前者列仍被强制赋默认值，后者会直接落库默认值）。
//
// 因此这里退一步：插入前记下调用方想写的值，插入后若被库默认值改写，则按主键补写一次
// （幂等，只在「显式传零值」这一种情况下多发一条 UPDATE）。
//
// columns 传**数据库列名**，且只应传「零值代表明确意图」的字段：
// 传进来就表示该字段的零值要保留，而不再回退到库默认值。
// 例如 menu.type（default:'directory'）就**不要**传 —— 它的零值语义是「用默认值」而非「显式置空」。
//
// passthrough：解析失败时回退为普通 Create，保证调用点不会因为本函数而不可用。
func CreatePreservingZeroValues(db *gorm.DB, dest any, columns ...string) error {
	// 只支持「指针 → 结构体」这一种形态（本项目所有 Create 都是这个用法）；
	// 其余形态（切片批量插入、map）直接按原样写库。
	rv := reflect.ValueOf(dest)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return db.Create(dest).Error
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return db.Create(dest).Error
	}

	// 注意：Statement.Parse 只解析 Schema/Table，不会设置 ReflectValue，必须自己取。
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(dest); err != nil || stmt.Schema == nil {
		return db.Create(dest).Error
	}
	ctx := context.Background()

	// 插入前记下调用方想写的值（此时结构体里就是请求体/业务层算出来的原值）
	type pending struct {
		field *schema.Field
		value any
	}
	wanted := make([]pending, 0, len(columns))
	for _, column := range columns {
		field := stmt.Schema.LookUpField(column)
		if field == nil {
			continue
		}
		value, _ := field.ValueOf(ctx, rv)
		wanted = append(wanted, pending{field: field, value: value})
	}

	if err := db.Create(dest).Error; err != nil {
		return err
	}
	if len(wanted) == 0 {
		return nil
	}

	// 只补写「被库默认值改写」的字段；值没变说明 GORM 正常写入了，无需额外开销
	fixes := make(map[string]any)
	for _, want := range wanted {
		now, _ := want.field.ValueOf(ctx, rv)
		if !reflect.DeepEqual(now, want.value) {
			fixes[want.field.DBName] = want.value
		}
	}
	if len(fixes) == 0 {
		return nil
	}
	// dest 是刚插入的实例（主键已回填），Updates 会自动带上主键条件，
	// 并把补写结果同步回结构体，调用方继续读字段也能拿到正确值。
	return db.Model(dest).Updates(fixes).Error
}
