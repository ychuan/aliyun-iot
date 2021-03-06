// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package aiot

import (
	"encoding/json"
	"sync/atomic"

	"github.com/thinkgos/aliyun-iot/uri"
)

// nextRequestID 获得下一个requestID,协程安全
func (sf *Client) nextRequestID() uint {
	return uint(atomic.AddUint32(&sf.requestID, 1))
}

// Request 发送请求,API内部已实现json序列化
// _uri 唯一定位服务器或(topic)
// requestID: 请求ID
// method: 方法
// params: 消息体Request的params
func (sf *Client) Request(_uri string, requestID uint, method string, params interface{}) error {
	out, err := json.Marshal(&Request{requestID, sf.version, params, method})
	if err != nil {
		return err
	}
	return sf.Publish(_uri, 1, out)
}

// SendRequest 发送请求,API内部已实现json序列化,requestID内部生成
// _uri 唯一定位服务器或(topic)
// method: 方法
// params: 消息体Request的params
func (sf *Client) SendRequest(_uri, method string, params interface{}) (*Token, error) {
	id := sf.nextRequestID()
	sf.Log.Debugf("%s @%d", method, id)
	if err := sf.Request(_uri, id, method, params); err != nil {
		return nil, err
	}
	return sf.putPending(id), nil
}

// Response 发送回复
// _uri 唯一定位服务器或(topic)
// Response: 回复
// API内部已实现json序列化
func (sf *Client) Response(_uri string, rsp Response) error {
	out, err := json.Marshal(rsp)
	if err != nil {
		return err
	}
	return sf.Publish(_uri, 1, out)
}

