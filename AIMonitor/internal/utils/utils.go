package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// PasswordUtils 密码工具
type PasswordUtils struct{}

// HashPassword 加密密码
func (p *PasswordUtils) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func (p *PasswordUtils) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomPassword 生成随机密码
func (p *PasswordUtils) GenerateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// TimeUtils 时间工具
type TimeUtils struct{}

// FormatTime 格式化时间
func (t *TimeUtils) FormatTime(time time.Time, layout string) string {
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return time.Format(layout)
}

// ParseTime 解析时间字符串
func (t *TimeUtils) ParseTime(timeStr, layout string) (time.Time, error) {
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return time.Parse(layout, timeStr)
}

// GetTimeRange 获取时间范围
func (t *TimeUtils) GetTimeRange(rangeType string) (start, end time.Time) {
	now := time.Now()
	switch rangeType {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start.Add(24 * time.Hour)
	case "yesterday":
		start = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
		end = start.Add(24 * time.Hour)
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
		end = start.Add(7 * 24 * time.Hour)
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 1, 0)
	case "year":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(1, 0, 0)
	default:
		start = now.Add(-24 * time.Hour)
		end = now
	}
	return
}

// IsExpired 检查是否过期
func (t *TimeUtils) IsExpired(expireTime time.Time) bool {
	return time.Now().After(expireTime)
}

// StringUtils 字符串工具
type StringUtils struct{}

// IsEmpty 检查字符串是否为空
func (s *StringUtils) IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// Truncate 截断字符串
func (s *StringUtils) Truncate(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length] + "..."
}

// CamelToSnake 驼峰转下划线
func (s *StringUtils) CamelToSnake(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// SnakeToCamel 下划线转驼峰
func (s *StringUtils) SnakeToCamel(str string) string {
	parts := strings.Split(str, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// GenerateRandomString 生成随机字符串
func (s *StringUtils) GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// MaskSensitive 掩码敏感信息
func (s *StringUtils) MaskSensitive(str string, start, end int) string {
	if len(str) <= start+end {
		return strings.Repeat("*", len(str))
	}
	return str[:start] + strings.Repeat("*", len(str)-start-end) + str[len(str)-end:]
}

// FileUtils 文件工具
type FileUtils struct{}

// Exists 检查文件是否存在
func (f *FileUtils) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CreateDir 创建目录
func (f *FileUtils) CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// WriteFile 写入文件
func (f *FileUtils) WriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := f.CreateDir(dir); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ReadFile 读取文件
func (f *FileUtils) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// GetFileSize 获取文件大小
func (f *FileUtils) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileExt 获取文件扩展名
func (f *FileUtils) GetFileExt(path string) string {
	return filepath.Ext(path)
}

// GetFileName 获取文件名（不含扩展名）
func (f *FileUtils) GetFileName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// NetworkUtils 网络工具
type NetworkUtils struct{}

// IsValidIP 验证IP地址
func (n *NetworkUtils) IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidPort 验证端口号
func (n *NetworkUtils) IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}

// IsValidURL 验证URL
func (n *NetworkUtils) IsValidURL(urlStr string) bool {
	_, err := url.Parse(urlStr)
	return err == nil
}

// GetClientIP 获取客户端IP
func (n *NetworkUtils) GetClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查X-Real-IP头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// IsLocalIP 检查是否为本地IP
func (n *NetworkUtils) IsLocalIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.IsLoopback() || parsedIP.IsPrivate()
}

// ValidationUtils 验证工具
type ValidationUtils struct{}

// IsValidEmail 验证邮箱
func (v *ValidationUtils) IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsValidPhone 验证手机号（中国）
func (v *ValidationUtils) IsValidPhone(phone string) bool {
	regex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return regex.MatchString(phone)
}

// IsValidPassword 验证密码强度
func (v *ValidationUtils) IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// IsValidUsername 验证用户名
func (v *ValidationUtils) IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	regex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return regex.MatchString(username)
}

// CryptoUtils 加密工具
type CryptoUtils struct{}

// MD5 计算MD5哈希
func (c *CryptoUtils) MD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA256 计算SHA256哈希
func (c *CryptoUtils) SHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomBytes 生成随机字节
func (c *CryptoUtils) GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	return bytes, err
}

// GenerateRandomHex 生成随机十六进制字符串
func (c *CryptoUtils) GenerateRandomHex(length int) (string, error) {
	bytes, err := c.GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// JSONUtils JSON工具
type JSONUtils struct{}

// ToJSON 转换为JSON字符串
func (j *JSONUtils) ToJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON 从JSON字符串解析
func (j *JSONUtils) FromJSON(jsonStr string, v interface{}) error {
	return json.Unmarshal([]byte(jsonStr), v)
}

// ToJSONIndent 转换为格式化的JSON字符串
func (j *JSONUtils) ToJSONIndent(v interface{}) (string, error) {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// IsValidJSON 验证JSON格式
func (j *JSONUtils) IsValidJSON(jsonStr string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// ConvertUtils 转换工具
type ConvertUtils struct{}

// StringToInt 字符串转整数
func (c *ConvertUtils) StringToInt(str string) (int, error) {
	return strconv.Atoi(str)
}

// StringToInt64 字符串转64位整数
func (c *ConvertUtils) StringToInt64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

// StringToFloat64 字符串转64位浮点数
func (c *ConvertUtils) StringToFloat64(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

// StringToBool 字符串转布尔值
func (c *ConvertUtils) StringToBool(str string) (bool, error) {
	return strconv.ParseBool(str)
}

// IntToString 整数转字符串
func (c *ConvertUtils) IntToString(i int) string {
	return strconv.Itoa(i)
}

// Int64ToString 64位整数转字符串
func (c *ConvertUtils) Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Float64ToString 64位浮点数转字符串
func (c *ConvertUtils) Float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// BoolToString 布尔值转字符串
func (c *ConvertUtils) BoolToString(b bool) string {
	return strconv.FormatBool(b)
}

// ReflectUtils 反射工具
type ReflectUtils struct{}

// GetStructFields 获取结构体字段
func (r *ReflectUtils) GetStructFields(v interface{}) []string {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	var fields []string
	for i := 0; i < val.NumField(); i++ {
		fields = append(fields, val.Type().Field(i).Name)
	}
	return fields
}

// GetFieldValue 获取字段值
func (r *ReflectUtils) GetFieldValue(v interface{}, fieldName string) interface{} {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	fieldVal := val.FieldByName(fieldName)
	if !fieldVal.IsValid() {
		return nil
	}

	return fieldVal.Interface()
}

// SetFieldValue 设置字段值
func (r *ReflectUtils) SetFieldValue(v interface{}, fieldName string, value interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("v must be a pointer to struct")
	}

	val = val.Elem()
	fieldVal := val.FieldByName(fieldName)
	if !fieldVal.IsValid() {
		return fmt.Errorf("field %s not found", fieldName)
	}

	if !fieldVal.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}

	fieldVal.Set(reflect.ValueOf(value))
	return nil
}

// 全局工具实例
var (
	Password   = &PasswordUtils{}
	Time       = &TimeUtils{}
	String     = &StringUtils{}
	File       = &FileUtils{}
	Network    = &NetworkUtils{}
	Validation = &ValidationUtils{}
	Crypto     = &CryptoUtils{}
	JSON       = &JSONUtils{}
	Convert    = &ConvertUtils{}
	Reflect    = &ReflectUtils{}
)