package sign

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
)

// Option option
type Option func(*Sign)

// WithSignMethod 设置签名方法,目前只支持hmacsha1,hmacsha256,hmacmd5(默认)
func WithSignMethod(method string) Option {
	return func(ms *Sign) {
		switch method {
		case hmacsha1:
			ms.extParams["signmethod"] = hmacsha1
			ms.hfc = sha1.New
		case hmacsha256:
			ms.extParams["signmethod"] = hmacsha256
			ms.hfc = sha256.New
		case hmacmd5:
			fallthrough
		default:
			ms.extParams["signmethod"] = hmacmd5
			ms.hfc = md5.New
		}
	}
}

// WithSecureMode 设置支持的安全模式
func WithSecureMode(mode SecureMode) Option {
	return func(ms *Sign) {
		switch mode {
		case SecureModeTLSGuider:
			ms.enableTLS = true
			ms.extParams["securemode"] = modeTLSGuider
		case SecureModeTLSDirect:
			ms.enableTLS = true
			ms.extParams["securemode"] = modeTLSDirect
		case SecureModeITLSDNSID2:
			ms.enableTLS = true
			ms.extParams["securemode"] = modeITLSDNSID2
		case SecureModeTCPDirectPlain:
			fallthrough
		default:
			ms.enableTLS = false
			ms.extParams["securemode"] = modeTCPDirectPlain
		}
	}
}

// WithEnableDeviceModel 设置是否支持物模型
func WithEnableDeviceModel(enable bool) Option {
	return func(ms *Sign) {
		if enable {
			ms.extParams["v"] = alinkVersion
			delete(ms.extParams, "gw")
			delete(ms.extParams, "ext")
		} else {
			ms.extParams["gw"] = "0"
			ms.extParams["ext"] = "0"
			delete(ms.extParams, "v")
		}
	}
}

// WithExtRRPC 支持扩展RRPC 仅物模型下支持
func WithExtRRPC() Option {
	return func(ms *Sign) {
		if _, ok := ms.extParams["v"]; ok {
			ms.extParams["ext"] = "1"
		}
	}
}

// WithSDKVersion 设备SDK版本
func WithSDKVersion(ver string) Option {
	return func(ms *Sign) {
		ms.extParams["_v"] = ver
	}
}

// WithExtParamsKV 添加一个扩展参数的键值对,键值对将被添加到clientID的扩展参数上
func WithExtParamsKV(key, value string) Option {
	return func(ms *Sign) {
		ms.extParams[key] = value
	}
}