// SubscribeAllTopic 对某个设备类型订阅相关所有主题
func (sf *Client) SubscribeAllTopic(productKey, deviceName string, isSub bool) error {
	var err error
	var _uri string

	if sf.mode == ModeHTTP {
		return nil
	}
	// model raw
	_uri = uri.URI(uri.SysPrefix, uri.ThingModelUpRawReply, productKey, deviceName)
	if err = sf.Subscribe(_uri, ProcThingModelUpRawReply); err != nil {
		sf.Log.Warnf(err.Error())
	}
	_uri = uri.URI(uri.SysPrefix, uri.ThingModelDownRaw, productKey, deviceName)
	if err = sf.Subscribe(_uri, ProcThingModelDownRaw); err != nil {
		sf.Log.Warnf(err.Error())
	}

	// 网络探针
	if err = sf.Subscribe(uri.ExtNetworkProbe, ProcExtNetworkProbeRequest); err != nil {
		sf.Log.Warnf(err.Error())
	}
	// 只使能model raw
	if !sf.hasRawModel {
		// desired 期望属性订阅
		if sf.hasDesired {
			_uri = uri.URI(uri.SysPrefix, uri.ThingDesiredPropertyGetReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingDesiredPropertyGetReply); err != nil {
				sf.Log.Warnf(err.Error())
			}
			_uri = uri.URI(uri.SysPrefix, uri.ThingDesiredPropertyDeleteReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingDesiredPropertyDeleteReply); err != nil {
				sf.Log.Warnf(err.Error())
			}
		}

		// ntp订阅, 只有网关和独立设备支持ntp
		if sf.hasNTP && !isSub {
			_uri = uri.URI(uri.ExtNtpPrefix, uri.NtpResponse, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcExtNtpResponse); err != nil {
				sf.Log.Warnf(err.Error())
			}
		}

		// diag
		if sf.hasDiag && !isSub {
			_uri = uri.URI(uri.SysPrefix, uri.ThingDiagPostReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingDialPostReply); err != nil {
				sf.Log.Warnf(err.Error())
			}
		}

		if sf.hasExtRRPC {
			if err = sf.Subscribe(uri.ExtRRPCWildcardSome, ProcExtRRPCRequest); err != nil {
				sf.Log.Warnf(err.Error())
			}
		}

		// event 主题订阅
		_uri = uri.URI(uri.SysPrefix, uri.ThingEventPostReplyWildcardOne, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingEventPostReply); err != nil {
			sf.Log.Warnf(err.Error())
		}

		// event 主题订阅
		_uri = uri.URI(uri.SysPrefix, uri.ThingEventPropertyHistoryPostReply, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingEventPropertyHistoryPostReply); err != nil {
			sf.Log.Warnf(err.Error())
		}

		// deviceInfo 主题订阅
		_uri = uri.URI(uri.SysPrefix, uri.ThingDeviceInfoUpdateReply, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingDeviceInfoUpdateReply); err != nil {
			sf.Log.Warnf(err.Error())
		}
		_uri = uri.URI(uri.SysPrefix, uri.ThingDeviceInfoDeleteReply, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingDeviceInfoDeleteReply); err != nil {
			sf.Log.Warnf(err.Error())
		}

		// service
		_uri = uri.URI(uri.SysPrefix, uri.ThingServiceRequestWildcardSome, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingServiceRequest); err != nil {
			sf.Log.Warnf(err.Error())
		}

		// dsltemplate 订阅
		_uri = uri.URI(uri.SysPrefix, uri.ThingDslTemplateGetReply, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingDsltemplateGetReply); err != nil {
			sf.Log.Warnf(err.Error())
		}
		// dynamictsl
		_uri = uri.URI(uri.SysPrefix, uri.ThingDynamicTslGetReply, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingDynamictslGetReply); err != nil {
			sf.Log.Warnf(err.Error())
		}

		// Log
		_uri = uri.URI(uri.SysPrefix, uri.ThingConfigLogGetReply, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingConfigLogGetReply); err != nil {
			sf.Log.Warnf(err.Error())
		}
		_uri = uri.URI(uri.SysPrefix, uri.ThingLogPostReply, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingLogPostReply); err != nil {
			sf.Log.Warnf(err.Error())
		}
		_uri = uri.URI(uri.SysPrefix, uri.ThingConfigLogPush, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingConfigLogPush); err != nil {
			sf.Log.Warnf(err.Error())
		}

		// RRPC
		_uri = uri.URI(uri.SysPrefix, uri.RRPCRequestWildcardOne, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcRRPCRequest); err != nil {
			sf.Log.Warnf(err.Error())
		}

		// config 主题订阅
		_uri = uri.URI(uri.SysPrefix, uri.ThingConfigGetReply, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingConfigGetReply); err != nil {
			sf.Log.Warnf(err.Error())
		}
		_uri = uri.URI(uri.SysPrefix, uri.ThingConfigPush, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingConfigPush); err != nil {
			sf.Log.Warnf(err.Error())
		}

		// error 订阅
		_uri = uri.URI(uri.ExtErrorPrefix, "", productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcExtErrorResponse); err != nil {
			sf.Log.Warnf(err.Error())
		}
	}

	if sf.isGateway {
		if isSub {
			// 子设备禁用,启用,删除
			_uri = uri.URI(uri.SysPrefix, uri.ThingDisable, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingDisable); err != nil {
				sf.Log.Warnf(err.Error())
			}
			_uri = uri.URI(uri.SysPrefix, uri.ThingEnable, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingEnable); err != nil {
				sf.Log.Warnf(err.Error())
			}
			_uri = uri.URI(uri.SysPrefix, uri.ThingDelete, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingDelete); err != nil {
				sf.Log.Warnf(err.Error())
			}
		} else {
			// 子设备动态注册,topic需要用网关的productKey,deviceName
			_uri = uri.URI(uri.SysPrefix, uri.ThingSubRegisterReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingSubRegisterReply); err != nil {
				sf.Log.Warnf(err.Error())
			}
			// 子设备上线,下线,topic需要用网关的productKey,deviceName,
			// 使用的是网关的通道,所以子设备不注册相关主题
			_uri = uri.URI(uri.ExtSessionPrefix, uri.CombineLoginReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcExtCombineLoginReply); err != nil {
				sf.Log.Warnf(err.Error())
			}
			_uri = uri.URI(uri.ExtSessionPrefix, uri.CombineLogoutReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcExtCombineLogoutReply); err != nil {
				sf.Log.Warnf(err.Error())
			}
			_uri = uri.URI(uri.ExtSessionPrefix, uri.CombineBatchLoginReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcExtCombineBatchLoginReply); err != nil {
				sf.Log.Warnf(err.Error())
			}
			_uri = uri.URI(uri.ExtSessionPrefix, uri.CombineBatchLogoutReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcExtCombineBatchLogoutReply); err != nil {
				sf.Log.Warnf(err.Error())
			}

			// 网关批量上报数据,topic需要用网关的productKey,deviceName
			_uri = uri.URI(uri.SysPrefix, uri.ThingEventPropertyPackPostReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingEventPropertyPackPostReply); err != nil {
				sf.Log.Warnf(err.Error())
			}

			// 添加该网关和子设备的拓扑关系,topic需要用网关的productKey,deviceName
			_uri = uri.URI(uri.SysPrefix, uri.ThingTopoAddReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingTopoAddReply); err != nil {
				sf.Log.Warnf(err.Error())
			}

			// 删除该网关和子设备的拓扑关系,topic需要用网关的productKey,deviceName
			_uri = uri.URI(uri.SysPrefix, uri.ThingTopoDeleteReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingTopoDeleteReply); err != nil {
				sf.Log.Warnf(err.Error())
			}

			// 获取该网关和子设备的拓扑关系,topic需要用网关的productKey,deviceName
			_uri = uri.URI(uri.SysPrefix, uri.ThingTopoGetReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingTopoGetReply); err != nil {
				sf.Log.Warnf(err.Error())
			}

			// 发现设备列表上报,topic需要用网关的productKey,deviceName
			_uri = uri.URI(uri.SysPrefix, uri.ThingListFoundReply, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingListFoundReply); err != nil {
				sf.Log.Warnf(err.Error())
			}

			// 添加设备拓扑关系通知,topic需要用网关的productKey,deviceName
			_uri = uri.URI(uri.SysPrefix, uri.ThingTopoAddNotify, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingTopoAddNotify); err != nil {
				sf.Log.Warnf(err.Error())
			}

			// 网关网络拓扑关系变化通知,topic需要用网关的productKey,deviceName
			_uri = uri.URI(uri.SysPrefix, uri.ThingTopoChange, productKey, deviceName)
			if err = sf.Subscribe(_uri, ProcThingTopoChange); err != nil {
				sf.Log.Warnf(err.Error())
			}
		}
	}

	// OTA
	if sf.hasOTA {
		// OTA升级通知
		_uri = uri.URI(uri.OtaDeviceUpgradePrefix, "", productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcOtaUpgrade); err != nil {
			sf.Log.Warnf(err.Error())
		}

		// OTA 固件版本查询应答
		_uri = uri.URI(uri.SysPrefix, uri.ThingOtaFirmwareGetReply, productKey, deviceName)
		if err = sf.Subscribe(_uri, ProcThingOtaFirmwareGetReply); err != nil {
			sf.Log.Warnf(err.Error())
		}
	}

	return nil
}

