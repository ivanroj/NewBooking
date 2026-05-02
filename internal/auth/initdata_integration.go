//go:build integration

package auth

import (
	"net/url"
	"strconv"
)

// SignedInitDataForTests builds a valid init_data string for integration tests using the official hash scheme.
func SignedInitDataForTests(botToken string, userJSON string, authDateUnix int64) (string, error) {
	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(authDateUnix, 10))
	values.Set("user", userJSON)
	values.Set("query_id", "integration-test-query")

	dc := buildDataCheckString(values)
	sig, err := computeWebAppSignature(botToken, dc)
	if err != nil {
		return "", err
	}
	values.Set("hash", sig)
	return values.Encode(), nil
}
