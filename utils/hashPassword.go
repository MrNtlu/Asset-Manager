package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hashedPassword)
}

func CheckPassword(hashedPassword, databasePassword []byte) error {
	err := bcrypt.CompareHashAndPassword(hashedPassword, databasePassword)

	return err
}
