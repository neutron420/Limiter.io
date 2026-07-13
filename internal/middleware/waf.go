package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"limiter.io/internal/dto"
)

type WafConfig struct {
	BlockedCIDRs     []*net.IPNet
	BlockedCountries map[string]bool
}

func WafMiddleware(cfg WafConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIPStr := c.ClientIP()
		clientIP := net.ParseIP(clientIPStr)

		// 1. Evaluate IP CIDR blacklists
		if clientIP != nil {
			for _, ipNet := range cfg.BlockedCIDRs {
				if ipNet.Contains(clientIP) {
					c.JSON(http.StatusForbidden, dto.ErrorResponse{
						Error: "Access Denied: Your IP is blacklisted by WAF rules",
					})
					c.Abort()
					return
				}
			}
		}

		// 2. Evaluate Country restrictions
		countryCode := strings.ToUpper(c.GetHeader("CF-IPCountry"))
		if countryCode == "" {
			countryCode = strings.ToUpper(c.GetHeader("X-Country-Code"))
		}

		if countryCode != "" && cfg.BlockedCountries[countryCode] {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: "Access Denied: Traffic restricted from this country",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
