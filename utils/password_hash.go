package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pw string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}

// Fungsi untuk memeriksa apakah password sesuai dengan hash
func CheckPasswordHash(hashedPwd, plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
	return err == nil
}
