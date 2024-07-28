package middleware

import (
    "net/http"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

func CheckAuthenticated() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString, err := c.Cookie("token")
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
            c.Abort()
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return []byte(os.Getenv("JWT_SECRET")), nil
        })

        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
            c.Abort()
            return
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            if exp, ok := claims["exp"].(float64); ok {
                if float64(time.Now().Unix()) > exp {
                    c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
                    c.Abort()
                    return
                }
            } else {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid expiration claim"})
                c.Abort()
                return
            }

            userID, ok := claims["sub"].(float64)
            if !ok {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
                c.Abort()
                return
            }

            c.Set("user_id", uint(userID))
            c.Next()
        } else {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            c.Abort()
        }
    }
}
