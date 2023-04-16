package aozoraconv

import (
	"testing"
)

func TestEscaperRuby(t *testing.T) {
	t.Run("<< ruby >> only", func(tt *testing.T) {
		e := newRubyEscaper()
		out, ok := e.Escape(`田住生《たずみせい》`)
		if ok != true {
			tt.Errorf("always true")
		}
		if out != "田住生" {
			tt.Errorf("escape << and >>: actual=%s", out)
		}
	})
	t.Run("text | text << ruby >>", func(tt *testing.T) {
		e := newRubyEscaper()
		out, ok := e.Escape(`晩｜停車場《ステーション》`)
		if ok != true {
			tt.Errorf("always true")
		}
		if out != "晩停車場" {
			tt.Errorf("escape | and << and >>: actual=%s", out)
		}
	})
	t.Run("multiple", func(tt *testing.T) {
		e := newRubyEscaper()
		out, ok := e.Escape(`下宿屋は二階中を開《あけ》ひろげて蚊帳《かや》や蒲団《ふとん》を乾して居る`)
		if ok != true {
			tt.Errorf("always true")
		}
		if out != "下宿屋は二階中を開ひろげて蚊帳や蒲団を乾して居る" {
			tt.Errorf("multiple actual=%s", out)
		}
	})
	t.Run("no ruby", func(tt *testing.T) {
		e := newRubyEscaper()
		out, ok := e.Escape(`停車場`)
		if ok != true {
			tt.Errorf("always true")
		}
		if out != "停車場" {
			tt.Errorf("passthrough actual=%s", out)
		}
	})
}

func TestEscaperAnnotation(t *testing.T) {
	t.Run("[# annote] text [# annote]", func(tt *testing.T) {
		e := newAnnotationEscaper()
		out, ok := e.Escape(`［＃７字下げ］二［＃「二」は中見出し］`)
		if ok != true {
			tt.Errorf("always true")
		}
		if out != "二" {
			tt.Errorf("escape [#] actual=%s", out)
		}
	})
	t.Run("[#annote] text", func(tt *testing.T) {
		e := newAnnotationEscaper()
		out, ok := e.Escape(`［＃地から２字上げ］（明治四十年十一月）`)
		if ok != true {
			tt.Errorf("always true")
		}
		if out != "（明治四十年十一月）" {
			tt.Errorf("escape [#] actual=%s", out)
		}
	})
}

func TestEscaperRepeatTwo(t *testing.T) {
	t.Run("single", func(tt *testing.T) {
		e := newRepeatTwoEscaper()
		out, ok := e.Escape(`頭をフラ／＼`)
		if ok != true {
			tt.Errorf("always true")
		}
		if out != "頭をフラフラ" {
			tt.Errorf("escape repeat2 actual=%s", out)
		}
	})
}

var (
	testHeader1 = `茗荷畠
眞山青果

-------------------------------------------------------
【テキスト中に現れる記号について】

《》：ルビ
（例）田住生《たずみせい》

｜：ルビの付く文字列の始まりを特定する記号
（例）晩｜停車場《ステーション》

［＃］：入力者注　主に外字の説明や、傍点の位置の指定
　　　（数字は、JIS X 0213の面区点番号またはUnicode、底本のページと行数）
（例）※［＃「足へん＋宛」、第3水準1-92-36］

／＼：二倍の踊り字（「く」を縦に長くしたような形の繰り返し記号）
（例）フラ／＼
-------------------------------------------------------
`
	expectHeader1 = `茗荷畠
眞山青果
`
)

func TestEscaperHeader(t *testing.T) {
	t.Run("header1", func(tt *testing.T) {
		e := newHeaderEscaper()
		out, ok := e.Escape(testHeader1)
		if ok != true {
			tt.Errorf("match")
		}
		if out != expectHeader1 {
			tt.Errorf("actual=%s", out)
		}
	})
}
