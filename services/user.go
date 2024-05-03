package services

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"go.uber.org/zap"
	"online-judge/dao/mysql"
	"online-judge/dao/redis"
	"online-judge/pkg"
	"online-judge/pkg/jwt"
	my_captcha "online-judge/pkg/my-captcha"
	"online-judge/pkg/resp"
	"time"
)

type UserService struct {
	UserID   int64  `form:"user_id" json:"user_id"`
	UserName string `form:"username" json:"username" validate:"required"`
	Password string `form:"password" json:"password" validate:"required"`
	Email    string `form:"email" json:"email" validate:"required"`
	Code     string `form:"code" json:"code" validate:"required"`
	//Role             bool      `form:"role" json:"role"`
	// true is Admin, false is user
}

// Register 用户注册服务
func (u *UserService) Register() (response resp.RegisterResponse) {

	// 检查数据库中是否已经有这个email
	var countEmail, countUsername int64
	err := mysql.CheckEmail(u.Email, &countEmail)
	// UID作为唯一标识，但是email和username不能重复
	switch {
	case err != nil:
		response.Code = resp.SearchDBError
		zap.L().Error("services-SearchDBError", zap.Error(err))
		return
	case countEmail > 0:
		response.Code = resp.EmailAlreadyExist
		zap.L().Error("services-" + fmt.Sprintf("Email %s aleardy exist", u.Email))
		return
	}

	// 检查数据库中是否已经有这个username
	err = mysql.CheckUsername(u.UserName, &countUsername)
	switch {
	case err != nil:
		response.Code = resp.SearchDBError
		zap.L().Error("services-SearchDBError", zap.Error(err))
		return
	case countUsername > 0:
		response.Code = resp.UsernameAlreadyExist
		zap.L().Error("services-" + fmt.Sprintf("username %s aleardy exist", u.UserName))
		return
	}

	//// 检验邮箱是否有用
	//ok := pkg.ValidateEmail(u.Email)
	//if !ok {
	//	resp.Code = InvalidateEmailFormat
	//	zap.L().Error(fmt.Sprintf("email %s invalid", u.Email))
	//	return
	//}

	//验证码获取及验证
	code, err := redis.GetVerificationCode(u.Email)
	// 验证码过期
	if errors.Is(err, fmt.Errorf("verification code expired")) {
		response.Code = resp.ExpiredVerCode
		return
	}
	// 验证码错误
	if code != u.Code {
		response.Code = resp.ErrorVerCode
		zap.L().Error("services-" + fmt.Sprintf("Error verfiction code %s:%s", u.Email, code))
		return
	}

	// 生成唯一的ID
	node, err := snowflake.NewNode(1)
	if err != nil {
		response.Code = resp.GenerateNodeError
		zap.L().Error("services-generate new node error", zap.Error(err))
		return
	}

	u.UserID = int64(node.Generate())
	// 密码加密
	u.Password, err = pkg.CryptoPwd(u.Password)
	if err != nil {
		response.Code = resp.EncryptPwdError
		zap.L().Error("services-encrypt pwd error ", zap.Error(err))
		return
	}

	uID := u.UserID
	userName := u.UserName
	passWord := u.Password
	email := u.Email
	// 插入数据库
	if err = mysql.InsertNewUser(uID, userName, passWord, email); err != nil {
		response.Code = resp.InsertNewUserError
		zap.L().Error("services-"+fmt.Sprintf("insert new user %s error", u.UserName), zap.Error(err))
		return
	}
	// 生成token
	token := jwt.GenerateToken(u.UserName)
	//if err != nil {
	//	response.Code = resp.GenerateTokenError
	//	zap.L().Error("generate token error")
	//	return
	//}
	// 设置响应的状态码和 token
	response.Code = resp.Success
	response.Token = token
	return
}

// Login 用户登录服务
func (u *UserService) Login() (response resp.RegisterResponse) {
	//验证码获取及验证
	code, err := redis.GetVerificationCode(u.Email)
	// 验证码过期
	if errors.Is(err, fmt.Errorf("verification code expired")) {
		response.Code = resp.ExpiredVerCode
		return
	}
	// 验证码错误
	if code != u.Code {
		response.Code = resp.ErrorVerCode
		zap.L().Error("services-" + fmt.Sprintf("Error verfiction code %s:%s", u.Email, code))
		return
	}
	// 检验是否有这个用户名
	var UserNameCount int64
	err = mysql.CheckUsername(u.UserName, &UserNameCount)
	if err != nil {
		response.Code = resp.SearchDBError
		zap.L().Error("services-SearchDBError", zap.Error(err))
		return
	}
	if UserNameCount == 0 {
		zap.L().Error("services-" + fmt.Sprintf("Do not have this username: %s", u.UserName))
		response.Code = resp.NotExistUsername
		return
	}

	// 检查密码是否正确
	ok, err := mysql.CheckPwd(u.UserName, u.Password)
	if err != nil {
		response.Code = resp.SearchDBError
		zap.L().Error("services-Search DB error: ", zap.Error(err))
		return
	}
	// 密码错误
	if !ok {
		response.Code = resp.ErrorPwd
		zap.L().Error("services-Password error")
		return
	}
	// 成功返回
	// 生成token
	token := jwt.GenerateToken(u.UserName)

	// 设置响应的状态码和 token
	response.Code = resp.Success
	response.Token = token
	return
}

