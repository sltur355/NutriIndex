package redis

import (
	"context"
	"time"
)

const jwtPrefix = "jwt."

func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

// WriteJWTToBlacklist добавляет JWT токен в черный список
func (c *Client) WriteJWTToBlacklist(ctx context.Context, jwtStr string, jwtTTL time.Duration) error {
	return c.client.Set(ctx, getJWTKey(jwtStr), "blacklisted", jwtTTL).Err()
}

// CheckJWTInBlacklist проверяет находится ли JWT токен в черном списке
func (c *Client) CheckJWTInBlacklist(ctx context.Context, jwtStr string) (bool, error) {
	result, err := c.client.Exists(ctx, getJWTKey(jwtStr)).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// RemoveJWTFromBlacklist удаляет JWT токен из черного списка (для тестов/админки)
func (c *Client) RemoveJWTFromBlacklist(ctx context.Context, jwtStr string) error {
	return c.client.Del(ctx, getJWTKey(jwtStr)).Err()
}
