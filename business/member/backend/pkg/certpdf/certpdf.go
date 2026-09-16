// Package certpdf 生成会员证书 PDF（默认 A4 横向；也可按上传的 PDF 模板套打）。
// 使用纯 Go 的 go-pdf/fpdf（叠加 gofpdi 导入模板页），中文通过内嵌 TTF/OTF 字体子集实现。
package certpdf

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/phpdave11/gofpdi"
)

// 默认版式页面尺寸（A4 横向，单位 mm）与单位换算
const (
	pageW  = 297.0
	pageH  = 210.0
	ptToMM = 25.4 / 72.0
)

// Data 渲染证书所需的业务数据。
type Data struct {
	Title        string    // 证书标题（如「会员证书」或样式名称）
	OrgName      string    // 入会机构
	IssuerName   string    // 发证机构（总会；留空则用 OrgName）
	MemberName   string    // 单位名称或个人姓名
	MemberType   string    // 会员类型中文标签（如「单位会员」）
	LevelName    string    // 会员等级
	CertNo       string    // 证书编号
	IssuedAt     time.Time // 颁发日期
	ExpireAt     time.Time // 有效期至
	TemplateName string    // 样式名称（可选，仅作展示兜底）
}

// Generate 生成证书 PDF 字节。fontPath 为空时自动查找系统常见中文字体。
func Generate(d Data, fontPath string) ([]byte, error) {
	path, err := resolveFont(fontPath)
	if err != nil {
		return nil, err
	}

	pdf := fpdf.NewCustom(&fpdf.InitType{UnitStr: "mm", Size: fpdf.SizeType{Wd: pageW, Ht: pageH}})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	pdf.AddUTF8Font("cert", "", path)
	if pdf.Error() != nil {
		return nil, fmt.Errorf("加载证书字体失败（仅支持 TTF/OTF）：%w", pdf.Error())
	}

	title := strings.TrimSpace(d.Title)
	if title == "" {
		title = strings.TrimSpace(d.TemplateName)
	}
	if title == "" {
		title = "会员证书"
	}
	issuer := strings.TrimSpace(d.IssuerName)
	if issuer == "" {
		issuer = strings.TrimSpace(d.OrgName)
	}

	// ---- 双线边框 ----
	pdf.SetDrawColor(11, 47, 107) // #0b2f6b
	pdf.SetLineWidth(1.4)
	pdf.Rect(10, 10, pageW-20, pageH-20, "D")
	pdf.SetDrawColor(198, 158, 74) // #c69e4a
	pdf.SetLineWidth(0.4)
	pdf.Rect(14, 14, pageW-28, pageH-28, "D")

	// ---- 标题 ----
	pdf.SetTextColor(11, 47, 107)
	pdf.SetFont("cert", "", 30)
	pdf.SetXY(0, 32)
	pdf.CellFormat(pageW, 16, title, "", 0, "C", false, 0, "")

	// ---- 入会机构（副标题）----
	if d.OrgName != "" {
		pdf.SetFont("cert", "", 12)
		pdf.SetTextColor(96, 96, 96)
		pdf.SetXY(0, 52)
		pdf.CellFormat(pageW, 8, d.OrgName, "", 0, "C", false, 0, "")
	}

	// ---- 金色分隔线 ----
	pdf.SetDrawColor(198, 158, 74)
	pdf.SetLineWidth(0.6)
	pdf.Line(pageW/2-45, 64, pageW/2+45, 64)

	// ---- 信息行 ----
	rows := infoRows(d)
	drawInfoRows(pdf, rows, pageW, rowsTopCentered(pageH, len(rows)))

	// ---- 页脚说明 + 发证机构 ----
	pdf.SetFont("cert", "", 9)
	pdf.SetTextColor(140, 140, 140)
	footer := fmt.Sprintf("本证书由 %s 颁发，证书编号 %s，可在会员系统核验。", fallback(issuer, "-"), fallback(d.CertNo, "-"))
	pdf.SetXY(15, pageH-24)
	pdf.CellFormat(pageW-30, 6, footer, "", 0, "C", false, 0, "")

	if issuer != "" {
		pdf.SetFont("cert", "", 11)
		pdf.SetTextColor(11, 47, 107)
		pdf.SetXY(pageW-15-90, pageH-45)
		pdf.CellFormat(90, 8, issuer+"（盖章）", "", 0, "R", false, 0, "")
	}

	return outputBytes(pdf)
}

// infoRows 证书上的信息行（标签, 值）。
func infoRows(d Data) [][2]string {
	return [][2]string{
		{"会员名称", fallback(d.MemberName, "-")},
		{"会员类型", fallback(d.MemberType, "-")},
		{"会员等级", fallback(d.LevelName, "-")},
		{"证书编号", fallback(d.CertNo, "-")},
		{"有效期至", formatDate(d.ExpireAt)},
		{"颁发日期", formatDate(d.IssuedAt)},
	}
}

// drawInfoRows 居中绘制信息行（标签右对齐 + 值左对齐）；默认版式与模板套打共用。
func drawInfoRows(pdf *fpdf.Fpdf, rows [][2]string, pageWidth, topY float64) {
	const (
		labelW = 30.0
		valueW = 130.0
		rowH   = 10.5
	)
	labelX := pageWidth/2 - 70
	valueX := labelX + labelW + 6
	for i, r := range rows {
		y := topY + float64(i)*rowH
		pdf.SetFont("cert", "", 12)
		pdf.SetTextColor(122, 122, 122)
		pdf.SetXY(labelX, y)
		pdf.CellFormat(labelW, 8, r[0], "", 0, "R", false, 0, "")

		pdf.SetFont("cert", "", 13)
		pdf.SetTextColor(34, 34, 34)
		pdf.SetXY(valueX, y)
		pdf.CellFormat(valueW, 8, r[1], "", 0, "L", false, 0, "")
	}
}

