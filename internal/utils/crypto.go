package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var hmacSecret = []byte(os.Getenv("HMAC_SECRET"))

// EncryptCardData - временная версия без PGP
func EncryptCardData(plaintext string) (string, error) {
	// Просто возвращаем данные как есть
	return plaintext, nil
}

// DecryptCardData - временная версия без PGP
func DecryptCardData(ciphertext string) (string, error) {
	return ciphertext, nil
}

func HashCVV(cvv string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckCVV(cvvHash, cvv string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(cvvHash), []byte(cvv))
	return err == nil
}

func ComputeHMAC(data string) string {
	if len(hmacSecret) == 0 {
		hmacSecret = []byte("default-hmac-key")
	}
	h := hmac.New(sha256.New, hmacSecret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func ValidateHMAC(data, hmacValue string) bool {
	expected := ComputeHMAC(data)
	return hmac.Equal([]byte(expected), []byte(hmacValue))
}

func GenerateCardNumber() string {
	bin := "453275"
	remaining := make([]byte, 9)
	for i := 0; i < 9; i++ {
		remaining[i] = byte('0' + i%10)
	}
	numberWithoutCheck := bin + string(remaining)
	checkDigit := calculateLuhn(numberWithoutCheck)
	return numberWithoutCheck + strconv.Itoa(checkDigit)
}

func calculateLuhn(number string) int {
	var sum int
	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))
		if (len(number)-i)%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + 1
			}
		}
		sum += digit
	}
	return (10 - (sum % 10)) % 10
}

func ValidateCardNumber(number string) bool {
	if len(number) != 16 {
		return false
	}
	checkDigit, _ := strconv.Atoi(string(number[15]))
	calculated := calculateLuhn(number[:15])
	return checkDigit == calculated
}

func GenerateExpiry() string {
	expiry := time.Now().AddDate(3, 0, 0)
	return expiry.Format("0106")
}

func HashExpiry(expiry string) string {
	return ComputeHMAC(expiry)
}
