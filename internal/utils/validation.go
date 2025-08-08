package utils

import (
	"errors"
	"regexp"
)

// ValidateUsername 用户名验证
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("用户名至少需要3个字符")
	}
	if len(username) > 50 {
		return errors.New("用户名不能超过50个字符")
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	if !matched {
		return errors.New("用户名只能包含字母、数字、下划线和连字符")
	}

	return nil
}

// ValidateEmail 邮箱验证
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("邮箱不能为空")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("邮箱格式不正确")
	}

	return nil
}

// ValidatePassword 密码验证
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("密码至少需要6个字符")
	}
	if len(password) > 128 {
		return errors.New("密码不能超过128个字符")
	}

	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasLetter || !hasNumber {
		return errors.New("密码必须包含至少一个字母和一个数字")
	}

	return nil
}