// rowsTopCentered 信息块在给定页高下垂直居中。
func rowsTopCentered(pageHeight float64, rowCount int) float64 {
	const rowH = 10.5
	top := (pageHeight - float64(rowCount)*rowH) / 2
	if top < 20 {
		top = 20
	}
	return top
}

// outputBytes 输出 PDF 字节并检查延迟错误。
func outputBytes(pdf *fpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	if err := pdf.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateWithTemplate 以模板 PDF 第 1 页为底图套打会员数据。
// 页面尺寸直接取模板的 MediaBox（不强制 A4），只叠加信息行（边框/标题由模板提供）。
// gofpdi 内部失败会 panic，这里统一 recover 转成错误，避免拖垮调用方。
func GenerateWithTemplate(templatePath string, d Data, fontPath string) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			out, err = nil, fmt.Errorf("解析证书模板失败：%v", r)
		}
	}()

	path, err := resolveFont(fontPath)
	if err != nil {
		return nil, err
	}

	imp := gofpdi.NewImporter()
	imp.SetSourceFile(templatePath)

	sizes := imp.GetPageSizes()
	box, ok := sizes[1]["/MediaBox"]
	if !ok || box["w"] <= 0 || box["h"] <= 0 {
		return nil, errors.New("证书模板第 1 页缺少有效的 MediaBox")
	}
	wMM := box["w"] * ptToMM
	hMM := box["h"] * ptToMM

	pdf := fpdf.NewCustom(&fpdf.InitType{UnitStr: "mm", Size: fpdf.SizeType{Wd: wMM, Ht: hMM}})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	pdf.AddUTF8Font("cert", "", path)
	if e := pdf.Error(); e != nil {
		return nil, fmt.Errorf("加载证书字体失败（仅支持 TTF/OTF）：%w", e)
	}

	// 导入模板第 1 页，作为铺满整页的背景 XObject。
	// 注意顺序：必须先 PutFormXobjectsUnordered（gofpdi 在此真正构建并记录导入对象），
	// 再取对象内容/引用位置/模板名映射，否则 XObject 资源指向空对象（渲染为空白）。
	tplID := imp.ImportPage(1, "/MediaBox")
	tplNameToHash := imp.PutFormXobjectsUnordered()
	pdf.ImportObjects(imp.GetImportedObjectsUnordered())
	pdf.ImportObjPos(imp.GetImportedObjHashPos())
	pdf.ImportTemplates(tplNameToHash)
	// fpdf 的 UseImportedTemplate 参数顺序为 (scaleX, scaleY, tX, tY)，
	// 与 gofpdi UseTemplate 的返回值顺序一致，切勿打乱（打乱会得到退化矩阵，渲染为空白页）
	name, scaleX, scaleY, tx, ty := imp.UseTemplate(tplID, 0, 0, wMM, hMM)
	if name == "" {
		return nil, errors.New("证书模板导入失败")
	}
	pdf.UseImportedTemplate(name, scaleX, scaleY, tx, ty)

	// 套打数据行（不画边框/标题/页脚，模板自带）
	rows := infoRows(d)
	drawInfoRows(pdf, rows, wMM, rowsTopCentered(hMM, len(rows)))

	return outputBytes(pdf)
}

// FindFont 查找系统中可用的中文字体（TTF/OTF，按平台常见路径）。
func FindFont() (string, error) {
	for _, p := range fontCandidates() {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p, nil
		}
	}
	return "", errors.New("未找到可用的中文证书字体（TTF/OTF），请配置 certificate.font_path 或设置环境变量 MEMBER_CERT_FONT")
}

// resolveFont 校验显式指定的字体路径，未指定时自动查找。
func resolveFont(fontPath string) (string, error) {
	path := strings.TrimSpace(fontPath)
	if path == "" {
		return FindFont()
	}
	if fi, err := os.Stat(path); err != nil || fi.IsDir() {
		return "", fmt.Errorf("证书字体文件不可用：%s", path)
	}
	if strings.EqualFold(filepath.Ext(path), ".ttc") {
		return "", fmt.Errorf("证书字体不支持 TTC 字体集合（%s），请改用 TTF/OTF", filepath.Base(path))
	}
	return path, nil
}

// fontCandidates 按平台返回候选字体路径（fpdf 的 UTF8 字体仅支持 TTF/OTF）。
func fontCandidates() []string {
	if v := strings.TrimSpace(os.Getenv("MEMBER_CERT_FONT")); v != "" {
		return []string{v}
	}
	switch runtime.GOOS {
	case "windows":
		root := os.Getenv("WINDIR")
		if root == "" {
			root = `C:\Windows`
		}
		return []string{
			filepath.Join(root, "Fonts", "simhei.ttf"), // 中易黑体
			filepath.Join(root, "Fonts", "Deng.ttf"),   // 等线
			filepath.Join(root, "Fonts", "simkai.ttf"), // 楷体
			filepath.Join(root, "Fonts", "simfang.ttf"),
		}
	case "darwin":
		return []string{
			"/Library/Fonts/Arial Unicode.ttf",
			"/System/Library/Fonts/Supplemental/Songti.ttf",
			"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
		}
	default:
		return []string{
			"/usr/share/fonts/opentype/noto/NotoSansCJKsc-Regular.otf",
			"/usr/share/fonts/truetype/wqy/wqy-microhei.ttf",
			"/usr/share/fonts/truetype/arphic/uming.ttf",
			"/usr/share/fonts/truetype/droid/DroidSansFallbackFull.ttf",
		}
	}
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02")
}

func fallback(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
