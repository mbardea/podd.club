package auth

import (
	"crypto/md5"
	"testing"
)

func TestPasswordEncode(t *testing.T) {
	salt := "_salt_"
	password := "_password_"
	hash := md5.Sum([]byte(salt + password))
	encoded := EncodePassword("md5", salt, hash[:])

	t.Logf("Encoded password: %s", encoded)
	if encoded != "md5:_salt_:a0sujt9aGyo67LeebAuWpA==" {
		t.Errorf("Password encoding failed")
	}

	verified, err := CheckPassword(password, encoded)
	if err != nil {
		t.Errorf("Password verification returned an error: %s", err)
	}
	if !verified {
		t.Errorf("Failed to check password")
	}

	verified, err = CheckPassword("wrong_password", encoded)
	if err != nil {
		t.Errorf("Password verification returned an error: %s", err)
	}
	if verified {
		t.Errorf("Password verification passed unexpectedly")
	}

	encoded2 := MakePassword("_password2_")
	t.Logf("Encoded password 2: %s", encoded2)
	verified, err = CheckPassword("_password2_", encoded2)
	if !verified {
		t.Errorf("Failed to verify password")
	}

}
