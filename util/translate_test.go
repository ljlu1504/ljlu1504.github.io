package util

import "testing"

func TestTranslateCh2En(t *testing.T) {
	if en := TranslateCh2En("中国"); en != "China" {
		t.Fatal("翻译错误")
	}
	RedisClient.Close()
}

func TestTranslateEn2Ch(t *testing.T) {
	if en := TranslateEn2Ch("England"); en != "英格兰" {
		t.Fatal("翻译错误")
	}
	RedisClient.Close()
}
