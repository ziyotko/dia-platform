package utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

const creditCodeCharset = "0123456789ABCDEFGHJKLMNPQRTUWXY"

var (
	mobileRe     = regexp.MustCompile(`^1[3-9]\d{9}$`)
	phoneRe      = regexp.MustCompile(`^(1[3-9]\d{9}|0\d{2,3}-?\d{7,8})$`)
	emailRe      = regexp.MustCompile(`^[\w.+-]+@[\w-]+(\.[\w-]+)+$`)
	creditCodeRe = regexp.MustCompile(`^[0-9A-HJ-NPQRTUWXY]{2}[0-9]{6}[0-9A-HJ-NPQRTUWXY]{10}$`)
	idCardRe     = regexp.MustCompile(`^\d{17}[\dX]$`)
)

// IsValidMobile 校验11位手机号
func IsValidMobile(v string) bool { return mobileRe.MatchString(v) }

// IsValidPhone 校验手机号或固定电话（固定电话可带区号与连字符）
func IsValidPhone(v string) bool { return phoneRe.MatchString(v) }

// IsValidEmail 校验邮箱格式（最长64位）
func IsValidEmail(v string) bool { return len(v) <= 64 && emailRe.MatchString(v) }

// IsValidCreditCode 校验统一社会信用代码（GB 32100-2015，18位含校验码），与前端一致
func IsValidCreditCode(v string) bool {
	code := strings.ToUpper(strings.TrimSpace(v))
	if !creditCodeRe.MatchString(code) {
		return false
	}
	weights := []int{1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28}
	sum := 0
	for i := 0; i < 17; i++ {
		idx := strings.IndexByte(creditCodeCharset, code[i])
		if idx < 0 {
			return false
		}
		sum += idx * weights[i]
	}
	return creditCodeCharset[(31-sum%31)%31] == code[17]
}

// IsValidIDCard 校验18位身份证号（出生日期 + 校验码），与前端一致
func IsValidIDCard(v string) bool {
	id := strings.ToUpper(strings.TrimSpace(v))
	if !idCardRe.MatchString(id) {
		return false
	}
	year, _ := strconv.Atoi(id[6:10])
	month, _ := strconv.Atoi(id[10:12])
	day, _ := strconv.Atoi(id[12:14])
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if t.Year() != year || int(t.Month()) != month || t.Day() != day {
		return false
	}
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	const codes = "10X98765432"
	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(id[i]-'0') * weights[i]
	}
	return codes[sum%11] == id[17]
}
