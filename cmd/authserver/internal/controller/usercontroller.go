package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mangohow/easygin"
	"github.com/mangohow/imchat/cmd/authserver/internal/log"
	"github.com/mangohow/imchat/cmd/authserver/internal/resultcode"
	"github.com/mangohow/imchat/cmd/authserver/internal/service"
	"github.com/mangohow/imchat/pkg/model"
	"github.com/mangohow/imchat/pkg/utils"
	"github.com/sirupsen/logrus"
)

/*
* @Author: mgh
* @Date: 2022/2/23 21:29
* @Desc:
 */

type UserController struct {
	userService *service.UserService
	logger      *logrus.Logger
}


func NewUserController() *UserController {
	return &UserController{
		userService: service.NewUserService(),
		logger:      log.Logger(),
	}
}

func init() {
	easygin.SetCodeMessager(resultcode.MessageFunc)
}

// GetPhoneVerificationCode 获取用户验证码
// GET /api/phoneCode param: phone number
func (c *UserController) GetPhoneVerificationCode(phone string) *easygin.Result {
	c.logger.Debug("phone:", phone)
	if ok := utils.ValidatePhoneNumber(phone); !ok {
		return easygin.Fail(resultcode.PhoneInvalid)
	}

	code, err := c.userService.GetPhoneVerificationCode(phone)
	if err != nil {
		if err == service.CodeGetFrequentError {
			return easygin.Fail(resultcode.CodeGetFrequently)
		}
		c.logger.Errorf("get phone code error:%v", err)
		return easygin.Fail(resultcode.ServerException)
	}

	return easygin.Ok(code)
}

// Register 用户注册
// POST /api/userRegister json: {username, password, phone, code}
func (c *UserController) Register(user *model.UserRegister) *easygin.Result {
	c.logger.Debug("user:", user)
	// 验证参数
	err := validate.Struct(user)
	if err != nil {
		c.logger.Errorf("validate failed:%v", err)
		return easygin.Fail(resultcode.UserInfoInvalid)
	}

	err = c.userService.CreateUser(user)
	if err != nil {
		switch err {
		case service.CodeInvalidError:
			return easygin.Fail(resultcode.PhoneCodeInvalid)
		case service.PhoneHadBeenRegisteredError:
			return easygin.Fail(resultcode.PhoneHadBeenRegistered)
		case service.UsernameUnavailable:
			return easygin.Fail(resultcode.UsernameUnavailable)
		}

		return easygin.Error(http.StatusInternalServerError, -1)
	}

	return easygin.Ok(nil)
}


// Login 用户登录认证
// GET /api/login
func (c *UserController) Login(user *model.UserLogin) *easygin.Result {
	// 参数验证
	err := validate.Struct(user)
	if err != nil {
		c.logger.Errorf("validate failed:%v", err)
		return easygin.Fail(resultcode.UserInfoInvalid)
	}

	c.logger.Debugf("username:%s, password:%s", user.Username, user.Password)

	u, ui, token, err := c.userService.CheckUserLogin(user)
	if err != nil {
		switch err {
		case service.UserBanned:
			return easygin.Fail(resultcode.UserBanned)
		case service.UserDestroy:
			return easygin.Fail(resultcode.UserDestroyed)
		}


		return easygin.Fail(resultcode.UsernameOrPasswordInvalid)
	}

	u.Password = ""
	return easygin.Ok(map[string]interface{}{
		"user": u,
		"userinfo": ui,
		"token": token,
	})
}


// GetUserInfo 获取其它用户信息
// GET /api/userinfo
func (c *UserController) GetUserInfo(id int64) *easygin.Result {
	if id == 0 {
		return easygin.Fail(resultcode.ParamInvalid)
	}
	userinfo := c.userService.GetUserInfo(id)
	userinfo.PhoneNum = ""

	return easygin.Ok(userinfo)
}

// GetSelfInfo 获取自己信息
// GET /api/selfinfo
func (c *UserController) GetSelfInfo(ctx *gin.Context) *easygin.Result {
	value, exists := ctx.Get("id")
	if !exists {
		return easygin.Error(http.StatusInternalServerError, -1)
	}

	id := value.(int64)
	userinfo := c.userService.GetUserInfo(id)
	if userinfo == nil {
		return easygin.Fail(resultcode.QueryFailed)
	}

	return easygin.Ok(userinfo)
}
