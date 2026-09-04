package utils
package utils

import (
	"strings"
	"testing"

	"github.com/mojocn/base64Captcha"
)

func TestCaptchaGlyphsCovered(t *testing.T) {
	font := base64Captcha.DefaultEmbeddedFonts.LoadFontByName("fonts/wqy-microhei.ttc")
	if font == nil {
		t.Fatal("font not loaded")
	}
	var missing []rune
	for _, r := range strings.ReplaceAll(captchaChineseSource, ",", "") {
		if font.Index(r) == 0 {
			missing = append(missing, r)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("characters not covered by wqy-microhei: %q", string(missing))
	}
}
