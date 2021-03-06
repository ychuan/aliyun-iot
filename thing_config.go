// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package aiot

import (
	"encoding/json"

	"github.com/thinkgos/aliyun-iot/infra"
	"github.com/thinkgos/aliyun-iot/uri"
)

// @see https://help.aliyun.com/document_detail/89308.html?spm=a2c4g.11186623.6.713.63661dcfgclzYs

// ConfigGetParams 获取配置的参数域
type ConfigGetParams struct {
	ConfigScope string `json:"configScope"` // 配置范围, 目前只支持产品维度配置. 取值: product
	GetType     string `json:"getType"`     // 获取配置类型. 目前支持文件类型,取值: file
}

// ConfigParamsData 配置获取回复数据域或配置推送参数域
type ConfigParamsData struct {
	ConfigID   string `json:"configId"`   // 配置文件的ID
	ConfigSize int64  `json:"configSize"` // 配置文件大小,按字节计算
	Sign       string `json:"sign"`       // 签名
	SignMethod string `json:"signMethod"` // 签名方法,仅支持Sha256
	URL        string `json:"url"`        // 存储配置文件的对象存储（OSS）地址
	GetType    string `json:"getType"`    // 同ConfigGetParams.GetType
}

// ConfigGetResponse 配置获取的回复
type ConfigGetResponse struct {
	ID      uint             `json:"id,string"`
	Code    int              `json:"code"`
	Data    ConfigParamsData `json:"data"`
	Message string           `json:"message,omitempty"`
}

// ConfigPushRequest 配置推送的请求
type ConfigPushRequest struct {
	ID      uint             `json:"id,string"`
	Version string           `json:"version"`
	Params  ConfigParamsData `json:"params"`
	Method  string           `json:"method"`
}

// ThingConfigGet 获取配置参数
// request:  /sys/{productKey}/{deviceName}/thing/config/get
// response: /sys/{productKey}/{deviceName}/thing/config/get_reply
func (sf *Client) ThingConfigGet(pk, dn string) (*Token, error) {
	if !sf.IsActive(pk, dn) {
		return nil, ErrNotActive
	}
	_uri := uri.URI(uri.SysPrefix, uri.ThingConfigGet, pk, dn)
	return sf.SendRequest(_uri, infra.MethodConfigGet, ConfigGetParams{
		"product",
		"file",
	})
}

// ProcThingConfigGetReply 处理获取配置的应答
// 上行
// request:   /sys/{productKey}/{deviceName}/thing/config/get
// response:  /sys/{productKey}/{deviceName}/thing/config/get_reply
// subscribe: /sys/{productKey}/{deviceName}/thing/config/get_reply
func ProcThingConfigGetReply(c *Client, rawURI string, payload []byte) error {
	uris := uri.Spilt(rawURI)
	if len(uris) < 6 {
		return ErrInvalidURI
	}

	rsp := &ConfigGetResponse{}
	err := json.Unmarshal(payload, rsp)
	if err != nil {
		return err
	}

	if rsp.Code != infra.CodeSuccess {
		err = infra.NewCodeError(rsp.Code, rsp.Message)
	}
	c.signalPending(Message{rsp.ID, rsp.Data, err})
	pk, dn := uris[1], uris[2]
	c.Log.Debugf("thing.config.get.reply @%d", rsp.ID)
	return c.cb.ThingConfigGetReply(c, err, pk, dn, rsp.Data)
}

// ProcThingConfigPush 处理配置推送,已做回复
// 下行
// request:   /sys/{productKey}/{deviceName}/thing/config/push
// response:  /sys/{productKey}/{deviceName}/thing/config/push_reply
// subscribe: /sys/{productKey}/{deviceName}/thing/config/push
func ProcThingConfigPush(c *Client, rawURI string, payload []byte) error {
	c.Log.Debugf("thing.config.push")
	uris := uri.Spilt(rawURI)
	if len(uris) < 6 {
		return ErrInvalidURI
	}
	req := &ConfigPushRequest{}
	if err := json.Unmarshal(payload, req); err != nil {
		return err
	}
	_uri := uri.ReplyWithRequestURI(rawURI)
	err := c.Response(_uri, Response{ID: req.ID, Code: infra.CodeSuccess, Data: "{}"})
	if err != nil {
		c.Log.Errorf("thing.config.push.reply %+v", err)
	}
	pk, dn := uris[1], uris[2]
	return c.cb.ThingConfigPush(c, pk, dn, req.Params)
}