// GetUserDetail 获取用户详细信息
func (u *UserService) GetUserDetail() (response resp.GetDetailResponse) {
	// 检验是否有这个用户ID
	var UserIDCount int64
	err := mysql.CheckUserID(u.UserID, &UserIDCount)
	if err != nil {
		response.Code = resp.SearchDBError
		zap.L().Error("services-SearchDBError", zap.Error(err))
		return
	}
	if UserIDCount == 0 {
		zap.L().Error("services-" + fmt.Sprintf("Do not have this userID: %d", u.UserID))
		response.Code = resp.NotExistUserID
		return
	}
	// 去数据库获取详细信息
	data, err := mysql.GetUserDetail(u.UserID)
	if err != nil {
		zap.L().Error("services-"+fmt.Sprintf("Db scan userID %d detail information failed", u.UserID), zap.Error(err))
		response.Code = resp.SearchDBError
		return
	}
	response.Data = data
	response.Code = resp.Success
	return
}

// DeleteUser 删除用户
func (u *UserService) DeleteUser() (response resp.DeleteUserResponse) {
	// 检验是否有这个用户ID
	var UserIDCount int64
	err := mysql.CheckUserID(u.UserID, &UserIDCount)
	if err != nil {
		response.Code = resp.SearchDBError
		zap.L().Error("services-SearchDBError", zap.Error(err))
		return
	}
	if UserIDCount == 0 {
		zap.L().Error("services-"+fmt.Sprintf("Do not have this userID: %d", u.UserID), zap.Error(err))
		response.Code = resp.NotExistUserID
		return
	}
	// 删除用户
	err = mysql.DeleteUser(u.UserID)
	if err != nil {
		zap.L().Error("services-"+fmt.Sprintf("Delete userID %d failed", u.UserID), zap.Error(err))
		response.Code = resp.DBDeleteError
		return
	}
	// 删除成功
	response.Code = resp.Success
	return
}

// UpdateUserDetail 更新用户详细信息
func (u *UserService) UpdateUserDetail() (response resp.UpdateUserDetailResponse) {
	// 检验是否有这个用户ID
	var UserIDCount int64
	err := mysql.CheckUserID(u.UserID, &UserIDCount)
	if err != nil {
		response.Code = resp.SearchDBError
		zap.L().Error("services-SearchDBError", zap.Error(err))
		return
	}
	if UserIDCount == 0 {
		zap.L().Error("services-" + fmt.Sprintf("Do not have this user_id: %d", u.UserID))
		response.Code = resp.NotExistUserID
		return
	}
	// 更新用户信息
	if u.Email != "" {
		//验证码获取及验证
		code, err := redis.GetVerificationCode(u.Email)
		// 验证码过期
		if errors.Is(err, fmt.Errorf("verification code expired")) {
			response.Code = resp.ExpiredVerCode
			return
		}
		// 验证码错误
		if code != u.Code {
			response.Code = resp.ErrorVerCode
			zap.L().Error("services-"+fmt.Sprintf("Error verfiction code %s:%s", u.Email, code), zap.Error(err))
			return
		}
	}

	if u.Password != "" {
		// 密码加密
		u.Password, err = pkg.CryptoPwd(u.Password)
		if err != nil {
			response.Code = resp.EncryptPwdError
			zap.L().Error("services-encrypt pwd error ", zap.Error(err))
			return
		}
	}
	err = mysql.UpdateUserDetail(u.UserID, u.Email, u.Password)
	if err != nil {
		zap.L().Error("services-"+fmt.Sprintf("Db update userID %d failed", u.UserID), zap.Error(err))
		response.Code = resp.SearchDBError
		return
	}
	response.Code = resp.Success
	return
}

// SendEmailCode 发送验证码接口
func SendEmailCode(userEmail string) (resCode int) {
	// 验证邮箱是否合法
	if !pkg.ValidateEmail(userEmail) {
		resCode = resp.InvalidateEmailFormat
		return
	}
	// 创建验证码
	code, ts := pkg.CreateVerificationCode()
	// 发送验证码
	err := pkg.SendCode(userEmail, code)
	if err != nil {
		resCode = resp.SendCodeError
		return
	}
	// redis保存email和验证码的键值对
	err = redis.StoreVerificationCode(userEmail, code, ts)
	if err != nil {
		resCode = resp.StoreVerCodeError
		return
	}

	resCode = resp.Success
	return
}

func SendCode(username string) (pic string, err error) {
	// 单例模式的验证码实例
	_, b64s, ans, err := my_captcha.GenerateCaptcha()

	if err != nil {
		return "", err
	}
	// 获取当前时间
	ts := time.Now().Unix()
	err = redis.StorePictureCode(username, ans, ts)
	if err != nil {
		return "", err
	}
	return b64s, nil
}

func CheckCode(username, code string) (bool, error) {
	ans, err := redis.GetPictureCode(username)
	if err != nil {
		zap.L().Error("services-CheckCode error ", zap.Error(err))
		return true, err
	}
	if ans != code {
		zap.L().Error("services-CheckCode error wrong picture code from: " + username)
		return false, nil
	}

	return true, nil
}

func GetUserID(username string) (uid int64, err error) {
	uid, err = mysql.GetUserID(username)
	if err != nil {
		zap.L().Error("services-GetUserID "+username+"err: ", zap.Error(err))
		return 0, err
	}
	return uid, nil
}
