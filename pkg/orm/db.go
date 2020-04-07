package orm

// Options 数据库配置
type Options struct {
	Driver       string `json:"driver"`
	Host         string `json:"host"`
	DBName       string `json:"dbName"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DataSource   string `json:"dataSource"`
	EnableLog    bool   `json:"enable_log"`
	LogLevel     int    `json:"logLevel"`
	MaxIdleConns int    `json:"maxIdleConns"`
	MaxOpenConns int    `json:"maxOpenConns"`
}
