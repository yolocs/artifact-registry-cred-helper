package netrc

import "testing"

func TestUpdate(t *testing.T) {
	n := &NetRC{content: ""}

	hosts := []string{"host1.pkg.dev", "host2.pkg.dev"}
	token := "test-token"
	n.SetToken(hosts, token, false)

	expectedContent := tokenFormat(hosts[0], token) + tokenFormat(hosts[1], token)
	if n.content != expectedContent {
		t.Errorf("incorrect content:\ngot:\n%s\nwant:\n%s", n.content, expectedContent)
	}

	key := "test-key"
	n.SetJSONKey(hosts, key, false)
	expectedContent = jsonKeyFormat(hosts[0], key) + jsonKeyFormat(hosts[1], key)
	if n.content != expectedContent {
		t.Errorf("incorrect content:\ngot:\n%s\nwant:\n%s", n.content, expectedContent)
	}
}

func TestRefresh(t *testing.T) {
	n := &NetRC{content: `
machine host1.pkg.dev
login oauth2accesstoken
password old-token

machine host2.pkg.dev
login _json_key_base64
password test-key
`}

	newToken := "new-token"
	n.Refresh(newToken)

	expectedContent := `
machine host1.pkg.dev
login oauth2accesstoken
password new-token

machine host2.pkg.dev
login _json_key_base64
password test-key
`
	if n.content != expectedContent {
		t.Errorf("incorrect content:\ngot:\n%s\nwant:\n%s", n.content, expectedContent)
	}

}