// UnSubscribeAllTopic 取消订阅设备相关所有主题
func (sf *Client) UnSubscribeAllTopic(productKey, deviceName string, isSub bool) error {
	var topicList []string

	if sf.mode == ModeHTTP {
		return nil
	}

	if !sf.hasRawModel {
		// desired 期望属性取消订阅
		if sf.hasDesired {
			topicList = append(topicList,
				uri.URI(uri.SysPrefix, uri.ThingDesiredPropertyGetReply, productKey, deviceName),
				uri.URI(uri.SysPrefix, uri.ThingDesiredPropertyDeleteReply, productKey, deviceName),
			)
		}
		if sf.hasNTP && !isSub {
			topicList = append(topicList, uri.URI(uri.ExtNtpPrefix, uri.NtpResponse, productKey, deviceName))
		}
		if sf.hasDiag && !isSub {
			topicList = append(topicList, uri.URI(uri.SysPrefix, uri.ThingDiagPostReply, productKey, deviceName))
		}

		if sf.hasExtRRPC {
			topicList = append(topicList, uri.ExtRRPCWildcardSome)
		}
		topicList = append(topicList,
			// event 取消订阅
			uri.URI(uri.SysPrefix, uri.ThingEventPostReplyWildcardOne, productKey, deviceName),
			uri.URI(uri.SysPrefix, uri.ThingEventPropertyHistoryPostReply, productKey, deviceName),
			// deviceInfo
			uri.URI(uri.SysPrefix, uri.ThingDeviceInfoUpdateReply, productKey, deviceName),
			uri.URI(uri.SysPrefix, uri.ThingDeviceInfoDeleteReply, productKey, deviceName),
			// service
			uri.URI(uri.SysPrefix, uri.ThingServiceRequestWildcardSome, productKey, deviceName),
			// dystemplate
			uri.URI(uri.SysPrefix, uri.ThingDslTemplateGetReply, productKey, deviceName),
			// dynamictsl
			uri.URI(uri.SysPrefix, uri.ThingDynamicTslGetReply, productKey, deviceName),
			// log
			uri.URI(uri.SysPrefix, uri.ThingConfigLogGetReply, productKey, deviceName),
			uri.URI(uri.SysPrefix, uri.ThingLogPostReply, productKey, deviceName),
			uri.URI(uri.SysPrefix, uri.ThingConfigLogPush, productKey, deviceName),
			// RRPC
			uri.URI(uri.SysPrefix, uri.RRPCRequestWildcardOne, productKey, deviceName),
			// config
			uri.URI(uri.SysPrefix, uri.ThingConfigGetReply, productKey, deviceName),
			uri.URI(uri.SysPrefix, uri.ThingConfigPush, productKey, deviceName),
			// error
			uri.URI(uri.ExtErrorPrefix, "", productKey, deviceName),
		)
	}

	if sf.isGateway {
		if isSub {
			topicList = append(topicList,
				// 子设备禁用,启用,删除
				uri.URI(uri.SysPrefix, uri.ThingDisable, productKey, deviceName),
				uri.URI(uri.SysPrefix, uri.ThingEnable, productKey, deviceName),
				uri.URI(uri.SysPrefix, uri.ThingDelete, productKey, deviceName),
			)
		} else {
			topicList = append(topicList,
				// 子设备动态注册,topic需要用网关的productKey,deviceName
				uri.URI(uri.SysPrefix, uri.ThingSubRegisterReply, productKey, deviceName),
				// 子设备上线,下线,topic需要用网关的productKey,deviceName,
				// 使用的是网关的通道,所以子设备不注册相关主题
				uri.URI(uri.ExtSessionPrefix, uri.CombineLoginReply, productKey, deviceName),
				uri.URI(uri.ExtSessionPrefix, uri.CombineLogoutReply, productKey, deviceName),
				uri.URI(uri.ExtSessionPrefix, uri.CombineBatchLoginReply, productKey, deviceName),
				uri.URI(uri.ExtSessionPrefix, uri.CombineBatchLogoutReply, productKey, deviceName),
				// 网关批量上报数据,topic需要用网关的productKey,deviceName
				uri.URI(uri.SysPrefix, uri.ThingEventPropertyPackPostReply, productKey, deviceName),
				// 添加该网关和子设备的拓扑关系,topic需要用网关的productKey,deviceName
				uri.URI(uri.SysPrefix, uri.ThingTopoAddReply, productKey, deviceName),
				// 删除该网关和子设备的拓扑关系,topic需要用网关的productKey,deviceName
				uri.URI(uri.SysPrefix, uri.ThingTopoDeleteReply, productKey, deviceName),
				// 获取该网关和子设备的拓扑关系,topic需要用网关的productKey,deviceName
				uri.URI(uri.SysPrefix, uri.ThingTopoGetReply, productKey, deviceName),
				// 发现设备列表上报,topic需要用网关的productKey,deviceName
				uri.URI(uri.SysPrefix, uri.ThingListFoundReply, productKey, deviceName),
				// 添加设备拓扑关系通知,topic需要用网关的productKey,deviceName
				uri.URI(uri.SysPrefix, uri.ThingTopoAddNotify, productKey, deviceName),
				// 网关网络拓扑关系变化通知,topic需要用网关的productKey,deviceName
				uri.URI(uri.SysPrefix, uri.ThingTopoChange, productKey, deviceName),
			)
		}
	}

	// OTA
	if sf.hasOTA {
		topicList = append(topicList,
			// OTA升级通知
			uri.URI(uri.OtaDeviceUpgradePrefix, "", productKey, deviceName),
			// OTA 固件版本查询应答
			uri.URI(uri.SysPrefix, uri.ThingOtaFirmwareGetReply, productKey, deviceName),
		)
	}

	topicList = append(topicList,
		// model raw 取消订阅
		uri.URI(uri.SysPrefix, uri.ThingModelUpRawReply, productKey, deviceName),
		uri.URI(uri.SysPrefix, uri.ThingModelDownRaw, productKey, deviceName),
		// 网络探针
		uri.ExtNetworkProbe,
	)
	return sf.UnSubscribe(topicList...)
}
