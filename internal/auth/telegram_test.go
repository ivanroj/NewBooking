package auth

import (
	"net/url"
	"testing"
)

func TestBuildAndVerifyInitDataGolden(t *testing.T) {
	t.Parallel()

	const botToken = "1234567890:ABCDEFExampleBotTokenXXXXXXXX"
	userJSON := `{"id":42424242,"username":"stu"}`

	v := url.Values{}
	v.Set("auth_date", "2000000000")
	v.Set("user", userJSON)
	v.Set("query_id", "qqq")

	dc := buildDataCheckString(v)
	hash, err := computeWebAppSignature(botToken, dc)
	if err != nil {
		t.Fatal(err)
	}
	v.Set("hash", hash)
	raw := v.Encode()

	got, err := ValidateTelegramInitData(raw, botToken, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42424242 {
		t.Fatalf("user id: want 42424242, got %d", got)
	}
}
