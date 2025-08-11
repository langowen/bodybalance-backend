package api

type HealthCheck struct {
	DBStatus    string
	RedisStatus string
}
