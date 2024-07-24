// middleware/auth.go
package middleware

import (
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

func CheckAuthenticated() gin.HandlerFunc {
    return func(c *gin.Context) {
        cookie, err := c.Cookie("token")
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
            c.Abort()
            return
        }

        token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return []byte(os.Getenv("JWT_SECRET")), nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
            c.Abort()
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            c.Abort()
            return
        }

        // Extract user ID and ensure it's a float64 (JSON numbers are float64 by default)
        userID, ok := claims["sub"].(float64)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
            c.Abort()
            return
        }

        // Set user ID as uint in the context
        c.Set("user_id", uint(userID))
        c.Next()
    }
}

