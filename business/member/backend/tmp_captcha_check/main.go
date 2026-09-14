package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/mojocn/base64Captcha"
)

const captchaSource = "234679ACDEFGHJKMNPQRTUVWXY"

func main() {
	driver := base64Captcha.NewDriverString(
		100, 300,
		30,
		base64Captcha.OptionShowHollowLine|base64Captcha.OptionShowSlimeLine|base64Captcha.OptionShowSineLine,
		5,
		captchaSource,
		nil,
		nil,
		nil,
	)
	c := base64Captcha.NewCaptcha(driver, base64Captcha.DefaultMemStore)
	_, b64s, answer, err := c.Generate()
	if err != nil {
		panic(err)
	}
	fmt.Printf("answer=%s len=%d\n", answer, len(answer))

	raw := b64s
	if i := strings.Index(raw, ","); i >= 0 {
		raw = raw[i+1:]
	}
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(`D:\Projects\dia-platform\captcha_check.png`, data, 0o644); err != nil {
		panic(err)
	}
}
