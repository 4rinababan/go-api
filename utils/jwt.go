package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("SECRET_KEY_YANG_AMAN") // sebaiknya pakai os.Getenv()
var ErrTokenExpired = errors.New("token expired")

type Claims struct {
	UserID   string `json:"user_id"`
	Phone    string `json:"phone"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	Regency  string `json:"regency"`
	District string `json:"district"`
	Lang     string `json:"lang"`
	Lat      string `json:"lat"`
	PhotoUrl string `json:"photo_url"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
	jwt.RegisteredClaims
}

// GenerateJWT membuat token JWT dengan semua info user
func GenerateJWT(userID, phone, name, email, address, regency, district, lang, lat, photoUrl, role string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 1 hari
	claims := &Claims{
		UserID:   userID,
		Phone:    phone,
		Name:     name,
		Email:    email,
		Address:  address,
		Regency:  regency,
		District: district,
		Lang:     lang,
		Lat:      lat,
		PhotoUrl: photoUrl,
		Role:     role,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateJWT validasi & ambil klaim dari token
func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	// parsing token dengan Claims
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		// cek apakah error karena expired
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, err
	}

	// cek Exp secara manual (opsional)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return claims, nil
}
